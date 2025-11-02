package services

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a test database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate core models
	err = db.AutoMigrate(
		&MarginAccount{}, &MarginPosition{}, &MarginLoan{},
		&OptionContract{}, &OptionPosition{},
		&Trader{}, &FollowRelation{}, &CopiedOrder{},
		&TradingCommunity{}, &CommunityMember{}, &Post{}, &Comment{}, &Like{},
		&GridStrategy{}, &GridLevel{},
		&DCAStrategy{},
	)
	require.NoError(t, err)

	return db
}

// TestMarginTradingService tests margin trading functionality
func TestMarginTradingService(t *testing.T) {
	db := setupTestDB(t)
	orderService := NewOrderService(nil, nil, nil)
	service := NewMarginTradingService(orderService, db)
	ctx := context.Background()

	t.Run("CreateMarginAccount", func(t *testing.T) {
		account, err := service.GetOrCreateMarginAccount(ctx, 1)
		require.NoError(t, err)
		assert.NotNil(t, account)
		assert.Equal(t, uint(1), account.UserID)
	})

	t.Run("DepositCollateral", func(t *testing.T) {
		amount := decimal.NewFromInt(10000)
		err := service.Deposit(ctx, 1, amount)
		require.NoError(t, err)

		account, err := service.GetOrCreateMarginAccount(ctx, 1)
		require.NoError(t, err)
		assert.True(t, account.Collateral.GreaterThanOrEqual(amount))
	})
}

// TestOptionsTradingService tests options trading functionality
func TestOptionsTradingService(t *testing.T) {
	db := setupTestDB(t)
	service := NewOptionsTradingService(db)
	ctx := context.Background()

	t.Run("CreateCallOption", func(t *testing.T) {
		contract, err := service.CreateOptionContract(ctx, "BTC_USDT", "call",
			decimal.NewFromInt(55000), decimal.NewFromInt(500),
			decimal.NewFromInt(10), time.Now().Add(30*24*time.Hour),
			decimal.NewFromInt(50000))
		require.NoError(t, err)
		assert.Equal(t, "call", contract.Type)
		assert.Equal(t, "active", contract.Status)
	})

	t.Run("BuyCallOption", func(t *testing.T) {
		contract, _ := service.CreateOptionContract(ctx, "BTC_USDT", "call",
			decimal.NewFromInt(55000), decimal.NewFromInt(500),
			decimal.NewFromInt(10), time.Now().Add(30*24*time.Hour),
			decimal.NewFromInt(50000))

		position, err := service.BuyOption(ctx, 1, contract.ID, decimal.NewFromInt(1))
		require.NoError(t, err)
		assert.Equal(t, "long", position.Position)
		assert.Equal(t, "open", position.Status)
	})
}

// TestCopyTradingService tests copy trading functionality
func TestCopyTradingService(t *testing.T) {
	db := setupTestDB(t)
	orderService := NewOrderService(nil, nil, nil)
	service := NewCopyTradingService(orderService, db)
	ctx := context.Background()

	t.Run("RegisterTrader", func(t *testing.T) {
		trader, err := service.RegisterTrader(ctx, 1, "ProTrader123",
			"Experienced trader",
			decimal.NewFromInt(100),
			decimal.NewFromFloat(0.1))
		require.NoError(t, err)
		assert.Equal(t, "ProTrader123", trader.Username)
		assert.True(t, trader.IsActive)
	})

	t.Run("FollowTrader", func(t *testing.T) {
		trader, _ := service.RegisterTrader(ctx, 1, "ProTrader123", "",
			decimal.NewFromInt(100), decimal.NewFromFloat(0.1))

		relation, err := service.FollowTrader(ctx, 2, trader.ID,
			decimal.NewFromFloat(0.5), decimal.NewFromInt(1000), nil, nil)
		require.NoError(t, err)
		assert.Equal(t, uint(2), relation.FollowerID)
		assert.Equal(t, trader.ID, relation.TraderID)
		assert.True(t, relation.IsActive)
	})
}

// TestCommunityService tests trading community functionality
func TestCommunityService(t *testing.T) {
	db := setupTestDB(t)
	service := NewCommunityService(db)
	ctx := context.Background()

	t.Run("CreateCommunity", func(t *testing.T) {
		community, err := service.CreateCommunity(ctx, 1, "Crypto Traders",
			"A community for crypto enthusiasts", "general", false)
		require.NoError(t, err)
		assert.Equal(t, "Crypto Traders", community.Name)
		// TradingCommunity model doesn't have IsActive field, skip that check
	})

	t.Run("CreatePost", func(t *testing.T) {
		community, _ := service.CreateCommunity(ctx, 1, "Test Community",
			"Test", "general", false)

		post, err := service.CreatePost(ctx, 1, community.ID, "Test Post",
			"This is a test post", "general", nil)
		require.NoError(t, err)
		assert.Equal(t, "Test Post", post.Title)
		assert.Equal(t, community.ID, post.CommunityID)
	})
}

// TestServiceCompilation ensures all services can be instantiated
func TestServiceCompilation(t *testing.T) {
	db := setupTestDB(t)
	orderService := NewOrderService(nil, nil, nil)

	t.Run("InstantiateServices", func(t *testing.T) {
		// Test that all services can be created
		marginService := NewMarginTradingService(orderService, db)
		assert.NotNil(t, marginService)

		optionsService := NewOptionsTradingService(db)
		assert.NotNil(t, optionsService)

		copyTradingService := NewCopyTradingService(orderService, db)
		assert.NotNil(t, copyTradingService)

		communityService := NewCommunityService(db)
		assert.NotNil(t, communityService)

		gridService := NewGridTradingService(orderService, db)
		assert.NotNil(t, gridService)

		dcaService := NewDCAService(orderService, db)
		assert.NotNil(t, dcaService)
	})
}
