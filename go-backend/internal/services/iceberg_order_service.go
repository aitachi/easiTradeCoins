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

// IcebergOrder represents an iceberg order that hides the true order size
// 冰山订单: 隐藏大额订单的真实数量,分批显示和执行
type IcebergOrder struct {
	ID              string           `json:"id" gorm:"primaryKey"`
	UserID          uint             `json:"user_id" gorm:"index"`
	Symbol          string           `json:"symbol" gorm:"index"`
	Side            models.OrderSide `json:"side"`
	Type            models.OrderType `json:"type"` // limit only
	Price           decimal.Decimal  `json:"price" gorm:"type:decimal(36,18)"`

	// 总数量 (隐藏)
	TotalQuantity decimal.Decimal `json:"total_quantity" gorm:"type:decimal(36,18)"`

	// 每次显示的数量
	DisplayQuantity decimal.Decimal `json:"display_quantity" gorm:"type:decimal(36,18)"`

	// 已执行数量
	ExecutedQuantity decimal.Decimal `json:"executed_quantity" gorm:"type:decimal(36,18);default:0"`

	// 当前活跃的子订单ID
	CurrentChildOrderID *string `json:"current_child_order_id,omitempty"`

	Status     string    `json:"status" gorm:"index"` // pending, active, filled, cancelled
	CreateTime time.Time `json:"create_time" gorm:"index"`
	UpdateTime time.Time `json:"update_time"`

	// 配置
	MinDisplayQuantity decimal.Decimal `json:"min_display_quantity" gorm:"type:decimal(36,18)"` // 最小显示数量
	VariancePercent    decimal.Decimal `json:"variance_percent" gorm:"type:decimal(5,2)"`       // 随机变化百分比
}

// IcebergOrderService manages iceberg orders
type IcebergOrderService struct {
	orderService *OrderService
	mutex        sync.RWMutex
	db           *gorm.DB
}

// NewIcebergOrderService creates a new iceberg order service
func NewIcebergOrderService(orderService *OrderService, db *gorm.DB) *IcebergOrderService {
	return &IcebergOrderService{
		orderService: orderService,
		db:           db,
	}
}

// CreateIcebergOrder creates a new iceberg order
func (s *IcebergOrderService) CreateIcebergOrder(
	ctx context.Context,
	userID uint,
	symbol string,
	side models.OrderSide,
	price decimal.Decimal,
	totalQuantity decimal.Decimal,
	displayQuantity decimal.Decimal,
	variancePercent decimal.Decimal, // 随机变化百分比 (0-100)
) (*IcebergOrder, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证参数
	if err := s.validateParams(totalQuantity, displayQuantity, variancePercent); err != nil {
		return nil, err
	}

	icebergOrder := &IcebergOrder{
		ID:                 fmt.Sprintf("ICEBERG-%d-%d", userID, time.Now().UnixNano()),
		UserID:             userID,
		Symbol:             symbol,
		Side:               side,
		Type:               models.OrderTypeLimit,
		Price:              price,
		TotalQuantity:      totalQuantity,
		DisplayQuantity:    displayQuantity,
		ExecutedQuantity:   decimal.Zero,
		MinDisplayQuantity: displayQuantity.Mul(decimal.NewFromFloat(0.5)), // 最小为显示数量的50%
		VariancePercent:    variancePercent,
		Status:             "pending",
		CreateTime:         time.Now(),
		UpdateTime:         time.Now(),
	}

	// 保存冰山订单
	if err := s.db.Create(icebergOrder).Error; err != nil {
		return nil, fmt.Errorf("failed to create iceberg order: %w", err)
	}

	// 创建第一个子订单
	if err := s.createNextChildOrder(ctx, icebergOrder); err != nil {
		return nil, fmt.Errorf("failed to create first child order: %w", err)
	}

	// 启动监控
	go s.monitorIcebergOrder(ctx, icebergOrder.ID)

	return icebergOrder, nil
}

// validateParams validates iceberg order parameters
func (s *IcebergOrderService) validateParams(totalQty, displayQty, variance decimal.Decimal) error {
	if totalQty.LessThanOrEqual(decimal.Zero) {
		return errors.New("total quantity must be positive")
	}

	if displayQty.LessThanOrEqual(decimal.Zero) {
		return errors.New("display quantity must be positive")
	}

	if displayQty.GreaterThan(totalQty) {
		return errors.New("display quantity cannot exceed total quantity")
	}

	if variance.LessThan(decimal.Zero) || variance.GreaterThan(decimal.NewFromInt(100)) {
		return errors.New("variance percent must be between 0 and 100")
	}

	return nil
}

// createNextChildOrder creates the next child order with randomized quantity
func (s *IcebergOrderService) createNextChildOrder(ctx context.Context, icebergOrder *IcebergOrder) error {
	// 计算剩余数量
	remainingQty := icebergOrder.TotalQuantity.Sub(icebergOrder.ExecutedQuantity)
	if remainingQty.LessThanOrEqual(decimal.Zero) {
		return errors.New("no remaining quantity")
	}

	// 计算显示数量 (带随机变化)
	displayQty := s.calculateDisplayQuantity(icebergOrder, remainingQty)

	// 创建子订单
	childOrder := &models.Order{
		UserID:      icebergOrder.UserID,
		Symbol:      icebergOrder.Symbol,
		Side:        icebergOrder.Side,
		Type:        models.OrderTypeLimit,
		Price:       icebergOrder.Price,
		Quantity:    displayQty,
		Status:      models.OrderStatusPending,
		TimeInForce: models.TimeInForceGTC,
		CreateTime:  time.Now(),
	}

	if err := s.db.Create(childOrder).Error; err != nil {
		return err
	}

	// 更新冰山订单
	icebergOrder.CurrentChildOrderID = &childOrder.ID
	icebergOrder.Status = "active"
	icebergOrder.UpdateTime = time.Now()

	return s.db.Save(icebergOrder).Error
}

// calculateDisplayQuantity calculates the display quantity with random variance
func (s *IcebergOrderService) calculateDisplayQuantity(order *IcebergOrder, remainingQty decimal.Decimal) decimal.Decimal {
	baseQty := order.DisplayQuantity

	// 如果剩余数量小于显示数量,使用剩余数量
	if remainingQty.LessThan(baseQty) {
		return remainingQty
	}

	// 添加随机变化 (±variance%)
	if order.VariancePercent.GreaterThan(decimal.Zero) {
		// 随机值 -variance% 到 +variance%
		randomPercent := decimal.NewFromFloat(
			(float64(time.Now().UnixNano()%200) - 100) / 100.0, // -1.0 到 +1.0
		).Mul(order.VariancePercent).Div(decimal.NewFromInt(100))

		variance := baseQty.Mul(randomPercent)
		adjustedQty := baseQty.Add(variance)

		// 确保不低于最小显示数量
		if adjustedQty.LessThan(order.MinDisplayQuantity) {
			adjustedQty = order.MinDisplayQuantity
		}

		// 确保不超过剩余数量
		if adjustedQty.GreaterThan(remainingQty) {
			adjustedQty = remainingQty
		}

		return adjustedQty
	}

	return baseQty
}

// monitorIcebergOrder monitors an iceberg order and creates new child orders
func (s *IcebergOrderService) monitorIcebergOrder(ctx context.Context, icebergOrderID string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if shouldStop := s.checkAndUpdateIcebergOrder(ctx, icebergOrderID); shouldStop {
				return
			}
		}
	}
}

// checkAndUpdateIcebergOrder checks the current child order and creates new ones if needed
func (s *IcebergOrderService) checkAndUpdateIcebergOrder(ctx context.Context, icebergOrderID string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var icebergOrder IcebergOrder
	if err := s.db.First(&icebergOrder, "id = ?", icebergOrderID).Error; err != nil {
		return true
	}

	// 如果已完成或取消,停止监控
	if icebergOrder.Status == "filled" || icebergOrder.Status == "cancelled" {
		return true
	}

	// 检查当前子订单
	if icebergOrder.CurrentChildOrderID != nil {
		var childOrder models.Order
		if err := s.db.First(&childOrder, "id = ?", *icebergOrder.CurrentChildOrderID).Error; err == nil {
			// 如果子订单完全成交
			if childOrder.Status == models.OrderStatusFilled {
				// 更新已执行数量
				icebergOrder.ExecutedQuantity = icebergOrder.ExecutedQuantity.Add(childOrder.FilledQty)
				icebergOrder.UpdateTime = time.Now()

				// 检查是否全部完成
				if icebergOrder.ExecutedQuantity.GreaterThanOrEqual(icebergOrder.TotalQuantity) {
					icebergOrder.Status = "filled"
					icebergOrder.CurrentChildOrderID = nil
					s.db.Save(&icebergOrder)
					return true
				}

				// 创建下一个子订单
				if err := s.createNextChildOrder(ctx, &icebergOrder); err != nil {
					fmt.Printf("Error creating next child order: %v\n", err)
					return true
				}
			} else if childOrder.Status == models.OrderStatusCancelled {
				// 如果子订单被取消,取消整个冰山订单
				icebergOrder.Status = "cancelled"
				icebergOrder.CurrentChildOrderID = nil
				s.db.Save(&icebergOrder)
				return true
			} else if childOrder.Status == models.OrderStatusPartial {
				// 部分成交,更新已执行数量
				icebergOrder.ExecutedQuantity = icebergOrder.ExecutedQuantity.Add(childOrder.FilledQty)
				icebergOrder.UpdateTime = time.Now()
				s.db.Save(&icebergOrder)
			}
		}
	}

	return false
}

// CancelIcebergOrder cancels an iceberg order
func (s *IcebergOrderService) CancelIcebergOrder(ctx context.Context, icebergOrderID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var icebergOrder IcebergOrder
	if err := s.db.First(&icebergOrder, "id = ?", icebergOrderID).Error; err != nil {
		return fmt.Errorf("iceberg order not found: %w", err)
	}

	if icebergOrder.Status == "filled" || icebergOrder.Status == "cancelled" {
		return fmt.Errorf("cannot cancel iceberg order with status: %s", icebergOrder.Status)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 取消当前子订单
		if icebergOrder.CurrentChildOrderID != nil {
			if err := tx.Model(&models.Order{}).
				Where("id = ?", *icebergOrder.CurrentChildOrderID).
				Update("status", models.OrderStatusCancelled).Error; err != nil {
				return err
			}
		}

		// 更新冰山订单状态
		icebergOrder.Status = "cancelled"
		icebergOrder.CurrentChildOrderID = nil
		icebergOrder.UpdateTime = time.Now()
		return tx.Save(&icebergOrder).Error
	})
}

// GetIcebergOrder retrieves an iceberg order by ID
func (s *IcebergOrderService) GetIcebergOrder(ctx context.Context, icebergOrderID string) (*IcebergOrder, error) {
	var order IcebergOrder
	if err := s.db.First(&order, "id = ?", icebergOrderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetUserIcebergOrders retrieves all iceberg orders for a user
func (s *IcebergOrderService) GetUserIcebergOrders(ctx context.Context, userID uint, limit, offset int) ([]IcebergOrder, int64, error) {
	var orders []IcebergOrder
	var total int64

	if err := s.db.Model(&IcebergOrder{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
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
