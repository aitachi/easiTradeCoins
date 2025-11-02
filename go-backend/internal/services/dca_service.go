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

// DCAStrategy represents a Dollar Cost Averaging strategy
// 定投策略: 定期定额投资,分散市场风险
type DCAStrategy struct {
	ID     string `json:"id" gorm:"primaryKey"`
	UserID uint   `json:"user_id" gorm:"index"`
	Symbol string `json:"symbol" gorm:"index"` // 交易对

	// 投资参数
	AmountPerPeriod decimal.Decimal `json:"amount_per_period" gorm:"type:decimal(36,18)"` // 每期投资金额
	Frequency       string          `json:"frequency"`                                     // daily, weekly, monthly
	DayOfWeek       *int            `json:"day_of_week,omitempty"`                         // 周几执行 (weekly)
	DayOfMonth      *int            `json:"day_of_month,omitempty"`                        // 每月几号执行 (monthly)
	HourOfDay       int             `json:"hour_of_day" gorm:"default:0"`                  // 每天几点执行

	// 执行条件
	MaxPrice     *decimal.Decimal `json:"max_price,omitempty" gorm:"type:decimal(36,18)"` // 最高买入价格
	MinPrice     *decimal.Decimal `json:"min_price,omitempty" gorm:"type:decimal(36,18)"` // 最低买入价格
	StopLoss     *decimal.Decimal `json:"stop_loss,omitempty" gorm:"type:decimal(36,18)"` // 止损价
	TakeProfit   *decimal.Decimal `json:"take_profit,omitempty" gorm:"type:decimal(36,18)"` // 止盈价

	// 时间控制
	StartDate time.Time  `json:"start_date"`                 // 开始日期
	EndDate   *time.Time `json:"end_date,omitempty"`         // 结束日期 (可选)
	NextRun   time.Time  `json:"next_run"`                   // 下次执行时间

	// 统计数据
	TotalInvested     decimal.Decimal `json:"total_invested" gorm:"type:decimal(36,18);default:0"`      // 总投入
	TotalQuantity     decimal.Decimal `json:"total_quantity" gorm:"type:decimal(36,18);default:0"`      // 总持仓数量
	AverageCost       decimal.Decimal `json:"average_cost" gorm:"type:decimal(36,18);default:0"`        // 平均成本
	TotalExecutions   int             `json:"total_executions" gorm:"default:0"`                        // 总执行次数
	SuccessExecutions int             `json:"success_executions" gorm:"default:0"`                      // 成功次数
	FailedExecutions  int             `json:"failed_executions" gorm:"default:0"`                       // 失败次数

	Status     string    `json:"status" gorm:"index"` // pending, active, paused, stopped, completed
	CreateTime time.Time `json:"create_time" gorm:"index"`
	UpdateTime time.Time `json:"update_time"`
}

// DCAExecution represents a single DCA execution
type DCAExecution struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	StrategyID string          `json:"strategy_id" gorm:"index"`
	OrderID    *string         `json:"order_id,omitempty" gorm:"index"` // 订单ID
	Amount     decimal.Decimal `json:"amount" gorm:"type:decimal(36,18)"`
	Price      decimal.Decimal `json:"price" gorm:"type:decimal(36,18)"`
	Quantity   decimal.Decimal `json:"quantity" gorm:"type:decimal(36,18)"`
	Status     string          `json:"status"` // pending, success, failed, skipped
	Reason     string          `json:"reason,omitempty"` // 执行原因或失败原因
	ScheduledAt time.Time      `json:"scheduled_at"`
	ExecutedAt  *time.Time     `json:"executed_at,omitempty"`
	CreateTime  time.Time      `json:"create_time"`
}

// DCAService manages DCA strategies
type DCAService struct {
	orderService *OrderService
	mutex        sync.RWMutex
	db           *gorm.DB
}

// NewDCAService creates a new DCA service
func NewDCAService(orderService *OrderService, db *gorm.DB) *DCAService {
	return &DCAService{
		orderService: orderService,
		db:           db,
	}
}

// CreateDCAStrategy creates a new DCA strategy
func (s *DCAService) CreateDCAStrategy(
	ctx context.Context,
	userID uint,
	symbol string,
	amountPerPeriod decimal.Decimal,
	frequency string,
	startDate time.Time,
	endDate *time.Time,
	dayOfWeek, dayOfMonth *int,
	hourOfDay int,
	maxPrice, minPrice, stopLoss, takeProfit *decimal.Decimal,
) (*DCAStrategy, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证参数
	if err := s.validateParams(amountPerPeriod, frequency, dayOfWeek, dayOfMonth); err != nil {
		return nil, err
	}

	// 计算下次执行时间
	nextRun := s.calculateNextRun(startDate, frequency, dayOfWeek, dayOfMonth, hourOfDay)

	strategy := &DCAStrategy{
		ID:              fmt.Sprintf("DCA-%d-%d", userID, time.Now().UnixNano()),
		UserID:          userID,
		Symbol:          symbol,
		AmountPerPeriod: amountPerPeriod,
		Frequency:       frequency,
		DayOfWeek:       dayOfWeek,
		DayOfMonth:      dayOfMonth,
		HourOfDay:       hourOfDay,
		MaxPrice:        maxPrice,
		MinPrice:        minPrice,
		StopLoss:        stopLoss,
		TakeProfit:      takeProfit,
		StartDate:       startDate,
		EndDate:         endDate,
		NextRun:         nextRun,
		TotalInvested:   decimal.Zero,
		TotalQuantity:   decimal.Zero,
		AverageCost:     decimal.Zero,
		Status:          "pending",
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
	}

	// 保存策略
	if err := s.db.Create(strategy).Error; err != nil {
		return nil, fmt.Errorf("failed to create DCA strategy: %w", err)
	}

	// 启动执行器
	go s.runDCAStrategy(ctx, strategy.ID)

	return strategy, nil
}

// validateParams validates DCA strategy parameters
func (s *DCAService) validateParams(
	amountPerPeriod decimal.Decimal,
	frequency string,
	dayOfWeek, dayOfMonth *int,
) error {
	if amountPerPeriod.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount per period must be positive")
	}

	validFrequencies := map[string]bool{"daily": true, "weekly": true, "monthly": true}
	if !validFrequencies[frequency] {
		return errors.New("frequency must be daily, weekly, or monthly")
	}

	if frequency == "weekly" && (dayOfWeek == nil || *dayOfWeek < 0 || *dayOfWeek > 6) {
		return errors.New("day of week must be between 0 (Sunday) and 6 (Saturday) for weekly frequency")
	}

	if frequency == "monthly" && (dayOfMonth == nil || *dayOfMonth < 1 || *dayOfMonth > 31) {
		return errors.New("day of month must be between 1 and 31 for monthly frequency")
	}

	return nil
}

// calculateNextRun calculates the next execution time
func (s *DCAService) calculateNextRun(
	startDate time.Time,
	frequency string,
	dayOfWeek, dayOfMonth *int,
	hourOfDay int,
) time.Time {
	now := time.Now()
	next := startDate

	// 如果开始日期在过去,从现在开始计算
	if next.Before(now) {
		next = now
	}

	// 设置执行时间为指定小时
	next = time.Date(next.Year(), next.Month(), next.Day(), hourOfDay, 0, 0, 0, next.Location())

	switch frequency {
	case "daily":
		// 如果今天的执行时间已过,设置为明天
		if next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}

	case "weekly":
		if dayOfWeek != nil {
			// 计算到下一个指定星期几
			daysUntilTarget := (*dayOfWeek - int(next.Weekday()) + 7) % 7
			if daysUntilTarget == 0 && next.Before(now) {
				daysUntilTarget = 7
			}
			next = next.AddDate(0, 0, daysUntilTarget)
		}

	case "monthly":
		if dayOfMonth != nil {
			// 设置为本月或下月的指定日期
			next = time.Date(next.Year(), next.Month(), *dayOfMonth, hourOfDay, 0, 0, 0, next.Location())
			if next.Before(now) {
				// 如果本月日期已过,设置为下月
				next = next.AddDate(0, 1, 0)
			}
		}
	}

	return next
}

// runDCAStrategy runs the DCA strategy
func (s *DCAService) runDCAStrategy(ctx context.Context, strategyID string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			shouldStop := s.checkAndExecuteDCA(ctx, strategyID)
			if shouldStop {
				return
			}
		}
	}
}

// checkAndExecuteDCA checks if DCA should execute and executes it
func (s *DCAService) checkAndExecuteDCA(ctx context.Context, strategyID string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var strategy DCAStrategy
	if err := s.db.First(&strategy, "id = ?", strategyID).Error; err != nil {
		return true
	}

	// 检查状态
	if strategy.Status != "pending" && strategy.Status != "active" {
		return true
	}

	now := time.Now()

	// 检查是否到达结束时间
	if strategy.EndDate != nil && now.After(*strategy.EndDate) {
		strategy.Status = "completed"
		strategy.UpdateTime = now
		s.db.Save(&strategy)
		return true
	}

	// 检查是否到达执行时间
	if now.Before(strategy.NextRun) {
		return false
	}

	// 激活策略
	if strategy.Status == "pending" {
		strategy.Status = "active"
	}

	// 执行定投
	execution := &DCAExecution{
		StrategyID:  strategy.ID,
		Amount:      strategy.AmountPerPeriod,
		Status:      "pending",
		ScheduledAt: strategy.NextRun,
		CreateTime:  now,
	}

	err := s.executeDCA(ctx, &strategy, execution)
	s.db.Create(execution)

	if err != nil {
		strategy.FailedExecutions++
		execution.Status = "failed"
		execution.Reason = err.Error()
		s.db.Save(execution)
	} else {
		strategy.SuccessExecutions++
	}

	strategy.TotalExecutions++

	// 计算下次执行时间
	strategy.NextRun = s.calculateNextRun(
		strategy.NextRun.AddDate(0, 0, 1), // 从下一天开始计算
		strategy.Frequency,
		strategy.DayOfWeek,
		strategy.DayOfMonth,
		strategy.HourOfDay,
	)

	strategy.UpdateTime = now
	s.db.Save(&strategy)

	return false
}

// executeDCA executes a DCA order
func (s *DCAService) executeDCA(ctx context.Context, strategy *DCAStrategy, execution *DCAExecution) error {
	// 获取当前市场价格
	var lastTrade models.Trade
	if err := s.db.Where("symbol = ?", strategy.Symbol).
		Order("trade_time DESC").
		First(&lastTrade).Error; err != nil {
		return fmt.Errorf("failed to get market price: %w", err)
	}

	currentPrice := lastTrade.Price
	execution.Price = currentPrice

	// 检查价格条件
	if strategy.MaxPrice != nil && currentPrice.GreaterThan(*strategy.MaxPrice) {
		execution.Status = "skipped"
		execution.Reason = fmt.Sprintf("Price %.2f exceeds max price %.2f",
			currentPrice.InexactFloat64(), strategy.MaxPrice.InexactFloat64())
		return errors.New(execution.Reason)
	}

	if strategy.MinPrice != nil && currentPrice.LessThan(*strategy.MinPrice) {
		execution.Status = "skipped"
		execution.Reason = fmt.Sprintf("Price %.2f below min price %.2f",
			currentPrice.InexactFloat64(), strategy.MinPrice.InexactFloat64())
		return errors.New(execution.Reason)
	}

	// 检查止损止盈
	if strategy.StopLoss != nil && currentPrice.LessThanOrEqual(*strategy.StopLoss) {
		strategy.Status = "stopped"
		execution.Status = "skipped"
		execution.Reason = "Stop loss triggered"
		return errors.New(execution.Reason)
	}

	if strategy.TakeProfit != nil && currentPrice.GreaterThanOrEqual(*strategy.TakeProfit) {
		strategy.Status = "stopped"
		execution.Status = "skipped"
		execution.Reason = "Take profit triggered"
		return errors.New(execution.Reason)
	}

	// 计算购买数量
	quantity := execution.Amount.Div(currentPrice)
	execution.Quantity = quantity

	// 创建市价买单
	order := &models.Order{
		UserID:      strategy.UserID,
		Symbol:      strategy.Symbol,
		Side:        models.OrderSideBuy,
		Type:        models.OrderTypeMarket,
		Price:       currentPrice,
		Quantity:    quantity,
		Status:      models.OrderStatusPending,
		TimeInForce: models.TimeInForceIOC,
		CreateTime:  time.Now(),
	}

	createdOrder, _, err := s.orderService.CreateOrder(order)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	execution.OrderID = &createdOrder.ID
	execution.Status = "success"
	now := time.Now()
	execution.ExecutedAt = &now

	// 更新策略统计
	strategy.TotalInvested = strategy.TotalInvested.Add(createdOrder.FilledAmount)
	strategy.TotalQuantity = strategy.TotalQuantity.Add(createdOrder.FilledQty)

	if strategy.TotalQuantity.GreaterThan(decimal.Zero) {
		strategy.AverageCost = strategy.TotalInvested.Div(strategy.TotalQuantity)
	}

	return nil
}

// PauseDCAStrategy pauses a DCA strategy
func (s *DCAService) PauseDCAStrategy(ctx context.Context, strategyID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var strategy DCAStrategy
	if err := s.db.First(&strategy, "id = ?", strategyID).Error; err != nil {
		return fmt.Errorf("strategy not found: %w", err)
	}

	if strategy.Status != "active" {
		return fmt.Errorf("cannot pause strategy with status: %s", strategy.Status)
	}

	strategy.Status = "paused"
	strategy.UpdateTime = time.Now()

	return s.db.Save(&strategy).Error
}

// ResumeDCAStrategy resumes a paused DCA strategy
func (s *DCAService) ResumeDCAStrategy(ctx context.Context, strategyID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var strategy DCAStrategy
	if err := s.db.First(&strategy, "id = ?", strategyID).Error; err != nil {
		return fmt.Errorf("strategy not found: %w", err)
	}

	if strategy.Status != "paused" {
		return fmt.Errorf("cannot resume strategy with status: %s", strategy.Status)
	}

	strategy.Status = "active"
	strategy.UpdateTime = time.Now()

	// 重新计算下次执行时间
	strategy.NextRun = s.calculateNextRun(
		time.Now(),
		strategy.Frequency,
		strategy.DayOfWeek,
		strategy.DayOfMonth,
		strategy.HourOfDay,
	)

	return s.db.Save(&strategy).Error
}

// StopDCAStrategy stops a DCA strategy
func (s *DCAService) StopDCAStrategy(ctx context.Context, strategyID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var strategy DCAStrategy
	if err := s.db.First(&strategy, "id = ?", strategyID).Error; err != nil {
		return fmt.Errorf("strategy not found: %w", err)
	}

	strategy.Status = "stopped"
	strategy.UpdateTime = time.Now()

	return s.db.Save(&strategy).Error
}

// GetDCAStrategy retrieves a DCA strategy by ID
func (s *DCAService) GetDCAStrategy(ctx context.Context, strategyID string) (*DCAStrategy, error) {
	var strategy DCAStrategy
	if err := s.db.First(&strategy, "id = ?", strategyID).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

// GetDCAExecutions retrieves all executions for a strategy
func (s *DCAService) GetDCAExecutions(ctx context.Context, strategyID string, limit, offset int) ([]DCAExecution, int64, error) {
	var executions []DCAExecution
	var total int64

	if err := s.db.Model(&DCAExecution{}).Where("strategy_id = ?", strategyID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Where("strategy_id = ?", strategyID).
		Order("scheduled_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&executions).Error; err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}

// GetUserDCAStrategies retrieves all DCA strategies for a user
func (s *DCAService) GetUserDCAStrategies(ctx context.Context, userID uint, limit, offset int) ([]DCAStrategy, int64, error) {
	var strategies []DCAStrategy
	var total int64

	if err := s.db.Model(&DCAStrategy{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
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
