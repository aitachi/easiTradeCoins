package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/easitradecoins/backend/internal/database"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OptionContract represents an option contract
// 期权合约
type OptionContract struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	Symbol          string          `json:"symbol" gorm:"index"`                   // 标的资产
	Type            string          `json:"type"`                                  // call/put
	StrikePrice     decimal.Decimal `json:"strike_price" gorm:"type:decimal(36,18)"` // 行权价
	Premium         decimal.Decimal `json:"premium" gorm:"type:decimal(36,18)"`      // 权利金
	Quantity        decimal.Decimal `json:"quantity" gorm:"type:decimal(36,18)"`     // 合约数量
	Expiry          time.Time       `json:"expiry" gorm:"index"`                     // 到期时间
	UnderlyingPrice decimal.Decimal `json:"underlying_price" gorm:"type:decimal(36,18)"` // 标的价格
	ImpliedVolatility decimal.Decimal `json:"implied_volatility" gorm:"type:decimal(10,4)"` // 隐含波动率
	Status          string          `json:"status" gorm:"default:active"` // active/expired/exercised
	CreateTime      time.Time       `json:"create_time"`
	UpdateTime      time.Time       `json:"update_time"`
}

// OptionPosition represents a user's option position
// 期权持仓
type OptionPosition struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	UserID          uint            `json:"user_id" gorm:"index"`
	ContractID      uint            `json:"contract_id" gorm:"index"`
	Position        string          `json:"position"`                          // long/short
	Quantity        decimal.Decimal `json:"quantity" gorm:"type:decimal(36,18)"`
	EntryPremium    decimal.Decimal `json:"entry_premium" gorm:"type:decimal(36,18)"` // 开仓权利金
	CurrentPremium  decimal.Decimal `json:"current_premium" gorm:"type:decimal(36,18)"` // 当前权利金
	UnrealizedPnL   decimal.Decimal `json:"unrealized_pnl" gorm:"type:decimal(36,18)"`
	RealizedPnL     decimal.Decimal `json:"realized_pnl" gorm:"type:decimal(36,18);default:0"`
	Status          string          `json:"status" gorm:"default:open"` // open/closed/exercised/expired
	OpenTime        time.Time       `json:"open_time"`
	CloseTime       *time.Time      `json:"close_time,omitempty"`
	UpdateTime      time.Time       `json:"update_time"`
}

// OptionsTradingService manages options trading
type OptionsTradingService struct {
	mutex sync.RWMutex
	db    *gorm.DB
}

// NewOptionsTradingService creates a new options trading service
func NewOptionsTradingService(db *gorm.DB) *OptionsTradingService {
	return &OptionsTradingService{
		db: db,
	}
}

// CreateOptionContract creates a new option contract
func (s *OptionsTradingService) CreateOptionContract(
	ctx context.Context,
	symbol, optionType string,
	strikePrice, premium decimal.Decimal,
	quantity decimal.Decimal,
	expiry time.Time,
	underlyingPrice decimal.Decimal,
) (*OptionContract, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if optionType != "call" && optionType != "put" {
		return nil, errors.New("option type must be 'call' or 'put'")
	}

	if expiry.Before(time.Now()) {
		return nil, errors.New("expiry time must be in the future")
	}

	contract := &OptionContract{
		Symbol:            symbol,
		Type:              optionType,
		StrikePrice:       strikePrice,
		Premium:           premium,
		Quantity:          quantity,
		Expiry:            expiry,
		UnderlyingPrice:   underlyingPrice,
		ImpliedVolatility: decimal.NewFromFloat(0.3), // Default 30%
		Status:            "active",
		CreateTime:        time.Now(),
		UpdateTime:        time.Now(),
	}

	if err := s.db.Create(contract).Error; err != nil {
		return nil, err
	}

	return contract, nil
}

// BuyOption buys an option (long position)
func (s *OptionsTradingService) BuyOption(
	ctx context.Context,
	userID, contractID uint,
	quantity decimal.Decimal,
) (*OptionPosition, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var contract OptionContract
	if err := s.db.First(&contract, contractID).Error; err != nil {
		return nil, errors.New("contract not found")
	}

	if contract.Status != "active" {
		return nil, errors.New("contract is not active")
	}

	if contract.Expiry.Before(time.Now()) {
		return nil, errors.New("contract has expired")
	}

	// Calculate total premium to pay
	totalPremium := contract.Premium.Mul(quantity)

	position := &OptionPosition{
		UserID:         userID,
		ContractID:     contractID,
		Position:       "long",
		Quantity:       quantity,
		EntryPremium:   contract.Premium,
		CurrentPremium: contract.Premium,
		UnrealizedPnL:  decimal.Zero,
		RealizedPnL:    decimal.Zero,
		Status:         "open",
		OpenTime:       time.Now(),
		UpdateTime:     time.Now(),
	}

	if err := s.db.Create(position).Error; err != nil {
		return nil, err
	}

	return position, nil
}

// SellOption sells an option (short position)
func (s *OptionsTradingService) SellOption(
	ctx context.Context,
	userID, contractID uint,
	quantity decimal.Decimal,
) (*OptionPosition, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var contract OptionContract
	if err := s.db.First(&contract, contractID).Error; err != nil {
		return nil, errors.New("contract not found")
	}

	if contract.Status != "active" {
		return nil, errors.New("contract is not active")
	}

	position := &OptionPosition{
		UserID:         userID,
		ContractID:     contractID,
		Position:       "short",
		Quantity:       quantity,
		EntryPremium:   contract.Premium,
		CurrentPremium: contract.Premium,
		UnrealizedPnL:  decimal.Zero,
		RealizedPnL:    decimal.Zero,
		Status:         "open",
		OpenTime:       time.Now(),
		UpdateTime:     time.Now(),
	}

	if err := s.db.Create(position).Error; err != nil {
		return nil, err
	}

	return position, nil
}

// ExerciseOption exercises an option position
func (s *OptionsTradingService) ExerciseOption(
	ctx context.Context,
	positionID uint,
	currentUnderlyingPrice decimal.Decimal,
) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var position OptionPosition
	if err := s.db.Preload("Contract").First(&position, positionID).Error; err != nil {
		return errors.New("position not found")
	}

	if position.Status != "open" {
		return errors.New("position is not open")
	}

	var contract OptionContract
	if err := s.db.First(&contract, position.ContractID).Error; err != nil {
		return err
	}

	if contract.Expiry.Before(time.Now()) {
		return errors.New("contract has expired")
	}

	// Calculate exercise value
	var exerciseValue decimal.Decimal

	if position.Position == "long" {
		if contract.Type == "call" {
			// Call option: max(S - K, 0)
			if currentUnderlyingPrice.GreaterThan(contract.StrikePrice) {
				exerciseValue = currentUnderlyingPrice.Sub(contract.StrikePrice).Mul(position.Quantity)
			}
		} else {
			// Put option: max(K - S, 0)
			if contract.StrikePrice.GreaterThan(currentUnderlyingPrice) {
				exerciseValue = contract.StrikePrice.Sub(currentUnderlyingPrice).Mul(position.Quantity)
			}
		}

		// Subtract premium paid
		position.RealizedPnL = exerciseValue.Sub(position.EntryPremium.Mul(position.Quantity))
	} else {
		// Short position
		if contract.Type == "call" {
			if currentUnderlyingPrice.GreaterThan(contract.StrikePrice) {
				exerciseValue = currentUnderlyingPrice.Sub(contract.StrikePrice).Mul(position.Quantity)
			}
		} else {
			if contract.StrikePrice.GreaterThan(currentUnderlyingPrice) {
				exerciseValue = contract.StrikePrice.Sub(currentUnderlyingPrice).Mul(position.Quantity)
			}
		}

		// Seller keeps premium but pays exercise value
		position.RealizedPnL = position.EntryPremium.Mul(position.Quantity).Sub(exerciseValue)
	}

	position.Status = "exercised"
	now := time.Now()
	position.CloseTime = &now
	position.UpdateTime = now

	return s.db.Save(&position).Error
}

// ClosePosition closes an option position before expiry
func (s *OptionsTradingService) ClosePosition(
	ctx context.Context,
	positionID uint,
	currentPremium decimal.Decimal,
) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var position OptionPosition
	if err := s.db.First(&position, positionID).Error; err != nil {
		return errors.New("position not found")
	}

	if position.Status != "open" {
		return errors.New("position is not open")
	}

	// Calculate P&L
	if position.Position == "long" {
		// Long: (current premium - entry premium) * quantity
		position.RealizedPnL = currentPremium.Sub(position.EntryPremium).Mul(position.Quantity)
	} else {
		// Short: (entry premium - current premium) * quantity
		position.RealizedPnL = position.EntryPremium.Sub(currentPremium).Mul(position.Quantity)
	}

	position.Status = "closed"
	now := time.Now()
	position.CloseTime = &now
	position.UpdateTime = now

	return s.db.Save(&position).Error
}

// UpdateContractPremium updates the option contract premium
func (s *OptionsTradingService) UpdateContractPremium(
	ctx context.Context,
	contractID uint,
	newPremium decimal.Decimal,
	underlyingPrice decimal.Decimal,
) error {
	var contract OptionContract
	if err := s.db.First(&contract, contractID).Error; err != nil {
		return err
	}

	contract.Premium = newPremium
	contract.UnderlyingPrice = underlyingPrice
	contract.UpdateTime = time.Now()

	// Update all open positions' unrealized P&L
	var positions []OptionPosition
	s.db.Where("contract_id = ? AND status = ?", contractID, "open").Find(&positions)

	for i := range positions {
		positions[i].CurrentPremium = newPremium

		if positions[i].Position == "long" {
			positions[i].UnrealizedPnL = newPremium.Sub(positions[i].EntryPremium).Mul(positions[i].Quantity)
		} else {
			positions[i].UnrealizedPnL = positions[i].EntryPremium.Sub(newPremium).Mul(positions[i].Quantity)
		}

		positions[i].UpdateTime = time.Now()
		s.db.Save(&positions[i])
	}

	return s.db.Save(&contract).Error
}

// ExpireContracts expires all contracts past their expiry time
func (s *OptionsTradingService) ExpireContracts(ctx context.Context) error {
	var contracts []OptionContract
	if err := s.db.Where("status = ? AND expiry < ?", "active", time.Now()).Find(&contracts).Error; err != nil {
		return err
	}

	for i := range contracts {
		contracts[i].Status = "expired"
		contracts[i].UpdateTime = time.Now()
		s.db.Save(&contracts[i])

		// Expire all open positions for this contract
		var positions []OptionPosition
		s.db.Where("contract_id = ? AND status = ?", contracts[i].ID, "open").Find(&positions)

		for j := range positions {
			// Options expire worthless if not exercised
			if positions[j].Position == "long" {
				// Long loses premium paid
				positions[j].RealizedPnL = positions[j].EntryPremium.Mul(positions[j].Quantity).Neg()
			} else {
				// Short keeps premium
				positions[j].RealizedPnL = positions[j].EntryPremium.Mul(positions[j].Quantity)
			}

			positions[j].Status = "expired"
			now := time.Now()
			positions[j].CloseTime = &now
			positions[j].UpdateTime = now
			s.db.Save(&positions[j])
		}
	}

	return nil
}

// GetActiveContracts gets all active option contracts
func (s *OptionsTradingService) GetActiveContracts(ctx context.Context, symbol string) ([]OptionContract, error) {
	var contracts []OptionContract
	query := s.db.Where("status = ? AND expiry > ?", "active", time.Now())

	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}

	if err := query.Order("expiry ASC").Find(&contracts).Error; err != nil {
		return nil, err
	}

	return contracts, nil
}

// GetUserPositions gets all option positions for a user
func (s *OptionsTradingService) GetUserPositions(ctx context.Context, userID uint, status string) ([]OptionPosition, error) {
	var positions []OptionPosition
	query := s.db.Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("open_time DESC").Find(&positions).Error; err != nil {
		return nil, err
	}

	return positions, nil
}

// CalculateOptionPrice calculates theoretical option price using Black-Scholes
func (s *OptionsTradingService) CalculateOptionPrice(
	spotPrice, strikePrice decimal.Decimal,
	timeToExpiry float64, // in years
	riskFreeRate, volatility float64,
	optionType string,
) decimal.Decimal {
	// Simplified Black-Scholes calculation
	// Note: This is a simplified version. Production should use proper math libraries

	// For now, return a basic intrinsic value calculation
	if optionType == "call" {
		intrinsic := spotPrice.Sub(strikePrice)
		if intrinsic.GreaterThan(decimal.Zero) {
			return intrinsic
		}
	} else {
		intrinsic := strikePrice.Sub(spotPrice)
		if intrinsic.GreaterThan(decimal.Zero) {
			return intrinsic
		}
	}

	return decimal.Zero
}
