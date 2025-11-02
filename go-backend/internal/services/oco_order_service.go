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

// OCOOrder represents a One-Cancels-Other order pair
// 一个订单成交后,自动取消另一个订单
type OCOOrder struct {
	ID              string           `json:"id" gorm:"primaryKey"`
	UserID          uint             `json:"user_id" gorm:"index"`
	Symbol          string           `json:"symbol" gorm:"index"`
	Side            models.OrderSide `json:"side"` // buy or sell
	Quantity        decimal.Decimal  `json:"quantity" gorm:"type:decimal(36,18)"`

	// 止损订单ID (Stop-Loss)
	StopLossOrderID string `json:"stop_loss_order_id" gorm:"index"`
	StopLossPrice   decimal.Decimal `json:"stop_loss_price" gorm:"type:decimal(36,18)"`

	// 止盈订单ID (Take-Profit)
	TakeProfitOrderID string `json:"take_profit_order_id" gorm:"index"`
	TakeProfitPrice   decimal.Decimal `json:"take_profit_price" gorm:"type:decimal(36,18)"`

	Status     string    `json:"status" gorm:"index"` // pending, partially_filled, filled, cancelled
	CreateTime time.Time `json:"create_time" gorm:"index"`
	UpdateTime time.Time `json:"update_time"`

	// 哪个订单被触发了
	TriggeredOrderID *string `json:"triggered_order_id,omitempty"`
	TriggerTime      *time.Time `json:"trigger_time,omitempty"`
}

// OCOOrderService manages OCO orders
type OCOOrderService struct {
	orderService *OrderService
	mutex        sync.RWMutex
	db           *gorm.DB
}

// NewOCOOrderService creates a new OCO order service
func NewOCOOrderService(orderService *OrderService, db *gorm.DB) *OCOOrderService {
	return &OCOOrderService{
		orderService: orderService,
		db:           db,
	}
}

// CreateOCOOrder creates a new OCO order with stop-loss and take-profit
func (s *OCOOrderService) CreateOCOOrder(
	ctx context.Context,
	userID uint,
	symbol string,
	side models.OrderSide,
	quantity decimal.Decimal,
	stopLossPrice decimal.Decimal,
	takeProfitPrice decimal.Decimal,
) (*OCOOrder, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证价格关系
	if err := s.validatePrices(side, stopLossPrice, takeProfitPrice); err != nil {
		return nil, err
	}

	ocoOrder := &OCOOrder{
		ID:                fmt.Sprintf("OCO-%d-%d", userID, time.Now().UnixNano()),
		UserID:            userID,
		Symbol:            symbol,
		Side:              side,
		Quantity:          quantity,
		StopLossPrice:     stopLossPrice,
		TakeProfitPrice:   takeProfitPrice,
		Status:            "pending",
		CreateTime:        time.Now(),
		UpdateTime:        time.Now(),
	}

	// 在数据库事务中创建OCO订单和两个子订单
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建止损订单
		stopLossOrder := &models.Order{
			UserID:           userID,
			Symbol:           symbol,
			Side:             side,
			Type:             models.OrderTypeStopLoss,
			Quantity:         quantity,
			StopPrice:        &stopLossPrice,
			TriggerCondition: s.getTriggerCondition(side, "stop_loss"),
			Status:           models.OrderStatusPending,
			TimeInForce:      models.TimeInForceGTC,
			IsTriggered:      false,
			CreateTime:       time.Now(),
		}

		if err := tx.Create(stopLossOrder).Error; err != nil {
			return fmt.Errorf("failed to create stop-loss order: %w", err)
		}
		ocoOrder.StopLossOrderID = stopLossOrder.ID

		// 2. 创建止盈订单
		takeProfitOrder := &models.Order{
			UserID:           userID,
			Symbol:           symbol,
			Side:             side,
			Type:             models.OrderTypeTakeProfit,
			Quantity:         quantity,
			TakeProfitPrice:  &takeProfitPrice,
			TriggerCondition: s.getTriggerCondition(side, "take_profit"),
			Status:           models.OrderStatusPending,
			TimeInForce:      models.TimeInForceGTC,
			IsTriggered:      false,
			CreateTime:       time.Now(),
		}

		if err := tx.Create(takeProfitOrder).Error; err != nil {
			return fmt.Errorf("failed to create take-profit order: %w", err)
		}
		ocoOrder.TakeProfitOrderID = takeProfitOrder.ID

		// 3. 保存OCO订单记录
		if err := tx.Create(ocoOrder).Error; err != nil {
			return fmt.Errorf("failed to create OCO order: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 启动监控goroutine
	go s.monitorOCOOrder(ctx, ocoOrder.ID)

	return ocoOrder, nil
}

// validatePrices validates the price relationship for OCO orders
func (s *OCOOrderService) validatePrices(side models.OrderSide, stopLoss, takeProfit decimal.Decimal) error {
	if stopLoss.LessThanOrEqual(decimal.Zero) || takeProfit.LessThanOrEqual(decimal.Zero) {
		return errors.New("prices must be positive")
	}

	if side == models.OrderSideSell {
		// 卖单: 止损价 > 止盈价
		if stopLoss.LessThanOrEqual(takeProfit) {
			return errors.New("for sell orders, stop-loss price must be higher than take-profit price")
		}
	} else {
		// 买单: 止损价 < 止盈价
		if stopLoss.GreaterThanOrEqual(takeProfit) {
			return errors.New("for buy orders, stop-loss price must be lower than take-profit price")
		}
	}

	return nil
}

// getTriggerCondition returns the trigger condition for an order
func (s *OCOOrderService) getTriggerCondition(side models.OrderSide, orderType string) string {
	if orderType == "stop_loss" {
		if side == models.OrderSideSell {
			return "<=" // 卖单止损: 价格下跌到止损价
		}
		return ">=" // 买单止损: 价格上涨到止损价
	} else { // take_profit
		if side == models.OrderSideSell {
			return ">=" // 卖单止盈: 价格上涨到止盈价
		}
		return "<=" // 买单止盈: 价格下跌到止盈价
	}
}

// monitorOCOOrder monitors an OCO order and cancels the other order when one is triggered
func (s *OCOOrderService) monitorOCOOrder(ctx context.Context, ocoOrderID string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if shouldStop := s.checkOCOOrderStatus(ctx, ocoOrderID); shouldStop {
				return
			}
		}
	}
}

// checkOCOOrderStatus checks if one of the OCO orders has been triggered
func (s *OCOOrderService) checkOCOOrderStatus(ctx context.Context, ocoOrderID string) bool {
	var ocoOrder OCOOrder
	if err := s.db.First(&ocoOrder, "id = ?", ocoOrderID).Error; err != nil {
		return true // Stop monitoring if OCO order not found
	}

	// 如果OCO订单已经完成或取消,停止监控
	if ocoOrder.Status == "filled" || ocoOrder.Status == "cancelled" {
		return true
	}

	// 检查止损订单
	var stopLossOrder models.Order
	if err := s.db.First(&stopLossOrder, "id = ?", ocoOrder.StopLossOrderID).Error; err == nil {
		if stopLossOrder.IsTriggered || stopLossOrder.Status == models.OrderStatusFilled {
			// 止损订单被触发,取消止盈订单
			s.cancelOtherOrder(ctx, &ocoOrder, ocoOrder.StopLossOrderID, ocoOrder.TakeProfitOrderID)
			return true
		}
	}

	// 检查止盈订单
	var takeProfitOrder models.Order
	if err := s.db.First(&takeProfitOrder, "id = ?", ocoOrder.TakeProfitOrderID).Error; err == nil {
		if takeProfitOrder.IsTriggered || takeProfitOrder.Status == models.OrderStatusFilled {
			// 止盈订单被触发,取消止损订单
			s.cancelOtherOrder(ctx, &ocoOrder, ocoOrder.TakeProfitOrderID, ocoOrder.StopLossOrderID)
			return true
		}
	}

	return false
}

// cancelOtherOrder cancels the other order when one is triggered
func (s *OCOOrderService) cancelOtherOrder(ctx context.Context, ocoOrder *OCOOrder, triggeredOrderID, cancelOrderID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 更新OCO订单状态
		now := time.Now()
		ocoOrder.Status = "filled"
		ocoOrder.TriggeredOrderID = &triggeredOrderID
		ocoOrder.TriggerTime = &now
		ocoOrder.UpdateTime = now

		if err := tx.Save(ocoOrder).Error; err != nil {
			return err
		}

		// 取消另一个订单
		return tx.Model(&models.Order{}).
			Where("id = ?", cancelOrderID).
			Updates(map[string]interface{}{
				"status":      models.OrderStatusCancelled,
				"update_time": now,
			}).Error
	})

	if err != nil {
		fmt.Printf("Error canceling other order: %v\n", err)
	}
}

// CancelOCOOrder cancels both orders in an OCO order
func (s *OCOOrderService) CancelOCOOrder(ctx context.Context, ocoOrderID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var ocoOrder OCOOrder
	if err := s.db.First(&ocoOrder, "id = ?", ocoOrderID).Error; err != nil {
		return fmt.Errorf("OCO order not found: %w", err)
	}

	if ocoOrder.Status != "pending" {
		return fmt.Errorf("cannot cancel OCO order with status: %s", ocoOrder.Status)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// 取消两个订单
		if err := tx.Model(&models.Order{}).
			Where("id IN (?)", []string{ocoOrder.StopLossOrderID, ocoOrder.TakeProfitOrderID}).
			Updates(map[string]interface{}{
				"status":      models.OrderStatusCancelled,
				"update_time": now,
			}).Error; err != nil {
			return err
		}

		// 更新OCO订单状态
		ocoOrder.Status = "cancelled"
		ocoOrder.UpdateTime = now
		return tx.Save(&ocoOrder).Error
	})
}

// GetOCOOrder retrieves an OCO order by ID
func (s *OCOOrderService) GetOCOOrder(ctx context.Context, ocoOrderID string) (*OCOOrder, error) {
	var ocoOrder OCOOrder
	if err := s.db.First(&ocoOrder, "id = ?", ocoOrderID).Error; err != nil {
		return nil, err
	}
	return &ocoOrder, nil
}

// GetUserOCOOrders retrieves all OCO orders for a user
func (s *OCOOrderService) GetUserOCOOrders(ctx context.Context, userID uint, limit, offset int) ([]OCOOrder, int64, error) {
	var orders []OCOOrder
	var total int64

	if err := s.db.Model(&OCOOrder{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
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
