package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MarginAccount represents a margin trading account
// 保证金账户: 用于杠杆交易
type MarginAccount struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	UserID          uint            `json:"user_id" gorm:"uniqueIndex"`
	Collateral      decimal.Decimal `json:"collateral" gorm:"type:decimal(36,18);default:0"`       // 抵押资产
	Borrowed        decimal.Decimal `json:"borrowed" gorm:"type:decimal(36,18);default:0"`         // 借贷金额
	Interest        decimal.Decimal `json:"interest" gorm:"type:decimal(36,18);default:0"`         // 累计利息
	Equity          decimal.Decimal `json:"equity" gorm:"type:decimal(36,18);default:0"`           // 净资产
	MarginLevel     decimal.Decimal `json:"margin_level" gorm:"type:decimal(10,4);default:0"`      // 保证金率
	Leverage        int             `json:"leverage" gorm:"default:1"`                              // 杠杆倍数 (1x, 2x, 3x, 5x, 10x)
	MaxLeverage     int             `json:"max_leverage" gorm:"default:10"`                         // 最大杠杆
	MaintenanceRate decimal.Decimal `json:"maintenance_rate" gorm:"type:decimal(10,4);default:0.1"` // 维持保证金率 (10%)
	Status          string          `json:"status" gorm:"default:active"`                           // active/liquidating/liquidated
	CreateTime      time.Time       `json:"create_time"`
	UpdateTime      time.Time       `json:"update_time"`
}

// MarginPosition represents a margin trading position
// 杠杆持仓
type MarginPosition struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	UserID          uint            `json:"user_id" gorm:"index"`
	Symbol          string          `json:"symbol" gorm:"index"`
	Side            string          `json:"side"` // long/short
	EntryPrice      decimal.Decimal `json:"entry_price" gorm:"type:decimal(36,18)"`
	CurrentPrice    decimal.Decimal `json:"current_price" gorm:"type:decimal(36,18)"`
	Quantity        decimal.Decimal `json:"quantity" gorm:"type:decimal(36,18)"`
	Leverage        int             `json:"leverage"`
	Margin          decimal.Decimal `json:"margin" gorm:"type:decimal(36,18)"`           // 保证金
	UnrealizedPnL   decimal.Decimal `json:"unrealized_pnl" gorm:"type:decimal(36,18)"`  // 未实现盈亏
	RealizedPnL     decimal.Decimal `json:"realized_pnl" gorm:"type:decimal(36,18)"`    // 已实现盈亏
	LiquidationPrice decimal.Decimal `json:"liquidation_price" gorm:"type:decimal(36,18)"` // 强平价格
	StopLoss        *decimal.Decimal `json:"stop_loss,omitempty" gorm:"type:decimal(36,18)"`
	TakeProfit      *decimal.Decimal `json:"take_profit,omitempty" gorm:"type:decimal(36,18)"`
	Status          string          `json:"status" gorm:"default:open"` // open/closed/liquidated
	OpenTime        time.Time       `json:"open_time"`
	CloseTime       *time.Time      `json:"close_time,omitempty"`
	UpdateTime      time.Time       `json:"update_time"`
}

// MarginLoan represents a margin loan
// 借贷记录
type MarginLoan struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	UserID        uint            `json:"user_id" gorm:"index"`
	Currency      string          `json:"currency"`
	Principal     decimal.Decimal `json:"principal" gorm:"type:decimal(36,18)"`      // 本金
	Interest      decimal.Decimal `json:"interest" gorm:"type:decimal(36,18)"`       // 利息
	InterestRate  decimal.Decimal `json:"interest_rate" gorm:"type:decimal(10,6)"`   // 日利率
	Repaid        decimal.Decimal `json:"repaid" gorm:"type:decimal(36,18);default:0"` // 已还款
	Status        string          `json:"status" gorm:"default:active"`              // active/repaid
	BorrowTime    time.Time       `json:"borrow_time"`
	RepayTime     *time.Time      `json:"repay_time,omitempty"`
	LastAccrueTime time.Time      `json:"last_accrue_time"` // 最后计息时间
}

// MarginTradingService manages margin trading
type MarginTradingService struct {
	orderService *OrderService
	mutex        sync.RWMutex
	db           *gorm.DB
}

// NewMarginTradingService creates a new margin trading service
func NewMarginTradingService(orderService *OrderService, db *gorm.DB) *MarginTradingService {
	return &MarginTradingService{
		orderService: orderService,
		db:           db,
	}
}

// GetOrCreateMarginAccount gets or creates a margin account for a user
func (s *MarginTradingService) GetOrCreateMarginAccount(ctx context.Context, userID uint) (*MarginAccount, error) {
	var account MarginAccount
	err := s.db.Where("user_id = ?", userID).First(&account).Error

	if err == gorm.ErrRecordNotFound {
		// Create new account
		account = MarginAccount{
			UserID:          userID,
			Collateral:      decimal.Zero,
			Borrowed:        decimal.Zero,
			Interest:        decimal.Zero,
			Equity:          decimal.Zero,
			MarginLevel:     decimal.Zero,
			Leverage:        1,
			MaxLeverage:     10,
			MaintenanceRate: decimal.NewFromFloat(0.1), // 10%
			Status:          "active",
			CreateTime:      time.Now(),
			UpdateTime:      time.Now(),
		}
		if err := s.db.Create(&account).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &account, nil
}

// Deposit deposits collateral to margin account
func (s *MarginTradingService) Deposit(ctx context.Context, userID uint, amount decimal.Decimal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	account, err := s.GetOrCreateMarginAccount(ctx, userID)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Update collateral
		account.Collateral = account.Collateral.Add(amount)
		account.Equity = account.Collateral.Sub(account.Borrowed).Sub(account.Interest)
		account.UpdateTime = time.Now()

		if account.Borrowed.GreaterThan(decimal.Zero) {
			account.MarginLevel = account.Equity.Div(account.Borrowed)
		}

		return tx.Save(account).Error
	})
}

// Withdraw withdraws collateral from margin account
func (s *MarginTradingService) Withdraw(ctx context.Context, userID uint, amount decimal.Decimal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	account, err := s.GetOrCreateMarginAccount(ctx, userID)
	if err != nil {
		return err
	}

	// Check if withdrawal would violate margin requirements
	newCollateral := account.Collateral.Sub(amount)
	newEquity := newCollateral.Sub(account.Borrowed).Sub(account.Interest)

	if account.Borrowed.GreaterThan(decimal.Zero) {
		newMarginLevel := newEquity.Div(account.Borrowed)
		if newMarginLevel.LessThan(account.MaintenanceRate) {
			return errors.New("withdrawal would violate margin requirements")
		}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		account.Collateral = newCollateral
		account.Equity = newEquity
		account.UpdateTime = time.Now()

		if account.Borrowed.GreaterThan(decimal.Zero) {
			account.MarginLevel = account.Equity.Div(account.Borrowed)
		}

		return tx.Save(account).Error
	})
}

// Borrow borrows funds for margin trading
func (s *MarginTradingService) Borrow(ctx context.Context, userID uint, currency string, amount decimal.Decimal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	account, err := s.GetOrCreateMarginAccount(ctx, userID)
	if err != nil {
		return err
	}

	// Check borrowing capacity
	maxBorrow := account.Collateral.Mul(decimal.NewFromInt(int64(account.Leverage))).Sub(account.Borrowed)
	if amount.GreaterThan(maxBorrow) {
		return fmt.Errorf("borrowing amount exceeds capacity: max=%s", maxBorrow.String())
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create loan record
		loan := MarginLoan{
			UserID:         userID,
			Currency:       currency,
			Principal:      amount,
			Interest:       decimal.Zero,
			InterestRate:   decimal.NewFromFloat(0.0001), // 0.01% per day
			Repaid:         decimal.Zero,
			Status:         "active",
			BorrowTime:     time.Now(),
			LastAccrueTime: time.Now(),
		}

		if err := tx.Create(&loan).Error; err != nil {
			return err
		}

		// Update account
		account.Borrowed = account.Borrowed.Add(amount)
		account.Equity = account.Collateral.Sub(account.Borrowed).Sub(account.Interest)
		account.MarginLevel = account.Equity.Div(account.Borrowed)
		account.UpdateTime = time.Now()

		return tx.Save(account).Error
	})
}

// Repay repays a margin loan
func (s *MarginTradingService) Repay(ctx context.Context, loanID uint, amount decimal.Decimal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var loan MarginLoan
	if err := s.db.First(&loan, loanID).Error; err != nil {
		return err
	}

	// Accrue interest first
	s.accrueInterest(&loan)

	totalOwed := loan.Principal.Add(loan.Interest).Sub(loan.Repaid)
	if amount.GreaterThan(totalOwed) {
		amount = totalOwed
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		loan.Repaid = loan.Repaid.Add(amount)

		if loan.Repaid.GreaterThanOrEqual(loan.Principal.Add(loan.Interest)) {
			loan.Status = "repaid"
			now := time.Now()
			loan.RepayTime = &now
		}

		if err := tx.Save(&loan).Error; err != nil {
			return err
		}

		// Update account
		var account MarginAccount
		if err := tx.Where("user_id = ?", loan.UserID).First(&account).Error; err != nil {
			return err
		}

		account.Borrowed = account.Borrowed.Sub(amount)
		account.Equity = account.Collateral.Sub(account.Borrowed).Sub(account.Interest)
		account.UpdateTime = time.Now()

		if account.Borrowed.GreaterThan(decimal.Zero) {
			account.MarginLevel = account.Equity.Div(account.Borrowed)
		}

		return tx.Save(&account).Error
	})
}

// accrueInterest accrues interest on a loan
func (s *MarginTradingService) accrueInterest(loan *MarginLoan) {
	now := time.Now()
	days := now.Sub(loan.LastAccrueTime).Hours() / 24.0

	if days > 0 {
		interest := loan.Principal.Mul(loan.InterestRate).Mul(decimal.NewFromFloat(days))
		loan.Interest = loan.Interest.Add(interest)
		loan.LastAccrueTime = now
	}
}

// OpenPosition opens a margin position
func (s *MarginTradingService) OpenPosition(
	ctx context.Context,
	userID uint,
	symbol string,
	side string, // long or short
	price decimal.Decimal,
	quantity decimal.Decimal,
	leverage int,
	stopLoss, takeProfit *decimal.Decimal,
) (*MarginPosition, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	account, err := s.GetOrCreateMarginAccount(ctx, userID)
	if err != nil {
		return nil, err
	}

	if leverage > account.MaxLeverage {
		return nil, fmt.Errorf("leverage %d exceeds max leverage %d", leverage, account.MaxLeverage)
	}

	// Calculate required margin
	positionValue := price.Mul(quantity)
	requiredMargin := positionValue.Div(decimal.NewFromInt(int64(leverage)))

	// Check available margin
	availableMargin := account.Equity
	if requiredMargin.GreaterThan(availableMargin) {
		return nil, errors.New("insufficient margin")
	}

	// Calculate liquidation price
	liquidationPrice := s.calculateLiquidationPrice(price, leverage, side, account.MaintenanceRate)

	position := &MarginPosition{
		UserID:           userID,
		Symbol:           symbol,
		Side:             side,
		EntryPrice:       price,
		CurrentPrice:     price,
		Quantity:         quantity,
		Leverage:         leverage,
		Margin:           requiredMargin,
		UnrealizedPnL:    decimal.Zero,
		RealizedPnL:      decimal.Zero,
		LiquidationPrice: liquidationPrice,
		StopLoss:         stopLoss,
		TakeProfit:       takeProfit,
		Status:           "open",
		OpenTime:         time.Now(),
		UpdateTime:       time.Now(),
	}

	if err := s.db.Create(position).Error; err != nil {
		return nil, err
	}

	return position, nil
}

// calculateLiquidationPrice calculates the liquidation price for a position
func (s *MarginTradingService) calculateLiquidationPrice(
	entryPrice decimal.Decimal,
	leverage int,
	side string,
	maintenanceRate decimal.Decimal,
) decimal.Decimal {
	leverageDec := decimal.NewFromInt(int64(leverage))

	if side == "long" {
		// Long liquidation: entryPrice * (1 - 1/leverage + maintenanceRate)
		return entryPrice.Mul(
			decimal.NewFromInt(1).Sub(decimal.NewFromInt(1).Div(leverageDec)).Add(maintenanceRate),
		)
	} else {
		// Short liquidation: entryPrice * (1 + 1/leverage - maintenanceRate)
		return entryPrice.Mul(
			decimal.NewFromInt(1).Add(decimal.NewFromInt(1).Div(leverageDec)).Sub(maintenanceRate),
		)
	}
}

// UpdatePosition updates position with current price and calculates PnL
func (s *MarginTradingService) UpdatePosition(ctx context.Context, positionID uint, currentPrice decimal.Decimal) error {
	var position MarginPosition
	if err := s.db.First(&position, positionID).Error; err != nil {
		return err
	}

	position.CurrentPrice = currentPrice

	// Calculate unrealized PnL
	if position.Side == "long" {
		position.UnrealizedPnL = currentPrice.Sub(position.EntryPrice).Mul(position.Quantity)
	} else {
		position.UnrealizedPnL = position.EntryPrice.Sub(currentPrice).Mul(position.Quantity)
	}

	position.UpdateTime = time.Now()

	// Check liquidation
	if s.shouldLiquidate(&position) {
		return s.LiquidatePosition(ctx, positionID)
	}

	// Check stop loss / take profit
	if position.StopLoss != nil && s.shouldTriggerStopLoss(&position) {
		return s.ClosePosition(ctx, positionID, currentPrice)
	}

	if position.TakeProfit != nil && s.shouldTriggerTakeProfit(&position) {
		return s.ClosePosition(ctx, positionID, currentPrice)
	}

	return s.db.Save(&position).Error
}

// shouldLiquidate checks if position should be liquidated
func (s *MarginTradingService) shouldLiquidate(position *MarginPosition) bool {
	if position.Side == "long" {
		return position.CurrentPrice.LessThanOrEqual(position.LiquidationPrice)
	} else {
		return position.CurrentPrice.GreaterThanOrEqual(position.LiquidationPrice)
	}
}

// shouldTriggerStopLoss checks if stop loss should trigger
func (s *MarginTradingService) shouldTriggerStopLoss(position *MarginPosition) bool {
	if position.StopLoss == nil {
		return false
	}

	if position.Side == "long" {
		return position.CurrentPrice.LessThanOrEqual(*position.StopLoss)
	} else {
		return position.CurrentPrice.GreaterThanOrEqual(*position.StopLoss)
	}
}

// shouldTriggerTakeProfit checks if take profit should trigger
func (s *MarginTradingService) shouldTriggerTakeProfit(position *MarginPosition) bool {
	if position.TakeProfit == nil {
		return false
	}

	if position.Side == "long" {
		return position.CurrentPrice.GreaterThanOrEqual(*position.TakeProfit)
	} else {
		return position.CurrentPrice.LessThanOrEqual(*position.TakeProfit)
	}
}

// ClosePosition closes a margin position
func (s *MarginTradingService) ClosePosition(ctx context.Context, positionID uint, closePrice decimal.Decimal) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var position MarginPosition
	if err := s.db.First(&position, positionID).Error; err != nil {
		return err
	}

	if position.Status != "open" {
		return errors.New("position is not open")
	}

	// Calculate realized PnL
	if position.Side == "long" {
		position.RealizedPnL = closePrice.Sub(position.EntryPrice).Mul(position.Quantity)
	} else {
		position.RealizedPnL = position.EntryPrice.Sub(closePrice).Mul(position.Quantity)
	}

	position.Status = "closed"
	now := time.Now()
	position.CloseTime = &now
	position.UpdateTime = now

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&position).Error; err != nil {
			return err
		}

		// Update margin account
		var account MarginAccount
		if err := tx.Where("user_id = ?", position.UserID).First(&account).Error; err != nil {
			return err
		}

		// Return margin and add/subtract PnL
		account.Collateral = account.Collateral.Add(position.Margin).Add(position.RealizedPnL)
		account.Equity = account.Collateral.Sub(account.Borrowed).Sub(account.Interest)
		account.UpdateTime = time.Now()

		if account.Borrowed.GreaterThan(decimal.Zero) {
			account.MarginLevel = account.Equity.Div(account.Borrowed)
		}

		return tx.Save(&account).Error
	})
}

// LiquidatePosition liquidates a position
func (s *MarginTradingService) LiquidatePosition(ctx context.Context, positionID uint) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var position MarginPosition
	if err := s.db.First(&position, positionID).Error; err != nil {
		return err
	}

	position.Status = "liquidated"
	position.RealizedPnL = position.LiquidationPrice.Sub(position.EntryPrice).Mul(position.Quantity)
	if position.Side == "short" {
		position.RealizedPnL = position.RealizedPnL.Neg()
	}

	now := time.Now()
	position.CloseTime = &now
	position.UpdateTime = now

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&position).Error; err != nil {
			return err
		}

		// Update margin account
		var account MarginAccount
		if err := tx.Where("user_id = ?", position.UserID).First(&account).Error; err != nil {
			return err
		}

		// Margin is lost in liquidation
		account.Collateral = account.Collateral.Sub(position.Margin)
		account.Equity = account.Collateral.Sub(account.Borrowed).Sub(account.Interest)
		account.UpdateTime = time.Now()

		if account.Borrowed.GreaterThan(decimal.Zero) {
			account.MarginLevel = account.Equity.Div(account.Borrowed)
		}

		// Check if account should be liquidated
		if account.MarginLevel.LessThan(account.MaintenanceRate) {
			account.Status = "liquidating"
		}

		return tx.Save(&account).Error
	})
}

// GetUserPositions gets all positions for a user
func (s *MarginTradingService) GetUserPositions(ctx context.Context, userID uint, status string) ([]MarginPosition, error) {
	var positions []MarginPosition
	query := s.db.Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("open_time DESC").Find(&positions).Error; err != nil {
		return nil, err
	}

	return positions, nil
}

// GetUserLoans gets all loans for a user
func (s *MarginTradingService) GetUserLoans(ctx context.Context, userID uint, status string) ([]MarginLoan, error) {
	var loans []MarginLoan
	query := s.db.Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("borrow_time DESC").Find(&loans).Error; err != nil {
		return nil, err
	}

	return loans, nil
}
