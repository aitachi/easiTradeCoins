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

// Trader represents a trader who publishes strategies
// 交易员: 发布交易策略供其他人跟随
type Trader struct {
	ID                uint            `json:"id" gorm:"primaryKey"`
	UserID            uint            `json:"user_id" gorm:"uniqueIndex"`
	Username          string          `json:"username"`
	Description       string          `json:"description" gorm:"type:text"`
	ROI               decimal.Decimal `json:"roi" gorm:"type:decimal(10,4);default:0"`           // 收益率
	TotalPnL          decimal.Decimal `json:"total_pnl" gorm:"type:decimal(36,18);default:0"`    // 总盈亏
	WinRate           decimal.Decimal `json:"win_rate" gorm:"type:decimal(5,4);default:0"`       // 胜率
	Followers         int             `json:"followers" gorm:"default:0"`                         // 跟随者数量
	TotalTrades       int             `json:"total_trades" gorm:"default:0"`                      // 总交易次数
	ProfitTrades      int             `json:"profit_trades" gorm:"default:0"`                     // 盈利交易次数
	LossTrades        int             `json:"loss_trades" gorm:"default:0"`                       // 亏损交易次数
	MaxDrawdown       decimal.Decimal `json:"max_drawdown" gorm:"type:decimal(10,4);default:0"`  // 最大回撤
	SharpeRatio       decimal.Decimal `json:"sharpe_ratio" gorm:"type:decimal(10,4);default:0"`  // 夏普比率
	IsActive          bool            `json:"is_active" gorm:"default:true"`                      // 是否开放跟单
	MinFollowAmount   decimal.Decimal `json:"min_follow_amount" gorm:"type:decimal(36,18)"`      // 最小跟随金额
	MaxFollowers      int             `json:"max_followers" gorm:"default:1000"`                  // 最大跟随人数
	CommissionRate    decimal.Decimal `json:"commission_rate" gorm:"type:decimal(5,4);default:0"` // 分成比例
	Ranking           int             `json:"ranking" gorm:"default:0"`                           // 排名
	VerificationLevel int             `json:"verification_level" gorm:"default:0"`                // 认证等级
	CreateTime        time.Time       `json:"create_time"`
	UpdateTime        time.Time       `json:"update_time"`
}

// FollowRelation represents a follower-trader relationship
// 跟单关系
type FollowRelation struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	FollowerID      uint            `json:"follower_id" gorm:"index"`                        // 跟随者ID
	TraderID        uint            `json:"trader_id" gorm:"index"`                          // 交易员ID
	AllocationRatio decimal.Decimal `json:"allocation_ratio" gorm:"type:decimal(5,4)"`       // 跟随比例 (0-1)
	MaxPerTrade     decimal.Decimal `json:"max_per_trade" gorm:"type:decimal(36,18)"`        // 单笔最大金额
	StopLoss        *decimal.Decimal `json:"stop_loss,omitempty" gorm:"type:decimal(10,4)"`   // 止损比例
	TakeProfit      *decimal.Decimal `json:"take_profit,omitempty" gorm:"type:decimal(10,4)"` // 止盈比例
	IsActive        bool            `json:"is_active" gorm:"default:true"`                    // 是否活跃
	TotalCopied     int             `json:"total_copied" gorm:"default:0"`                    // 已复制订单数
	TotalProfit     decimal.Decimal `json:"total_profit" gorm:"type:decimal(36,18);default:0"` // 总收益
	CreateTime      time.Time       `json:"create_time"`
	UpdateTime      time.Time       `json:"update_time"`
}

// CopiedOrder represents an order copied from a trader
// 复制的订单
type CopiedOrder struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	FollowerID      uint            `json:"follower_id" gorm:"index"`
	TraderID        uint            `json:"trader_id" gorm:"index"`
	OriginalOrderID string          `json:"original_order_id" gorm:"index"` // 原始订单ID
	CopiedOrderID   string          `json:"copied_order_id" gorm:"index"`   // 复制的订单ID
	Symbol          string          `json:"symbol"`
	Side            string          `json:"side"`
	Quantity        decimal.Decimal `json:"quantity" gorm:"type:decimal(36,18)"`
	Price           decimal.Decimal `json:"price" gorm:"type:decimal(36,18)"`
	Status          string          `json:"status"` // pending/filled/cancelled
	PnL             decimal.Decimal `json:"pnl" gorm:"type:decimal(36,18);default:0"`
	CreateTime      time.Time       `json:"create_time"`
	UpdateTime      time.Time       `json:"update_time"`
}

// TradingStrategy represents a published trading strategy
// 交易策略
type TradingStrategy struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	TraderID      uint            `json:"trader_id" gorm:"index"`
	Name          string          `json:"name"`
	Description   string          `json:"description" gorm:"type:text"`
	Category      string          `json:"category"` // trend/grid/scalping/etc
	RiskLevel     string          `json:"risk_level"` // low/medium/high
	MinInvestment decimal.Decimal `json:"min_investment" gorm:"type:decimal(36,18)"`
	ROI           decimal.Decimal `json:"roi" gorm:"type:decimal(10,4);default:0"`
	Subscribers   int             `json:"subscribers" gorm:"default:0"`
	IsPublic      bool            `json:"is_public" gorm:"default:true"`
	CreateTime    time.Time       `json:"create_time"`
	UpdateTime    time.Time       `json:"update_time"`
}

// CopyTradingService manages copy trading functionality
type CopyTradingService struct {
	orderService *OrderService
	mutex        sync.RWMutex
	db           *gorm.DB
}

// NewCopyTradingService creates a new copy trading service
func NewCopyTradingService(orderService *OrderService, db *gorm.DB) *CopyTradingService {
	return &CopyTradingService{
		orderService: orderService,
		db:           db,
	}
}

// RegisterTrader registers a user as a trader
func (s *CopyTradingService) RegisterTrader(
	ctx context.Context,
	userID uint,
	username, description string,
	minFollowAmount decimal.Decimal,
	commissionRate decimal.Decimal,
) (*Trader, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if already registered
	var existing Trader
	err := s.db.Where("user_id = ?", userID).First(&existing).Error
	if err == nil {
		return nil, errors.New("user is already registered as a trader")
	}

	trader := &Trader{
		UserID:            userID,
		Username:          username,
		Description:       description,
		ROI:               decimal.Zero,
		TotalPnL:          decimal.Zero,
		WinRate:           decimal.Zero,
		Followers:         0,
		TotalTrades:       0,
		ProfitTrades:      0,
		LossTrades:        0,
		MaxDrawdown:       decimal.Zero,
		SharpeRatio:       decimal.Zero,
		IsActive:          true,
		MinFollowAmount:   minFollowAmount,
		MaxFollowers:      1000,
		CommissionRate:    commissionRate,
		Ranking:           0,
		VerificationLevel: 0,
		CreateTime:        time.Now(),
		UpdateTime:        time.Now(),
	}

	if err := s.db.Create(trader).Error; err != nil {
		return nil, err
	}

	return trader, nil
}

// FollowTrader creates a follow relationship between follower and trader
func (s *CopyTradingService) FollowTrader(
	ctx context.Context,
	followerID, traderID uint,
	allocationRatio decimal.Decimal,
	maxPerTrade decimal.Decimal,
	stopLoss, takeProfit *decimal.Decimal,
) (*FollowRelation, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate trader exists and is active
	var trader Trader
	if err := s.db.First(&trader, traderID).Error; err != nil {
		return nil, errors.New("trader not found")
	}

	if !trader.IsActive {
		return nil, errors.New("trader is not accepting new followers")
	}

	if trader.Followers >= trader.MaxFollowers {
		return nil, errors.New("trader has reached maximum followers")
	}

	// Check if already following
	var existing FollowRelation
	err := s.db.Where("follower_id = ? AND trader_id = ?", followerID, traderID).First(&existing).Error
	if err == nil {
		return nil, errors.New("already following this trader")
	}

	// Validate allocation ratio
	if allocationRatio.LessThanOrEqual(decimal.Zero) || allocationRatio.GreaterThan(decimal.NewFromInt(1)) {
		return nil, errors.New("allocation ratio must be between 0 and 1")
	}

	relation := &FollowRelation{
		FollowerID:      followerID,
		TraderID:        traderID,
		AllocationRatio: allocationRatio,
		MaxPerTrade:     maxPerTrade,
		StopLoss:        stopLoss,
		TakeProfit:      takeProfit,
		IsActive:        true,
		TotalCopied:     0,
		TotalProfit:     decimal.Zero,
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
	}

	return relation, s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(relation).Error; err != nil {
			return err
		}

		// Update trader followers count
		trader.Followers++
		trader.UpdateTime = time.Now()
		return tx.Save(&trader).Error
	})
}

// UnfollowTrader removes a follow relationship
func (s *CopyTradingService) UnfollowTrader(ctx context.Context, followerID, traderID uint) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var relation FollowRelation
	if err := s.db.Where("follower_id = ? AND trader_id = ?", followerID, traderID).First(&relation).Error; err != nil {
		return errors.New("follow relationship not found")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete relation
		if err := tx.Delete(&relation).Error; err != nil {
			return err
		}

		// Update trader followers count
		var trader Trader
		if err := tx.First(&trader, traderID).Error; err != nil {
			return err
		}

		trader.Followers--
		if trader.Followers < 0 {
			trader.Followers = 0
		}
		trader.UpdateTime = time.Now()

		return tx.Save(&trader).Error
	})
}

// CopyOrder copies an order from a trader to followers
func (s *CopyTradingService) CopyOrder(ctx context.Context, traderID uint, originalOrder *models.Order) error {
	// Get all active followers of this trader
	var relations []FollowRelation
	if err := s.db.Where("trader_id = ? AND is_active = ?", traderID, true).Find(&relations).Error; err != nil {
		return err
	}

	for _, relation := range relations {
		// Calculate copy quantity based on allocation ratio
		copyQuantity := originalOrder.Quantity.Mul(relation.AllocationRatio)

		// Check max per trade limit
		tradeValue := copyQuantity.Mul(originalOrder.Price)
		if tradeValue.GreaterThan(relation.MaxPerTrade) {
			copyQuantity = relation.MaxPerTrade.Div(originalOrder.Price)
		}

		// Create copied order
		copiedOrder := &models.Order{
			UserID:      relation.FollowerID,
			Symbol:      originalOrder.Symbol,
			Side:        originalOrder.Side,
			Type:        originalOrder.Type,
			Price:       originalOrder.Price,
			Quantity:    copyQuantity,
			Status:      models.OrderStatusPending,
			TimeInForce: originalOrder.TimeInForce,
			CreateTime:  time.Now(),
		}

		// Execute the copied order
		created, _, err := s.orderService.CreateOrder(copiedOrder)
		if err != nil {
			// Log error but continue with other followers
			fmt.Printf("Failed to copy order for follower %d: %v\n", relation.FollowerID, err)
			continue
		}

		// Record the copied order
		copied := &CopiedOrder{
			FollowerID:      relation.FollowerID,
			TraderID:        traderID,
			OriginalOrderID: originalOrder.ID,
			CopiedOrderID:   created.ID,
			Symbol:          originalOrder.Symbol,
			Side:            string(originalOrder.Side),
			Quantity:        copyQuantity,
			Price:           originalOrder.Price,
			Status:          "pending",
			PnL:             decimal.Zero,
			CreateTime:      time.Now(),
			UpdateTime:      time.Now(),
		}

		s.db.Create(copied)

		// Update follow relation
		relation.TotalCopied++
		relation.UpdateTime = time.Now()
		s.db.Save(&relation)
	}

	return nil
}

// UpdateTraderStats updates trader statistics
func (s *CopyTradingService) UpdateTraderStats(ctx context.Context, traderID uint, pnl decimal.Decimal, isWin bool) error {
	var trader Trader
	if err := s.db.First(&trader, traderID).Error; err != nil {
		return err
	}

	trader.TotalTrades++
	trader.TotalPnL = trader.TotalPnL.Add(pnl)

	if isWin {
		trader.ProfitTrades++
	} else {
		trader.LossTrades++
	}

	// Calculate win rate
	if trader.TotalTrades > 0 {
		trader.WinRate = decimal.NewFromInt(int64(trader.ProfitTrades)).Div(
			decimal.NewFromInt(int64(trader.TotalTrades)),
		)
	}

	// TODO: Calculate ROI, max drawdown, Sharpe ratio
	// These require historical equity data

	trader.UpdateTime = time.Now()

	return s.db.Save(&trader).Error
}

// GetTopTraders gets top traders by ROI
func (s *CopyTradingService) GetTopTraders(ctx context.Context, limit int) ([]Trader, error) {
	var traders []Trader
	if err := s.db.Where("is_active = ?", true).
		Order("roi DESC, followers DESC").
		Limit(limit).
		Find(&traders).Error; err != nil {
		return nil, err
	}

	return traders, nil
}

// GetTraderFollowers gets all followers of a trader
func (s *CopyTradingService) GetTraderFollowers(ctx context.Context, traderID uint) ([]FollowRelation, error) {
	var relations []FollowRelation
	if err := s.db.Where("trader_id = ? AND is_active = ?", traderID, true).
		Order("create_time DESC").
		Find(&relations).Error; err != nil {
		return nil, err
	}

	return relations, nil
}

// GetFollowerTraders gets all traders a user is following
func (s *CopyTradingService) GetFollowerTraders(ctx context.Context, followerID uint) ([]FollowRelation, error) {
	var relations []FollowRelation
	if err := s.db.Where("follower_id = ? AND is_active = ?", followerID, true).
		Order("create_time DESC").
		Find(&relations).Error; err != nil {
		return nil, err
	}

	return relations, nil
}

// GetCopiedOrders gets all copied orders for a follower
func (s *CopyTradingService) GetCopiedOrders(ctx context.Context, followerID uint, limit, offset int) ([]CopiedOrder, int64, error) {
	var orders []CopiedOrder
	var total int64

	if err := s.db.Model(&CopiedOrder{}).Where("follower_id = ?", followerID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Where("follower_id = ?", followerID).
		Order("create_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// PublishStrategy publishes a trading strategy
func (s *CopyTradingService) PublishStrategy(
	ctx context.Context,
	traderID uint,
	name, description, category, riskLevel string,
	minInvestment decimal.Decimal,
) (*TradingStrategy, error) {
	// Verify trader exists
	var trader Trader
	if err := s.db.First(&trader, traderID).Error; err != nil {
		return nil, errors.New("trader not found")
	}

	strategy := &TradingStrategy{
		TraderID:      traderID,
		Name:          name,
		Description:   description,
		Category:      category,
		RiskLevel:     riskLevel,
		MinInvestment: minInvestment,
		ROI:           decimal.Zero,
		Subscribers:   0,
		IsPublic:      true,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
	}

	if err := s.db.Create(strategy).Error; err != nil {
		return nil, err
	}

	return strategy, nil
}

// GetTraderStrategies gets all strategies published by a trader
func (s *CopyTradingService) GetTraderStrategies(ctx context.Context, traderID uint) ([]TradingStrategy, error) {
	var strategies []TradingStrategy
	if err := s.db.Where("trader_id = ? AND is_public = ?", traderID, true).
		Order("subscribers DESC").
		Find(&strategies).Error; err != nil {
		return nil, err
	}

	return strategies, nil
}

// GetTraderPerformance gets detailed performance metrics for a trader
func (s *CopyTradingService) GetTraderPerformance(ctx context.Context, traderID uint) (map[string]interface{}, error) {
	var trader Trader
	if err := s.db.First(&trader, traderID).Error; err != nil {
		return nil, err
	}

	// Get recent trades (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var recentOrders []models.Order
	s.db.Where("user_id = ? AND create_time > ? AND status = ?",
		trader.UserID, thirtyDaysAgo, models.OrderStatusFilled).
		Order("create_time DESC").
		Limit(100).
		Find(&recentOrders)

	performance := map[string]interface{}{
		"trader_id":          trader.ID,
		"username":           trader.Username,
		"roi":                trader.ROI.String(),
		"total_pnl":          trader.TotalPnL.String(),
		"win_rate":           trader.WinRate.Mul(decimal.NewFromInt(100)).String() + "%",
		"total_trades":       trader.TotalTrades,
		"profit_trades":      trader.ProfitTrades,
		"loss_trades":        trader.LossTrades,
		"followers":          trader.Followers,
		"max_drawdown":       trader.MaxDrawdown.String(),
		"sharpe_ratio":       trader.SharpeRatio.String(),
		"recent_trades":      len(recentOrders),
		"commission_rate":    trader.CommissionRate.Mul(decimal.NewFromInt(100)).String() + "%",
		"verification_level": trader.VerificationLevel,
		"ranking":            trader.Ranking,
	}

	return performance, nil
}
