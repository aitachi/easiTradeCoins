//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

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

// GridStrategy represents a grid trading strategy
// 网格交易策略: 在价格区间内设置多个买卖网格,自动低买高卖
type GridStrategy struct {
	ID     string `json:"id" gorm:"primaryKey"`
	UserID uint   `json:"user_id" gorm:"index"`
	Symbol string `json:"symbol" gorm:"index"`

	// 价格区间
	LowerPrice decimal.Decimal `json:"lower_price" gorm:"type:decimal(36,18)"` // 下限价格
	UpperPrice decimal.Decimal `json:"upper_price" gorm:"type:decimal(36,18)"` // 上限价格

	// 网格参数
	GridNum int `json:"grid_num"` // 网格数量

	// 投资金额
	TotalInvestment decimal.Decimal `json:"total_investment" gorm:"type:decimal(36,18)"` // 总投资金额

	// 每格交易数量
	QuantityPerGrid decimal.Decimal `json:"quantity_per_grid" gorm:"type:decimal(36,18)"`

	// 统计数据
	TotalProfit       decimal.Decimal `json:"total_profit" gorm:"type:decimal(36,18);default:0"`        // 总利润
	CompletedGrids    int             `json:"completed_grids" gorm:"default:0"`                         // 完成的网格数
	ActiveBuyOrders   int             `json:"active_buy_orders" gorm:"default:0"`                       // 活跃买单数
	ActiveSellOrders  int             `json:"active_sell_orders" gorm:"default:0"`                      // 活跃卖单数

	// 策略配置
	AutoRestart bool            `json:"auto_restart" gorm:"default:true"` // 自动重启网格
	StopLoss    *decimal.Decimal `json:"stop_loss,omitempty" gorm:"type:decimal(36,18)"` // 止损价
	TakeProfit  *decimal.Decimal `json:"take_profit,omitempty" gorm:"type:decimal(36,18)"` // 止盈价

	Status     string    `json:"status" gorm:"index"` // pending, active, paused, stopped, completed
	CreateTime time.Time `json:"create_time" gorm:"index"`
	UpdateTime time.Time `json:"update_time"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	StopTime   *time.Time `json:"stop_time,omitempty"`
}

// GridLevel represents a single grid level
type GridLevel struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	StrategyID string          `json:"strategy_id" gorm:"index"`
	Level      int             `json:"level"` // 网格层级 (0 到 GridNum-1)
	Price      decimal.Decimal `json:"price" gorm:"type:decimal(36,18)"` // 该层级的价格

	// 订单
	BuyOrderID  *string `json:"buy_order_id,omitempty" gorm:"index"`  // 买单ID
	SellOrderID *string `json:"sell_order_id,omitempty" gorm:"index"` // 卖单ID

	// 状态
	BuyFilled  bool `json:"buy_filled" gorm:"default:false"`  // 买单是否成交
	SellFilled bool `json:"sell_filled" gorm:"default:false"` // 卖单是否成交

	// 利润
	Profit decimal.Decimal `json:"profit" gorm:"type:decimal(36,18);default:0"`

	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// GridTradingService manages grid trading strategies
type GridTradingService struct {
	orderService *OrderService
	mutex        sync.RWMutex
	db           *gorm.DB
}

// NewGridTradingService creates a new grid trading service
func NewGridTradingService(orderService *OrderService, db *gorm.DB) *GridTradingService {
	return &GridTradingService{
		orderService: orderService,
		db:           db,
	}
}

// CreateGridStrategy creates a new grid trading strategy
func (s *GridTradingService) CreateGridStrategy(
	ctx context.Context,
	userID uint,
	symbol string,
	lowerPrice, upperPrice decimal.Decimal,
	gridNum int,
	totalInvestment decimal.Decimal,
	autoRestart bool,
	stopLoss, takeProfit *decimal.Decimal,
) (*GridStrategy, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证参数
	if err := s.validateParams(lowerPrice, upperPrice, gridNum, totalInvestment); err != nil {
		return nil, err
	}

	// 计算每格价差
	priceStep := upperPrice.Sub(lowerPrice).Div(decimal.NewFromInt(int64(gridNum)))

	// 计算每格交易数量
	quantityPerGrid := totalInvestment.Div(decimal.NewFromInt(int64(gridNum)))

	strategy := &GridStrategy{
		ID:              fmt.Sprintf("GRID-%d-%d", userID, time.Now().UnixNano()),
		UserID:          userID,
		Symbol:          symbol,
		LowerPrice:      lowerPrice,
		UpperPrice:      upperPrice,
		GridNum:         gridNum,
		TotalInvestment: totalInvestment,
		QuantityPerGrid: quantityPerGrid,
		TotalProfit:     decimal.Zero,
		AutoRestart:     autoRestart,
		StopLoss:        stopLoss,
		TakeProfit:      takeProfit,
		Status:          "pending",
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
	}

	// 保存策略
	if err := s.db.Create(strategy).Error; err != nil {
		return nil, fmt.Errorf("failed to create grid strategy: %w", err)
	}

	// 创建网格层级
	if err := s.createGridLevels(ctx, strategy, priceStep); err != nil {
		return nil, fmt.Errorf("failed to create grid levels: %w", err)
	}

	// 启动网格
	go s.runGridStrategy(ctx, strategy.ID)

	return strategy, nil
}

// validateParams validates grid strategy parameters
func (s *GridTradingService) validateParams(
	lowerPrice, upperPrice decimal.Decimal,
	gridNum int,
	totalInvestment decimal.Decimal,
) error {
	if lowerPrice.LessThanOrEqual(decimal.Zero) {
		return errors.New("lower price must be positive")
	}

	if upperPrice.LessThanOrEqual(lowerPrice) {
		return errors.New("upper price must be greater than lower price")
	}

	if gridNum < 2 || gridNum > 200 {
		return errors.New("grid number must be between 2 and 200")
	}

	if totalInvestment.LessThanOrEqual(decimal.Zero) {
		return errors.New("total investment must be positive")
	}

	return nil
}

// createGridLevels creates all grid levels for the strategy
func (s *GridTradingService) createGridLevels(ctx context.Context, strategy *GridStrategy, priceStep decimal.Decimal) error {
	levels := make([]GridLevel, strategy.GridNum)

	for i := 0; i < strategy.GridNum; i++ {
		price := strategy.LowerPrice.Add(priceStep.Mul(decimal.NewFromInt(int64(i))))

		levels[i] = GridLevel{
			StrategyID: strategy.ID,
			Level:      i,
			Price:      price,
			BuyFilled:  false,
			SellFilled: false,
			Profit:     decimal.Zero,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
	}

	return s.db.Create(&levels).Error
}

// runGridStrategy runs the grid trading strategy
func (s *GridTradingService) runGridStrategy(ctx context.Context, strategyID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// 初始化网格订单
	s.initializeGridOrders(ctx, strategyID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			shouldStop := s.updateGridStrategy(ctx, strategyID)
			if shouldStop {
				return
			}
		}
	}
}

// initializeGridOrders creates initial buy orders for all grid levels
func (s *GridTradingService) initializeGridOrders(ctx context.Context, strategyID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var strategy GridStrategy
	if err := s.db.First(&strategy, "id = ?", strategyID).Error; err != nil {
		return
	}

	// 获取当前市场价格
	var lastTrade models.Trade
	if err := s.db.Where("symbol = ?", strategy.Symbol).
		Order("trade_time DESC").
		First(&lastTrade).Error; err != nil {
		return
	}

	currentPrice := lastTrade.Price

	// 获取所有网格层级
	var levels []GridLevel
	if err := s.db.Where("strategy_id = ?", strategyID).
		Order("level ASC").
		Find(&levels).Error; err != nil {
		return
	}

	// 为价格低于当前价的网格创建买单
	// 为价格高于当前价的网格创建卖单
	for i := range levels {
		level := &levels[i]

		if level.Price.LessThan(currentPrice) && level.BuyOrderID == nil {
			// 创建买单
			s.createBuyOrder(ctx, &strategy, level)
		} else if level.Price.GreaterThan(currentPrice) && level.SellOrderID == nil {
			// 创建卖单 (需要先有持仓)
			// 这里简化处理,实际应该检查用户持仓
		}
	}

	strategy.Status = "active"
	now := time.Now()
	strategy.StartTime = &now
	strategy.UpdateTime = now
	s.db.Save(&strategy)
}

// createBuyOrder creates a buy order for a grid level
func (s *GridTradingService) createBuyOrder(ctx context.Context, strategy *GridStrategy, level *GridLevel) error {
	order := &models.Order{
		UserID:      strategy.UserID,
		Symbol:      strategy.Symbol,
		Side:        models.OrderSideBuy,
		Type:        models.OrderTypeLimit,
		Price:       level.Price,
		Quantity:    strategy.QuantityPerGrid.Div(level.Price), // 计算购买数量
		Status:      models.OrderStatusPending,
		TimeInForce: models.TimeInForceGTC,
		CreateTime:  time.Now(),
	}

	createdOrder, _, err := s.orderService.CreateOrder(order)
	if err != nil {
		return err
	}

	level.BuyOrderID = &createdOrder.ID
	level.UpdateTime = time.Now()
	s.db.Save(level)

	strategy.ActiveBuyOrders++
	s.db.Save(strategy)

	return nil
}

// createSellOrder creates a sell order for a grid level
func (s *GridTradingService) createSellOrder(ctx context.Context, strategy *GridStrategy, level *GridLevel, quantity decimal.Decimal) error {
	// 在上一个网格层级的价格卖出
	var upperLevel GridLevel
	if err := s.db.Where("strategy_id = ? AND level = ?", strategy.ID, level.Level+1).
		First(&upperLevel).Error; err != nil {
		return err
	}

	order := &models.Order{
		UserID:      strategy.UserID,
		Symbol:      strategy.Symbol,
		Side:        models.OrderSideSell,
		Type:        models.OrderTypeLimit,
		Price:       upperLevel.Price,
		Quantity:    quantity,
		Status:      models.OrderStatusPending,
		TimeInForce: models.TimeInForceGTC,
		CreateTime:  time.Now(),
	}

	createdOrder, _, err := s.orderService.CreateOrder(order)
	if err != nil {
		return err
	}

	level.SellOrderID = &createdOrder.ID
	level.UpdateTime = time.Now()
	s.db.Save(level)

	strategy.ActiveSellOrders++
	s.db.Save(strategy)

	return nil
}

// updateGridStrategy updates grid strategy and checks for filled orders
func (s *GridTradingService) updateGridStrategy(ctx context.Context, strategyID string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var strategy GridStrategy
	if err := s.db.First(&strategy, "id = ?", strategyID).Error; err != nil {
		return true
	}

	if strategy.Status != "active" {
		return true
	}

	// 检查止损止盈
	if s.checkStopConditions(ctx, &strategy) {
		return true
	}

	// 获取所有网格层级
	var levels []GridLevel
	if err := s.db.Where("strategy_id = ?", strategyID).
		Order("level ASC").
		Find(&levels).Error; err != nil {
		return false
	}

	// 检查每个层级的订单状态
	for i := range levels {
		level := &levels[i]
		s.checkGridLevel(ctx, &strategy, level)
	}

	strategy.UpdateTime = time.Now()
	s.db.Save(&strategy)

	return false
}

// checkGridLevel checks the status of orders in a grid level
func (s *GridTradingService) checkGridLevel(ctx context.Context, strategy *GridStrategy, level *GridLevel) {
	// 检查买单
	if level.BuyOrderID != nil && !level.BuyFilled {
		var buyOrder models.Order
		if err := s.db.First(&buyOrder, "id = ?", *level.BuyOrderID).Error; err == nil {
			if buyOrder.Status == models.OrderStatusFilled {
				level.BuyFilled = true
				strategy.ActiveBuyOrders--

				// 创建卖单
				s.createSellOrder(ctx, strategy, level, buyOrder.FilledQty)

				level.UpdateTime = time.Now()
				s.db.Save(level)
			}
		}
	}

	// 检查卖单
	if level.SellOrderID != nil && !level.SellFilled {
		var sellOrder models.Order
		if err := s.db.First(&sellOrder, "id = ?", *level.SellOrderID).Error; err == nil {
			if sellOrder.Status == models.OrderStatusFilled {
				level.SellFilled = true
				strategy.ActiveSellOrders--
				strategy.CompletedGrids++

				// 计算利润
				var buyOrder models.Order
				if level.BuyOrderID != nil {
					s.db.First(&buyOrder, "id = ?", *level.BuyOrderID)
					profit := sellOrder.FilledAmount.Sub(buyOrder.FilledAmount)
					level.Profit = profit
					strategy.TotalProfit = strategy.TotalProfit.Add(profit)
				}

				// 如果自动重启,重新创建买单
				if strategy.AutoRestart {
					level.BuyFilled = false
					level.SellFilled = false
					level.BuyOrderID = nil
					level.SellOrderID = nil
					s.createBuyOrder(ctx, strategy, level)
				}

				level.UpdateTime = time.Now()
				s.db.Save(level)
			}
		}
	}
}

// checkStopConditions checks stop-loss and take-profit conditions
func (s *GridTradingService) checkStopConditions(ctx context.Context, strategy *GridStrategy) bool {
	var lastTrade models.Trade
	if err := s.db.Where("symbol = ?", strategy.Symbol).
		Order("trade_time DESC").
		First(&lastTrade).Error; err != nil {
		return false
	}

	currentPrice := lastTrade.Price

	// 检查止损
	if strategy.StopLoss != nil && currentPrice.LessThanOrEqual(*strategy.StopLoss) {
		s.stopStrategy(ctx, strategy, "stop_loss_triggered")
		return true
	}

	// 检查止盈
	if strategy.TakeProfit != nil && currentPrice.GreaterThanOrEqual(*strategy.TakeProfit) {
		s.stopStrategy(ctx, strategy, "take_profit_triggered")
		return true
	}

	return false
}

// StopGridStrategy stops a grid trading strategy
func (s *GridTradingService) StopGridStrategy(ctx context.Context, strategyID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var strategy GridStrategy
	if err := s.db.First(&strategy, "id = ?", strategyID).Error; err != nil {
		return fmt.Errorf("strategy not found: %w", err)
	}

	return s.stopStrategy(ctx, &strategy, "manual_stop")
}

// stopStrategy stops a grid strategy and cancels all active orders
func (s *GridTradingService) stopStrategy(ctx context.Context, strategy *GridStrategy, reason string) error {
	// 取消所有活跃订单
	var levels []GridLevel
	s.db.Where("strategy_id = ?", strategy.ID).Find(&levels)

	for _, level := range levels {
		if level.BuyOrderID != nil && !level.BuyFilled {
			s.db.Model(&models.Order{}).
				Where("id = ?", *level.BuyOrderID).
				Update("status", models.OrderStatusCancelled)
		}
		if level.SellOrderID != nil && !level.SellFilled {
			s.db.Model(&models.Order{}).
				Where("id = ?", *level.SellOrderID).
				Update("status", models.OrderStatusCancelled)
		}
	}

	strategy.Status = "stopped"
	now := time.Now()
	strategy.StopTime = &now
	strategy.UpdateTime = now
	strategy.ActiveBuyOrders = 0
	strategy.ActiveSellOrders = 0

	return s.db.Save(strategy).Error
}

// GetGridStrategy retrieves a grid strategy by ID
func (s *GridTradingService) GetGridStrategy(ctx context.Context, strategyID string) (*GridStrategy, error) {
	var strategy GridStrategy
	if err := s.db.First(&strategy, "id = ?", strategyID).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

// GetGridLevels retrieves all levels for a strategy
func (s *GridTradingService) GetGridLevels(ctx context.Context, strategyID string) ([]GridLevel, error) {
	var levels []GridLevel
	if err := s.db.Where("strategy_id = ?", strategyID).
		Order("level ASC").
		Find(&levels).Error; err != nil {
		return nil, err
	}
	return levels, nil
}

// GetUserGridStrategies retrieves all grid strategies for a user
func (s *GridTradingService) GetUserGridStrategies(ctx context.Context, userID uint, limit, offset int) ([]GridStrategy, int64, error) {
	var strategies []GridStrategy
	var total int64

	if err := s.db.Model(&GridStrategy{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Where("user_id = ?", userID).
		Order("create_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&strategies).Error; err != nil {
		return nil, 0, err
	}

	return strategies, total, nil
}
