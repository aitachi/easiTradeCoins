//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

package security

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/easitradecoins/backend/internal/database"
	"github.com/easitradecoins/backend/internal/models"
	"github.com/shopspring/decimal"
)

// RiskManager handles risk management and security
type RiskManager struct {
	maxOrderSize          decimal.Decimal
	dailyWithdrawalLimit  map[int]decimal.Decimal // KYC level -> limit
	apiRateLimit          int
	orderRateLimit        int
}

// NewRiskManager creates a new risk manager
func NewRiskManager() *RiskManager {
	return &RiskManager{
		maxOrderSize: decimal.NewFromInt(1000000),
		dailyWithdrawalLimit: map[int]decimal.Decimal{
			0: decimal.NewFromInt(1000),
			1: decimal.NewFromInt(10000),
			2: decimal.NewFromInt(100000),
		},
		apiRateLimit:   100,
		orderRateLimit: 10,
	}
}

// ValidateOrder validates order against risk rules
func (rm *RiskManager) ValidateOrder(ctx context.Context, order *models.Order, user *models.User) error {
	// Check order size
	orderValue := order.Quantity.Mul(order.Price)
	if orderValue.GreaterThan(rm.maxOrderSize) {
		return fmt.Errorf("order size exceeds maximum allowed: %s", rm.maxOrderSize.String())
	}

	// Check price deviation
	if err := rm.checkPriceDeviation(order); err != nil {
		return err
	}

	// Check order frequency
	if err := rm.checkOrderFrequency(ctx, user.ID); err != nil {
		return err
	}

	// Check user status
	if user.Status != 1 {
		return errors.New("user account is not active")
	}

	return nil
}

// checkPriceDeviation checks if order price deviates too much from market
func (rm *RiskManager) checkPriceDeviation(order *models.Order) error {
	if order.Type == models.OrderTypeMarket {
		return nil
	}

	// Get last trade price
	var lastTrade models.Trade
	if err := database.DB.Where("symbol = ?", order.Symbol).
		Order("trade_time DESC").
		First(&lastTrade).Error; err != nil {
		// No previous trades, allow any price
		return nil
	}

	// Check if price deviates more than 10% from last trade
	maxDeviation := decimal.NewFromFloat(0.1)
	priceDiff := order.Price.Sub(lastTrade.Price).Abs()
	deviationPercent := priceDiff.Div(lastTrade.Price)

	if deviationPercent.GreaterThan(maxDeviation) {
		return fmt.Errorf("price deviates more than 10%% from market price")
	}

	return nil
}

// checkOrderFrequency checks order frequency
func (rm *RiskManager) checkOrderFrequency(ctx context.Context, userID uint) error {
	// Check orders in last second
	oneSecondAgo := time.Now().Add(-time.Second)

	var count int64
	if err := database.DB.Model(&models.Order{}).
		Where("user_id = ? AND create_time > ?", userID, oneSecondAgo).
		Count(&count).Error; err != nil {
		return err
	}

	if count >= int64(rm.orderRateLimit) {
		return errors.New("order rate limit exceeded")
	}

	return nil
}

// ValidateWithdrawal validates withdrawal against risk rules
func (rm *RiskManager) ValidateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal, user *models.User) error {
	// Check KYC level
	if user.KYCLevel == 0 {
		return errors.New("KYC verification required for withdrawal")
	}

	// Check daily withdrawal limit
	limit, exists := rm.dailyWithdrawalLimit[user.KYCLevel]
	if !exists {
		limit = rm.dailyWithdrawalLimit[0]
	}

	// Calculate today's withdrawals
	today := time.Now().Truncate(24 * time.Hour)
	var totalWithdrawn decimal.Decimal

	var withdrawals []models.Withdrawal
	if err := database.DB.Where("user_id = ? AND create_time >= ? AND status IN ?",
		user.ID, today, []int{1, 2, 3}).Find(&withdrawals).Error; err != nil {
		return err
	}

	for _, w := range withdrawals {
		totalWithdrawn = totalWithdrawn.Add(w.Amount)
	}

	// Check if new withdrawal exceeds limit
	if totalWithdrawn.Add(withdrawal.Amount).GreaterThan(limit) {
		return fmt.Errorf("daily withdrawal limit exceeded: %s", limit.String())
	}

	// Check for suspicious patterns
	if err := rm.detectSuspiciousWithdrawal(ctx, withdrawal, user); err != nil {
		return err
	}

	return nil
}

// detectSuspiciousWithdrawal detects suspicious withdrawal patterns
func (rm *RiskManager) detectSuspiciousWithdrawal(ctx context.Context, withdrawal *models.Withdrawal, user *models.User) error {
	// Check if this is first time withdrawal to this address
	var count int64
	if err := database.DB.Model(&models.Withdrawal{}).
		Where("user_id = ? AND address = ? AND status = ?", user.ID, withdrawal.Address, 3).
		Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		// First time withdrawal to this address - require additional verification
		// In production, send email/SMS confirmation
		return errors.New("first time withdrawal to this address requires confirmation")
	}

	// Check for rapid deposit-withdrawal pattern (possible money laundering)
	oneHourAgo := time.Now().Add(-time.Hour)
	var recentDeposit models.Deposit
	if err := database.DB.Where("user_id = ? AND create_time > ? AND currency = ?",
		user.ID, oneHourAgo, withdrawal.Currency).
		First(&recentDeposit).Error; err == nil {
		// Found recent deposit, check if withdrawal is close to deposit amount
		diff := recentDeposit.Amount.Sub(withdrawal.Amount).Abs()
		if diff.LessThan(recentDeposit.Amount.Mul(decimal.NewFromFloat(0.1))) {
			// Withdrawal is within 10% of deposit - suspicious
			return errors.New("suspicious withdrawal pattern detected - manual review required")
		}
	}

	return nil
}

// DetectSelfTrading detects self-trading (wash trading)
func (rm *RiskManager) DetectSelfTrading(buyerID, sellerID uint) bool {
	return buyerID == sellerID
}

// DetectRelatedAccounts detects related accounts by IP
func (rm *RiskManager) DetectRelatedAccounts(ctx context.Context, userID uint) ([]uint, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	var relatedUsers []models.User
	if err := database.DB.Where("id != ? AND (register_ip = ? OR last_login_ip = ?)",
		userID, user.RegisterIP, user.LastLoginIP).
		Find(&relatedUsers).Error; err != nil {
		return nil, err
	}

	relatedIDs := make([]uint, len(relatedUsers))
	for i, u := range relatedUsers {
		relatedIDs[i] = u.ID
	}

	return relatedIDs, nil
}

// CalculateRiskScore calculates risk score for a user
func (rm *RiskManager) CalculateRiskScore(ctx context.Context, userID uint) (float64, error) {
	var score float64

	// Factor 1: Trading frequency (20%)
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	var orderCount int64
	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND create_time > ?", userID, thirtyDaysAgo).
		Count(&orderCount)

	if orderCount > 1000 {
		score += 20
	} else if orderCount > 100 {
		score += 10
	} else {
		score += 5
	}

	// Factor 2: Large transactions (30%)
	var orders []models.Order
	database.DB.Where("user_id = ? AND create_time > ?", userID, thirtyDaysAgo).
		Order("filled_amount DESC").
		Limit(10).
		Find(&orders)

	var totalAmount decimal.Decimal
	for _, order := range orders {
		totalAmount = totalAmount.Add(order.FilledAmount)
	}

	avgAmount, _ := totalAmount.Div(decimal.NewFromInt(int64(len(orders)))).Float64()
	if avgAmount > 100000 {
		score += 30
	} else if avgAmount > 10000 {
		score += 20
	} else {
		score += 10
	}

	// Factor 3: Related accounts (20%)
	relatedAccounts, _ := rm.DetectRelatedAccounts(ctx, userID)
	if len(relatedAccounts) > 5 {
		score += 20
	} else if len(relatedAccounts) > 0 {
		score += 10
	}

	// Factor 4: Geographic risk (15%)
	// TODO: Implement based on user location

	// Factor 5: Historical violations (15%)
	// TODO: Implement based on violation records

	return score, nil
}

// FreezeAccount freezes a user account
func (rm *RiskManager) FreezeAccount(ctx context.Context, userID uint, reason string) error {
	return database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("status", 2).Error
}

// UnfreezeAccount unfreezes a user account
func (rm *RiskManager) UnfreezeAccount(ctx context.Context, userID uint) error {
	return database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("status", 1).Error
}
