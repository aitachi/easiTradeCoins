package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/easitradecoins/backend/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TWAPOrder represents a Time-Weighted Average Price order
// TWAP订单: 在指定时间内,按固定时间间隔分批执行,以获得时间加权平均价格
type TWAPOrder struct {
	ID       string           `json:"id" gorm:"primaryKey"`
	UserID   uint             `json:"user_id" gorm:"index"`
	Symbol   string           `json:"symbol" gorm:"index"`
	Side     models.OrderSide `json:"side"`
	Type     models.OrderType `json:"type"` // market or limit

	// 总数量
	TotalQuantity decimal.Decimal `json:"total_quantity" gorm:"type:decimal(36,18)"`

	// 已执行数量
	ExecutedQuantity decimal.Decimal `json:"executed_quantity" gorm:"type:decimal(36,18);default:0"`

	// 已执行金额
	ExecutedAmount decimal.Decimal `json:"executed_amount" gorm:"type:decimal(36,18);default:0"`

	// 平均价格
	AveragePrice decimal.Decimal `json:"average_price" gorm:"type:decimal(36,18);default:0"`

	// 时间参数
	Duration  int64     `json:"duration" gorm:"not null"`           // 总持续时间 (秒)
	Intervals int       `json:"intervals" gorm:"not null"`          // 分割次数
	StartTime time.Time `json:"start_time" gorm:"index"`            // 开始时间
	EndTime   time.Time `json:"end_time"`                           // 结束时间
	NextSlice time.Time `json:"next_slice"`                         // 下次执行时间

	// 价格限制 (仅限价单)
	LimitPrice *decimal.Decimal `json:"limit_price,omitempty" gorm:"type:decimal(36,18)"`

	// 价格容差 (百分比) - 市价单保护
	PriceTolerance decimal.Decimal `json:"price_tolerance" gorm:"type:decimal(5,2);default:5"` // 默认5%

	Status     string    `json:"status" gorm:"index"` // pending, active, completed, cancelled, failed
	CreateTime time.Time `json:"create_time" gorm:"index"`
	UpdateTime time.Time `json:"update_time"`

	// 执行统计
	CompletedSlices int `json:"completed_slices" gorm:"default:0"` // 已完成的切片数
	FailedSlices    int `json:"failed_slices" gorm:"default:0"`    // 失败的切片数
}

// TWAPSlice represents a single execution slice of a TWAP order
type TWAPSlice struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	TWAPOrderID string          `json:"twap_order_id" gorm:"index"`
	OrderID     *string         `json:"order_id,omitempty" gorm:"index"` // 子订单ID
	SliceNumber int             `json:"slice_number"`                    // 第几个切片
	Quantity    decimal.Decimal `json:"quantity" gorm:"type:decimal(36,18)"`
	Price       decimal.Decimal `json:"price" gorm:"type:decimal(36,18)"`
	Status      string          `json:"status"` // pending, executing, completed, failed
	ScheduledAt time.Time       `json:"scheduled_at"`
	ExecutedAt  *time.Time      `json:"executed_at,omitempty"`
	Error       string          `json:"error,omitempty"`
	CreateTime  time.Time       `json:"create_time"`
}

// TWAPOrderService manages TWAP orders
type TWAPOrderService struct {
	orderService *OrderService
	mutex        sync.RWMutex
	db           *gorm.DB
}

// NewTWAPOrderService creates a new TWAP order service
func NewTWAPOrderService(orderService *OrderService, db *gorm.DB) *TWAPOrderService {
	return &TWAPOrderService{
		orderService: orderService,
		db:           db,
	}
}

// CreateTWAPOrder creates a new TWAP order
func (s *TWAPOrderService) CreateTWAPOrder(
	ctx context.Context,
	userID uint,
	symbol string,
	side models.OrderSide,
	orderType models.OrderType,
	totalQuantity decimal.Decimal,
	duration int64, // 秒
	intervals int,
	limitPrice *decimal.Decimal,
	priceTolerance decimal.Decimal,
) (*TWAPOrder, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证参数
	if err := s.validateParams(totalQuantity, duration, intervals, orderType, limitPrice); err != nil {
		return nil, err
	}

	now := time.Now()
	twapOrder := &TWAPOrder{
		ID:               fmt.Sprintf("TWAP-%d-%d", userID, now.UnixNano()),
		UserID:           userID,
		Symbol:           symbol,
		Side:             side,
		Type:             orderType,
		TotalQuantity:    totalQuantity,
		ExecutedQuantity: decimal.Zero,
		ExecutedAmount:   decimal.Zero,
		AveragePrice:     decimal.Zero,
		Duration:         duration,
		Intervals:        intervals,
		StartTime:        now,
		EndTime:          now.Add(time.Duration(duration) * time.Second),
		NextSlice:        now,
		LimitPrice:       limitPrice,
		PriceTolerance:   priceTolerance,
		Status:           "pending",
		CreateTime:       now,
		UpdateTime:       now,
		CompletedSlices:  0,
		FailedSlices:     0,
	}

	// 保存TWAP订单
	if err := s.db.Create(twapOrder).Error; err != nil {
		return nil, fmt.Errorf("failed to create TWAP order: %w", err)
	}

	// 启动执行器
	go s.executeTWAPOrder(ctx, twapOrder.ID)

	return twapOrder, nil
}

// validateParams validates TWAP order parameters
func (s *TWAPOrderService) validateParams(
	totalQty decimal.Decimal,
	duration int64,
	intervals int,
	orderType models.OrderType,
	limitPrice *decimal.Decimal,
) error {
	if totalQty.LessThanOrEqual(decimal.Zero) {
		return errors.New("total quantity must be positive")
	}

	if duration <= 0 {
		return errors.New("duration must be positive")
	}

	if intervals <= 0 {
		return errors.New("intervals must be positive")
	}

	if intervals > 1000 {
		return errors.New("intervals cannot exceed 1000")
	}

	if int64(intervals) > duration {
		return errors.New("intervals cannot exceed duration in seconds")
	}

	if orderType == models.OrderTypeLimit && limitPrice == nil {
		return errors.New("limit price is required for limit orders")
	}

	return nil
}

// executeTWAPOrder executes a TWAP order by creating slices at intervals
func (s *TWAPOrderService) executeTWAPOrder(ctx context.Context, twapOrderID string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			shouldStop := s.processNextSlice(ctx, twapOrderID)
			if shouldStop {
				return
			}
		}
	}
}

// processNextSlice processes the next slice of a TWAP order
func (s *TWAPOrderService) processNextSlice(ctx context.Context, twapOrderID string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var twapOrder TWAPOrder
	if err := s.db.First(&twapOrder, "id = ?", twapOrderID).Error; err != nil {
		return true
	}

	// 检查状态
	if twapOrder.Status == "completed" || twapOrder.Status == "cancelled" {
		return true
	}

	now := time.Now()

	// 检查是否到达下次执行时间
	if now.Before(twapOrder.NextSlice) {
		return false
	}

	// 检查是否超过结束时间
	if now.After(twapOrder.EndTime) {
		twapOrder.Status = "completed"
		twapOrder.UpdateTime = now
		s.db.Save(&twapOrder)
		return true
	}

	// 激活订单
	if twapOrder.Status == "pending" {
		twapOrder.Status = "active"
	}

	// 计算每个切片的数量
	sliceQuantity := twapOrder.TotalQuantity.Div(decimal.NewFromInt(int64(twapOrder.Intervals)))

	// 最后一个切片使用剩余数量
	if twapOrder.CompletedSlices == twapOrder.Intervals-1 {
		sliceQuantity = twapOrder.TotalQuantity.Sub(twapOrder.ExecutedQuantity)
	}

	// 创建切片订单
	sliceNumber := twapOrder.CompletedSlices + 1
	slice := &TWAPSlice{
		TWAPOrderID: twapOrder.ID,
		SliceNumber: sliceNumber,
		Quantity:    sliceQuantity,
		Status:      "pending",
		ScheduledAt: now,
		CreateTime:  now,
	}

	// 执行切片
	err := s.executeSlice(ctx, &twapOrder, slice)

	// 保存切片记录
	s.db.Create(slice)

	if err != nil {
		twapOrder.FailedSlices++
		slice.Status = "failed"
		slice.Error = err.Error()
		s.db.Save(slice)
	} else {
		twapOrder.CompletedSlices++
	}

	// 计算下次执行时间
	intervalDuration := time.Duration(twapOrder.Duration/int64(twapOrder.Intervals)) * time.Second
	twapOrder.NextSlice = now.Add(intervalDuration)

	// 检查是否完成
	if twapOrder.CompletedSlices >= twapOrder.Intervals {
		twapOrder.Status = "completed"
	} else if twapOrder.FailedSlices > twapOrder.Intervals/2 {
		// 如果失败次数超过一半,标记为失败
		twapOrder.Status = "failed"
	}

	twapOrder.UpdateTime = now
	s.db.Save(&twapOrder)

	return twapOrder.Status == "completed" || twapOrder.Status == "failed"
}

// executeSlice executes a single slice of the TWAP order
func (s *TWAPOrderService) executeSlice(ctx context.Context, twapOrder *TWAPOrder, slice *TWAPSlice) error {
	// 创建订单
	order := &models.Order{
		UserID:      twapOrder.UserID,
		Symbol:      twapOrder.Symbol,
		Side:        twapOrder.Side,
		Type:        twapOrder.Type,
		Quantity:    slice.Quantity,
		Status:      models.OrderStatusPending,
		TimeInForce: models.TimeInForceIOC, // 立即成交或取消
		CreateTime:  time.Now(),
	}

	// 设置价格
	if twapOrder.Type == models.OrderTypeLimit && twapOrder.LimitPrice != nil {
		order.Price = *twapOrder.LimitPrice
	} else {
		// 市价单 - 获取当前市场价格
		var lastTrade models.Trade
		if err := s.db.Where("symbol = ?", twapOrder.Symbol).
			Order("trade_time DESC").
			First(&lastTrade).Error; err != nil {
			return fmt.Errorf("failed to get market price: %w", err)
		}

		// 检查价格容差
		order.Price = lastTrade.Price
		slice.Price = lastTrade.Price
	}

	// 创建订单
	createdOrder, _, err := s.orderService.CreateOrder(order)
	if err != nil {
		return fmt.Errorf("failed to create slice order: %w", err)
	}

	slice.OrderID = &createdOrder.ID
	slice.Status = "completed"
	slice.Price = createdOrder.AvgPrice
	now := time.Now()
	slice.ExecutedAt = &now

	// 更新TWAP订单统计
	twapOrder.ExecutedQuantity = twapOrder.ExecutedQuantity.Add(createdOrder.FilledQty)
	twapOrder.ExecutedAmount = twapOrder.ExecutedAmount.Add(createdOrder.FilledAmount)

	if twapOrder.ExecutedQuantity.GreaterThan(decimal.Zero) {
		twapOrder.AveragePrice = twapOrder.ExecutedAmount.Div(twapOrder.ExecutedQuantity)
	}

	return nil
}

// CancelTWAPOrder cancels a TWAP order
func (s *TWAPOrderService) CancelTWAPOrder(ctx context.Context, twapOrderID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var twapOrder TWAPOrder
	if err := s.db.First(&twapOrder, "id = ?", twapOrderID).Error; err != nil {
		return fmt.Errorf("TWAP order not found: %w", err)
	}

	if twapOrder.Status == "completed" || twapOrder.Status == "cancelled" {
		return fmt.Errorf("cannot cancel TWAP order with status: %s", twapOrder.Status)
	}

	twapOrder.Status = "cancelled"
	twapOrder.UpdateTime = time.Now()

	return s.db.Save(&twapOrder).Error
}

// GetTWAPOrder retrieves a TWAP order by ID
func (s *TWAPOrderService) GetTWAPOrder(ctx context.Context, twapOrderID string) (*TWAPOrder, error) {
	var order TWAPOrder
	if err := s.db.First(&order, "id = ?", twapOrderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetTWAPOrderSlices retrieves all slices for a TWAP order
func (s *TWAPOrderService) GetTWAPOrderSlices(ctx context.Context, twapOrderID string) ([]TWAPSlice, error) {
	var slices []TWAPSlice
	if err := s.db.Where("twap_order_id = ?", twapOrderID).
		Order("slice_number ASC").
		Find(&slices).Error; err != nil {
		return nil, err
	}
	return slices, nil
}

// GetUserTWAPOrders retrieves all TWAP orders for a user
func (s *TWAPOrderService) GetUserTWAPOrders(ctx context.Context, userID uint, limit, offset int) ([]TWAPOrder, int64, error) {
	var orders []TWAPOrder
	var total int64

	if err := s.db.Model(&TWAPOrder{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Where("user_id = ?", userID).
		Order("create_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
