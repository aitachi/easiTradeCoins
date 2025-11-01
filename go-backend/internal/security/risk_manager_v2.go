package security

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/easitradecoins/backend/internal/database"
	"github.com/easitradecoins/backend/internal/models"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

// RiskConfig holds risk management configuration
type RiskConfig struct {
	MaxOrderSize             decimal.Decimal
	DailyWithdrawalLimit     map[int]decimal.Decimal
	APIRateLimit             int
	OrderRateLimit           int
	PriceDeviationLimit      decimal.Decimal
	OrderRateWindow          time.Duration
	LargeWithdrawalThreshold decimal.Decimal
	QuickInOutWindow         time.Duration
	HighRiskScoreThreshold   float64
	AutoFreezeEnabled        bool
}

// RiskManager handles comprehensive risk management
type RiskManager struct {
	config    *RiskConfig
	redis     *redis.Client
	mutex     sync.RWMutex
	whitelist map[string]bool // address whitelist
}

// NewRiskManager creates a new enhanced risk manager
func NewRiskManager(redisClient *redis.Client) *RiskManager {
	config := &RiskConfig{
		MaxOrderSize: decimal.NewFromInt(1000000),
		DailyWithdrawalLimit: map[int]decimal.Decimal{
			0: decimal.NewFromInt(1000),   // KYC Level 0
			1: decimal.NewFromInt(10000),  // KYC Level 1
			2: decimal.NewFromInt(100000), // KYC Level 2
		},
		APIRateLimit:             100,
		OrderRateLimit:           10,
		PriceDeviationLimit:      decimal.NewFromFloat(0.1), // 10%
		OrderRateWindow:          10 * time.Second,
		LargeWithdrawalThreshold: decimal.NewFromInt(10000),
		QuickInOutWindow:         24 * time.Hour,
		HighRiskScoreThreshold:   80.0,
		AutoFreezeEnabled:        true,
	}

	return &RiskManager{
		config:    config,
		redis:     redisClient,
		whitelist: make(map[string]bool),
	}
}

// ValidateOrder validates order with comprehensive risk checks
func (rm *RiskManager) ValidateOrder(ctx context.Context, order *models.Order, user *models.User) error {
	// 1. User status check
	if user.Status != 1 {
		rm.logRiskEvent(ctx, user.ID, "order_validation", "high",
			"Blocked order from frozen/inactive account", "", "blocked")
		return errors.New("user account is not active")
	}

	// 2. Distributed rate limiting check
	if err := rm.checkOrderFrequencyDistributed(ctx, user.ID); err != nil {
		rm.logRiskEvent(ctx, user.ID, "rate_limit_exceeded", "medium",
			"Order rate limit exceeded", fmt.Sprintf("limit: %d orders per %s", rm.config.OrderRateLimit, rm.config.OrderRateWindow), "blocked")
		return err
	}

	// 3. Order size check
	orderValue := order.Quantity.Mul(order.Price)
	if orderValue.GreaterThan(rm.config.MaxOrderSize) {
		rm.logRiskEvent(ctx, user.ID, "order_size_exceeded", "high",
			"Order size exceeds maximum", fmt.Sprintf("order value: %s, max: %s", orderValue, rm.config.MaxOrderSize), "blocked")
		return fmt.Errorf("order size exceeds maximum allowed: %s", rm.config.MaxOrderSize.String())
	}

	// 4. Price deviation check with cache
	if err := rm.checkPriceDeviationCached(ctx, order); err != nil {
		rm.logRiskEvent(ctx, user.ID, "price_deviation", "medium",
			"Order price deviates significantly from market", err.Error(), "blocked")
		return err
	}

	// 5. Check for abnormal trading patterns
	if err := rm.detectAbnormalTradingPattern(ctx, user.ID, order); err != nil {
		rm.logRiskEvent(ctx, user.ID, "abnormal_pattern", "high",
			"Abnormal trading pattern detected", err.Error(), "flagged")
		// Don't block, but flag for review
	}

	rm.logRiskEvent(ctx, user.ID, "order_validation", "low",
		"Order passed all risk checks", fmt.Sprintf("symbol: %s, size: %s", order.Symbol, orderValue), "allowed")

	return nil
}

// checkOrderFrequencyDistributed uses Redis sliding window for distributed rate limiting
func (rm *RiskManager) checkOrderFrequencyDistributed(ctx context.Context, userID uint) error {
	key := fmt.Sprintf("order_rate_limit:%d", userID)
	now := time.Now()
	windowStart := now.Add(-rm.config.OrderRateWindow)

	// Remove old entries outside the window
	rm.redis.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.Unix()))

	// Count orders in current window
	count, err := rm.redis.ZCount(ctx, key,
		fmt.Sprintf("%d", windowStart.Unix()),
		fmt.Sprintf("%d", now.Unix())).Result()
	if err != nil {
		return err
	}

	if count >= int64(rm.config.OrderRateLimit) {
		return fmt.Errorf("order rate limit exceeded: %d orders in %s", count, rm.config.OrderRateWindow)
	}

	// Add current order timestamp
	rm.redis.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.Unix()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	rm.redis.Expire(ctx, key, rm.config.OrderRateWindow+time.Minute)

	return nil
}

// checkPriceDeviationCached checks price deviation using Redis cache
func (rm *RiskManager) checkPriceDeviationCached(ctx context.Context, order *models.Order) error {
	if order.Type == models.OrderTypeMarket {
		return nil
	}

	// Try to get cached price first
	cacheKey := fmt.Sprintf("last_price:%s", order.Symbol)
	cachedPrice, err := rm.redis.Get(ctx, cacheKey).Result()

	var lastPrice decimal.Decimal

	if err == redis.Nil {
		// Cache miss - query database
		var lastTrade models.Trade
		if err := database.DB.Where("symbol = ?", order.Symbol).
			Order("trade_time DESC").
			First(&lastTrade).Error; err != nil {
			// No previous trades, allow any price
			return nil
		}
		lastPrice = lastTrade.Price

		// Update cache (TTL 5 seconds)
		rm.redis.Set(ctx, cacheKey, lastPrice.String(), 5*time.Second)
	} else if err != nil {
		// Redis error, fallback to database
		var lastTrade models.Trade
		if err := database.DB.Where("symbol = ?", order.Symbol).
			Order("trade_time DESC").
			First(&lastTrade).Error; err != nil {
			return nil
		}
		lastPrice = lastTrade.Price
	} else {
		// Cache hit
		lastPrice, _ = decimal.NewFromString(cachedPrice)
	}

	// Calculate deviation
	priceDiff := order.Price.Sub(lastPrice).Abs()
	deviationPercent := priceDiff.Div(lastPrice)

	// Get symbol-specific deviation limit (can be configured per trading pair)
	maxDeviation := rm.getMaxDeviationForSymbol(order.Symbol)

	if deviationPercent.GreaterThan(maxDeviation) {
		return fmt.Errorf("price deviates %.2f%% from market price (max: %.2f%%)",
			deviationPercent.Mul(decimal.NewFromInt(100)).InexactFloat64(),
			maxDeviation.Mul(decimal.NewFromInt(100)).InexactFloat64())
	}

	return nil
}

// getMaxDeviationForSymbol returns max price deviation for a symbol
func (rm *RiskManager) getMaxDeviationForSymbol(symbol string) decimal.Decimal {
	// TODO: Load from configuration per symbol
	// For now, return default
	return rm.config.PriceDeviationLimit
}

// detectAbnormalTradingPattern detects suspicious trading patterns
func (rm *RiskManager) detectAbnormalTradingPattern(ctx context.Context, userID uint, order *models.Order) error {
	// Check for rapid cancel pattern
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	var cancelCount int64
	database.DB.Model(&models.Order{}).
		Where("user_id = ? AND status = ? AND create_time > ?",
			userID, models.OrderStatusCancelled, fiveMinutesAgo).
		Count(&cancelCount)

	if cancelCount > 20 {
		return errors.New("excessive order cancellations detected")
	}

	// Check for fixed pattern trading (same amount/price repeatedly)
	var recentOrders []models.Order
	database.DB.Where("user_id = ? AND symbol = ? AND create_time > ?",
		userID, order.Symbol, fiveMinutesAgo).
		Order("create_time DESC").
		Limit(5).
		Find(&recentOrders)

	if len(recentOrders) >= 4 {
		// Check if all orders have similar quantity
		sameQuantity := true
		for i := 1; i < len(recentOrders); i++ {
			diff := recentOrders[i].Quantity.Sub(recentOrders[0].Quantity).Abs()
			if diff.GreaterThan(recentOrders[0].Quantity.Mul(decimal.NewFromFloat(0.01))) {
				sameQuantity = false
				break
			}
		}
		if sameQuantity {
			return errors.New("fixed pattern trading detected")
		}
	}

	return nil
}

// ValidateWithdrawal validates withdrawal with enhanced checks
func (rm *RiskManager) ValidateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal, user *models.User) error {
	// 1. KYC level check
	if user.KYCLevel == 0 {
		rm.logRiskEvent(ctx, user.ID, "withdrawal_validation", "high",
			"Withdrawal blocked: KYC required", "", "blocked")
		return errors.New("KYC verification required for withdrawal")
	}

	// 2. User status check
	if user.Status != 1 {
		return errors.New("user account is not active")
	}

	// 3. Check whitelist first
	if rm.isWhitelistedAddress(ctx, user.ID, withdrawal.Address, withdrawal.Currency) {
		rm.logRiskEvent(ctx, user.ID, "withdrawal_validation", "low",
			"Withdrawal to whitelisted address", withdrawal.Address, "allowed")
	} else {
		// 4. First-time address check (corrected logic)
		if err := rm.validateFirstTimeAddress(ctx, withdrawal, user); err != nil {
			return err
		}

		// 5. High-risk address check
		if rm.isHighRiskAddress(ctx, withdrawal.Address) {
			rm.logRiskEvent(ctx, user.ID, "high_risk_address", "critical",
				"Withdrawal to high-risk address blocked", withdrawal.Address, "blocked")
			return errors.New("withdrawal address is flagged as high risk")
		}
	}

	// 6. Daily withdrawal limit check
	if err := rm.checkDailyWithdrawalLimit(ctx, user, withdrawal); err != nil {
		return err
	}

	// 7. Large withdrawal check - require manual approval
	if withdrawal.Amount.GreaterThan(rm.config.LargeWithdrawalThreshold) {
		rm.logRiskEvent(ctx, user.ID, "large_withdrawal", "high",
			"Large withdrawal requires manual approval",
			fmt.Sprintf("amount: %s", withdrawal.Amount), "flagged")
		withdrawal.Status = 0 // Set to pending manual review
		return nil            // Don't block, but require approval
	}

	// 8. Quick in-out detection (improved)
	if err := rm.detectQuickInOut(ctx, withdrawal, user); err != nil {
		return err
	}

	// 9. Withdrawal frequency check
	if err := rm.checkWithdrawalFrequency(ctx, user.ID); err != nil {
		return err
	}

	rm.logRiskEvent(ctx, user.ID, "withdrawal_validation", "low",
		"Withdrawal passed all risk checks",
		fmt.Sprintf("amount: %s, address: %s", withdrawal.Amount, withdrawal.Address), "allowed")

	return nil
}

// validateFirstTimeAddress validates first-time withdrawal address
func (rm *RiskManager) validateFirstTimeAddress(ctx context.Context, withdrawal *models.Withdrawal, user *models.User) error {
	// Check all withdrawal records to this address (not just completed ones)
	var count int64
	if err := database.DB.Model(&models.Withdrawal{}).
		Where("user_id = ? AND address = ? AND currency = ?",
			user.ID, withdrawal.Address, withdrawal.Currency).
		Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		rm.logRiskEvent(ctx, user.ID, "first_time_address", "medium",
			"First-time withdrawal address detected",
			withdrawal.Address, "flagged")

		// In production: send SMS/email verification code
		// withdrawal.Status = models.WithdrawalStatusPendingVerification
		// return errors.New("first time withdrawal to this address requires verification")

		// For now, just flag but allow
		return nil
	}

	return nil
}

// isWhitelistedAddress checks if address is in whitelist
func (rm *RiskManager) isWhitelistedAddress(ctx context.Context, userID uint, address, currency string) bool {
	var whitelist models.WithdrawalWhitelist
	err := database.DB.Where("user_id = ? AND address = ? AND currency = ? AND is_active = ?",
		userID, address, currency, true).First(&whitelist).Error
	return err == nil
}

// isHighRiskAddress checks if address is flagged as high risk
func (rm *RiskManager) isHighRiskAddress(ctx context.Context, address string) bool {
	// TODO: Integrate with external risk scoring service
	// Check against known mixer addresses, sanctioned addresses, etc.

	// For now, check local blacklist
	key := fmt.Sprintf("blacklist:address:%s", address)
	exists, _ := rm.redis.Exists(ctx, key).Result()
	return exists > 0
}

// checkDailyWithdrawalLimit checks if user exceeds daily limit
func (rm *RiskManager) checkDailyWithdrawalLimit(ctx context.Context, user *models.User, withdrawal *models.Withdrawal) error {
	limit, exists := rm.config.DailyWithdrawalLimit[user.KYCLevel]
	if !exists {
		limit = rm.config.DailyWithdrawalLimit[0]
	}

	// Calculate today's withdrawals (use UTC to avoid timezone issues)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	var totalWithdrawn decimal.Decimal

	var withdrawals []models.Withdrawal
	if err := database.DB.Where("user_id = ? AND create_time >= ? AND status IN ?",
		user.ID, today, []int{1, 2, 3}).Find(&withdrawals).Error; err != nil {
		return err
	}

	for _, w := range withdrawals {
		totalWithdrawn = totalWithdrawn.Add(w.Amount)
	}

	if totalWithdrawn.Add(withdrawal.Amount).GreaterThan(limit) {
		rm.logRiskEvent(ctx, user.ID, "daily_limit_exceeded", "medium",
			"Daily withdrawal limit exceeded",
			fmt.Sprintf("total: %s, limit: %s", totalWithdrawn, limit), "blocked")
		return fmt.Errorf("daily withdrawal limit exceeded: %s / %s", totalWithdrawn.String(), limit.String())
	}

	return nil
}

// detectQuickInOut detects rapid deposit-withdrawal pattern (improved)
func (rm *RiskManager) detectQuickInOut(ctx context.Context, withdrawal *models.Withdrawal, user *models.User) error {
	windowStart := time.Now().Add(-rm.config.QuickInOutWindow)

	var recentDeposits []models.Deposit
	if err := database.DB.Where("user_id = ? AND create_time > ? AND currency = ? AND status = ?",
		user.ID, windowStart, withdrawal.Currency, 1). // status 1 = completed
		Find(&recentDeposits).Error; err != nil {
		return err
	}

	if len(recentDeposits) == 0 {
		return nil
	}

	var totalDeposited decimal.Decimal
	for _, deposit := range recentDeposits {
		totalDeposited = totalDeposited.Add(deposit.Amount)
	}

	// If withdrawal is > 80% of recent deposits, flag as suspicious
	threshold := totalDeposited.Mul(decimal.NewFromFloat(0.8))
	if withdrawal.Amount.GreaterThan(threshold) {
		rm.logRiskEvent(ctx, user.ID, "quick_in_out", "high",
			"Quick in-out pattern detected",
			fmt.Sprintf("deposited: %s, withdrawing: %s within %s",
				totalDeposited, withdrawal.Amount, rm.config.QuickInOutWindow), "flagged")

		// Don't block immediately, but flag for manual review
		withdrawal.Status = 0 // Pending review
		return nil
	}

	return nil
}

// checkWithdrawalFrequency prevents excessive small withdrawals
func (rm *RiskManager) checkWithdrawalFrequency(ctx context.Context, userID uint) error {
	oneHourAgo := time.Now().Add(-time.Hour)
	var count int64
	database.DB.Model(&models.Withdrawal{}).
		Where("user_id = ? AND create_time > ?", userID, oneHourAgo).
		Count(&count)

	if count >= 10 {
		rm.logRiskEvent(ctx, userID, "withdrawal_frequency", "medium",
			"Too many withdrawal requests", fmt.Sprintf("%d withdrawals in 1 hour", count), "blocked")
		return errors.New("too many withdrawal requests, please try again later")
	}

	return nil
}

// DetectSelfTrading detects self-trading with enhanced checks
func (rm *RiskManager) DetectSelfTrading(ctx context.Context, trade *models.Trade) (bool, string, error) {
	// 1. Direct self-trading
	if trade.BuyerID == trade.SellerID {
		rm.logViolation(ctx, trade.BuyerID, "self_trading", "direct", 8)
		return true, "direct_self_trading", nil
	}

	// 2. Related account trading
	relatedBuyer, _ := rm.DetectRelatedAccounts(ctx, trade.BuyerID)
	relatedSeller, _ := rm.DetectRelatedAccounts(ctx, trade.SellerID)

	if hasCommonAccount(relatedBuyer, relatedSeller) || contains(relatedBuyer, trade.SellerID) || contains(relatedSeller, trade.BuyerID) {
		rm.logViolation(ctx, trade.BuyerID, "self_trading", "related_accounts", 7)
		rm.logViolation(ctx, trade.SellerID, "self_trading", "related_accounts", 7)
		return true, "related_account_trading", nil
	}

	// 3. Rapid trading pattern (wash trading)
	recentTrades := rm.getRecentTradesBetween(ctx, trade.BuyerID, trade.SellerID, 5*time.Minute)
	if len(recentTrades) > 5 {
		rm.logViolation(ctx, trade.BuyerID, "wash_trading", "rapid_pattern", 6)
		return true, "rapid_trading_pattern", nil
	}

	// 4. Price abnormality check
	if rm.isTradePriceAbnormal(ctx, trade) {
		rm.logViolation(ctx, trade.BuyerID, "suspicious_trade", "abnormal_price", 5)
		return true, "abnormal_price", nil
	}

	return false, "", nil
}

// DetectRelatedAccounts detects related accounts based on IP, device fingerprint, etc.
func (rm *RiskManager) DetectRelatedAccounts(ctx context.Context, userID uint) ([]uint, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	var relatedUsers []models.User
	// Find users with same IP
	database.DB.Where("(register_ip = ? OR last_login_ip = ?) AND id != ?",
		user.RegisterIP, user.LastLoginIP, userID).
		Find(&relatedUsers)

	relatedIDs := make([]uint, len(relatedUsers))
	for i, u := range relatedUsers {
		relatedIDs[i] = u.ID
	}

	return relatedIDs, nil
}

// getRecentTradesBetween gets recent trades between two users
func (rm *RiskManager) getRecentTradesBetween(ctx context.Context, user1ID, user2ID uint, window time.Duration) []models.Trade {
	var trades []models.Trade
	windowStart := time.Now().Add(-window)

	database.DB.Where("((buyer_id = ? AND seller_id = ?) OR (buyer_id = ? AND seller_id = ?)) AND trade_time > ?",
		user1ID, user2ID, user2ID, user1ID, windowStart).
		Find(&trades)

	return trades
}

// isTradePriceAbnormal checks if trade price is abnormal
func (rm *RiskManager) isTradePriceAbnormal(ctx context.Context, trade *models.Trade) bool {
	// Get recent trades for this symbol
	var recentTrades []models.Trade
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	database.DB.Where("symbol = ? AND trade_time > ? AND id != ?",
		trade.Symbol, fiveMinutesAgo, trade.ID).
		Order("trade_time DESC").
		Limit(10).
		Find(&recentTrades)

	if len(recentTrades) == 0 {
		return false // No comparison data
	}

	// Calculate average price
	var total decimal.Decimal
	for _, t := range recentTrades {
		total = total.Add(t.Price)
	}
	avgPrice := total.Div(decimal.NewFromInt(int64(len(recentTrades))))

	// Check if current trade price deviates more than 20% from average
	deviation := trade.Price.Sub(avgPrice).Abs().Div(avgPrice)
	return deviation.GreaterThan(decimal.NewFromFloat(0.2))
}

// logRiskEvent logs a risk event to database
func (rm *RiskManager) logRiskEvent(ctx context.Context, userID uint, eventType, severity, description, details, action string) {
	event := models.RiskEvent{
		UserID:      userID,
		EventType:   eventType,
		Severity:    severity,
		Description: description,
		Details:     details,
		Action:      action,
		CreateTime:  time.Now(),
	}
	database.DB.Create(&event)
}

// logViolation logs a violation to database
func (rm *RiskManager) logViolation(ctx context.Context, userID uint, violationType, description string, severity int) {
	violation := models.Violation{
		UserID:      userID,
		Type:        violationType,
		Status:      "active",
		Severity:    severity,
		Description: description,
		CreateTime:  time.Now(),
	}
	database.DB.Create(&violation)
}

// Helper functions
func hasCommonAccount(list1, list2 []uint) bool {
	for _, id1 := range list1 {
		for _, id2 := range list2 {
			if id1 == id2 {
				return true
			}
		}
	}
	return false
}

func contains(list []uint, target uint) bool {
	for _, id := range list {
		if id == target {
			return true
		}
	}
	return false
}
