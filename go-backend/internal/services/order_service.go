package services

import (
	"context"
	"errors"

	"github.com/easitradecoins/backend/internal/database"
	"github.com/easitradecoins/backend/internal/matching"
	"github.com/easitradecoins/backend/internal/models"
	"github.com/easitradecoins/backend/internal/security"
	"gorm.io/gorm"
)

// OrderService handles order-related operations
type OrderService struct {
	engine       *matching.MatchingEngine
	assetService *AssetService
	riskManager  *security.RiskManager
}

// NewOrderService creates a new order service
func NewOrderService(engine *matching.MatchingEngine, assetService *AssetService, riskManager *security.RiskManager) *OrderService {
	return &OrderService{
		engine:       engine,
		assetService: assetService,
		riskManager:  riskManager,
	}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(order *models.Order) (*models.Order, []*models.Trade, error) {
	var trades []*models.Trade

	// Get user for risk validation
	var user models.User
	if err := database.DB.First(&user, order.UserID).Error; err != nil {
		return nil, nil, errors.New("user not found")
	}

	// Risk validation - validate order before processing
	if s.riskManager != nil {
		ctx := context.Background()
		if err := s.riskManager.ValidateOrder(ctx, order, &user); err != nil {
			return nil, nil, err
		}
	}

	// Use database transaction for entire order creation process
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Validate user has sufficient balance
		if err := s.validateOrderBalance(order); err != nil {
			return err
		}

		// Freeze assets within transaction
		if err := s.freezeOrderAssetsWithTx(tx, order); err != nil {
			return err
		}

		// Process order in matching engine
		matchedTrades, err := s.engine.ProcessOrder(order)
		if err != nil {
			return err
		}
		trades = matchedTrades

		// Save order to database
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// Save trades to database and check for self-trading
		for _, trade := range trades {
			// Check for self-trading before saving
			if s.riskManager != nil {
				isSelfTrading, reason, _ := s.riskManager.DetectSelfTrading(context.Background(), trade)
				if isSelfTrading {
					return errors.New("self-trading detected: " + reason)
				}
			}

			if err := tx.Create(trade).Error; err != nil {
				return err // Fail entire transaction if trade save fails
			}

			// Update user assets based on trades within transaction
			if err := s.processTradeSettlementWithTx(tx, trade); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return order, trades, nil
}

// CancelOrder cancels an order
func (s *OrderService) CancelOrder(orderID string, userID uint) error {
	// Get order from database
	var order models.Order
	if err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return err
	}

	// Check if order can be cancelled
	if order.Status == models.OrderStatusFilled || order.Status == models.OrderStatusCancelled {
		return errors.New("order cannot be cancelled")
	}

	// Cancel in matching engine
	if err := s.engine.CancelOrder(order.Symbol, orderID); err != nil {
		return err
	}

	// Update order status in database
	order.Status = models.OrderStatusCancelled
	if err := database.DB.Save(&order).Error; err != nil {
		return err
	}

	// Unfreeze remaining assets
	s.unfreezeOrderAssets(&order)

	return nil
}

// GetOrder gets an order by ID
func (s *OrderService) GetOrder(orderID string, userID uint) (*models.Order, error) {
	var order models.Order
	if err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOpenOrders gets all open orders for a user
func (s *OrderService) GetOpenOrders(userID uint, symbol string) ([]models.Order, error) {
	query := database.DB.Where("user_id = ? AND status IN ?", userID, []models.OrderStatus{
		models.OrderStatusPending,
		models.OrderStatusPartial,
	})

	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}

	var orders []models.Order
	if err := query.Order("create_time DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

// GetOrderHistory gets order history for a user
func (s *OrderService) GetOrderHistory(userID uint, symbol string, limit, offset int) ([]models.Order, error) {
	query := database.DB.Where("user_id = ?", userID)

	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}

	var orders []models.Order
	if err := query.Order("create_time DESC").Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

// GetOrderBookDepth gets order book depth
func (s *OrderService) GetOrderBookDepth(symbol string, depth int) ([]matching.PriceLevelInfo, []matching.PriceLevelInfo) {
	ob, exists := s.engine.GetOrderBook(symbol)
	if !exists {
		return []matching.PriceLevelInfo{}, []matching.PriceLevelInfo{}
	}

	bids, asks := ob.GetDepth(depth)
	return bids, asks
}

// GetRecentTrades gets recent trades for a symbol
func (s *OrderService) GetRecentTrades(symbol string, limit int) ([]models.Trade, error) {
	var trades []models.Trade
	if err := database.DB.Where("symbol = ?", symbol).
		Order("trade_time DESC").
		Limit(limit).
		Find(&trades).Error; err != nil {
		return nil, err
	}

	return trades, nil
}

// validateOrderBalance validates user has sufficient balance
func (s *OrderService) validateOrderBalance(order *models.Order) error {
	// Get trading pair to determine currencies
	var pair models.TradingPair
	if err := database.DB.Where("symbol = ?", order.Symbol).First(&pair).Error; err != nil {
		return errors.New("invalid trading pair")
	}

	var currency string
	var requiredAmount = order.Quantity

	if order.Side == models.OrderSideBuy {
		// For buy orders, check quote currency (e.g., USDT in BTC_USDT)
		currency = pair.QuoteCurrency
		if order.Type == models.OrderTypeLimit {
			requiredAmount = order.Quantity.Mul(order.Price)
		}
	} else {
		// For sell orders, check base currency (e.g., BTC in BTC_USDT)
		currency = pair.BaseCurrency
	}

	// Get user asset
	asset, err := s.assetService.GetUserAsset(order.UserID, currency, "ERC20")
	if err != nil {
		return err
	}

	if asset.Available.LessThan(requiredAmount) {
		return errors.New("insufficient balance")
	}

	return nil
}

// freezeOrderAssets freezes assets for an order
func (s *OrderService) freezeOrderAssets(order *models.Order) error {
	var pair models.TradingPair
	if err := database.DB.Where("symbol = ?", order.Symbol).First(&pair).Error; err != nil {
		return err
	}

	var currency string
	var amount = order.Quantity

	if order.Side == models.OrderSideBuy {
		currency = pair.QuoteCurrency
		if order.Type == models.OrderTypeLimit {
			amount = order.Quantity.Mul(order.Price)
		}
	} else {
		currency = pair.BaseCurrency
	}

	return s.assetService.FreezeAsset(context.Background(), order.UserID, currency, "ERC20", amount)
}

// unfreezeOrderAssets unfreezes assets for a cancelled order
func (s *OrderService) unfreezeOrderAssets(order *models.Order) error {
	var pair models.TradingPair
	if err := database.DB.Where("symbol = ?", order.Symbol).First(&pair).Error; err != nil {
		return err
	}

	var currency string
	remainingQty := order.Quantity.Sub(order.FilledQty)
	var amount = remainingQty

	if order.Side == models.OrderSideBuy {
		currency = pair.QuoteCurrency
		if order.Type == models.OrderTypeLimit {
			amount = remainingQty.Mul(order.Price)
		}
	} else {
		currency = pair.BaseCurrency
	}

	return s.assetService.UnfreezeAsset(context.Background(), order.UserID, currency, "ERC20", amount)
}

// processTradeSettlement processes trade settlement
func (s *OrderService) processTradeSettlement(trade *models.Trade) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Get trading pair
		var pair models.TradingPair
		if err := tx.Where("symbol = ?", trade.Symbol).First(&pair).Error; err != nil {
			return err
		}

		// Update buyer assets (give base currency, take quote currency)
		var buyerBaseAsset models.UserAsset
		if err := tx.Where("user_id = ? AND currency = ? AND chain = ?",
			trade.BuyerID, pair.BaseCurrency, "ERC20").First(&buyerBaseAsset).Error; err != nil {
			return err
		}

		buyerBaseAsset.Available = buyerBaseAsset.Available.Add(trade.Quantity)
		if err := tx.Save(&buyerBaseAsset).Error; err != nil {
			return err
		}

		var buyerQuoteAsset models.UserAsset
		if err := tx.Where("user_id = ? AND currency = ? AND chain = ?",
			trade.BuyerID, pair.QuoteCurrency, "ERC20").First(&buyerQuoteAsset).Error; err != nil {
			return err
		}

		buyerQuoteAsset.Frozen = buyerQuoteAsset.Frozen.Sub(trade.Amount.Add(trade.BuyerFee))
		if err := tx.Save(&buyerQuoteAsset).Error; err != nil {
			return err
		}

		// Update seller assets (take base currency, give quote currency)
		var sellerBaseAsset models.UserAsset
		if err := tx.Where("user_id = ? AND currency = ? AND chain = ?",
			trade.SellerID, pair.BaseCurrency, "ERC20").First(&sellerBaseAsset).Error; err != nil {
			return err
		}

		sellerBaseAsset.Frozen = sellerBaseAsset.Frozen.Sub(trade.Quantity.Add(trade.SellerFee))
		if err := tx.Save(&sellerBaseAsset).Error; err != nil {
			return err
		}

		var sellerQuoteAsset models.UserAsset
		if err := tx.Where("user_id = ? AND currency = ? AND chain = ?",
			trade.SellerID, pair.QuoteCurrency, "ERC20").First(&sellerQuoteAsset).Error; err != nil {
			return err
		}

		sellerQuoteAsset.Available = sellerQuoteAsset.Available.Add(trade.Amount)
		if err := tx.Save(&sellerQuoteAsset).Error; err != nil {
			return err
		}

		return nil
	})
}

// freezeOrderAssetsWithTx freezes assets for an order within a transaction
func (s *OrderService) freezeOrderAssetsWithTx(tx *gorm.DB, order *models.Order) error {
	var pair models.TradingPair
	if err := tx.Where("symbol = ?", order.Symbol).First(&pair).Error; err != nil {
		return err
	}

	var currency string
	var amount = order.Quantity

	if order.Side == models.OrderSideBuy {
		currency = pair.QuoteCurrency
		if order.Type == models.OrderTypeLimit {
			amount = order.Quantity.Mul(order.Price)
		}
	} else {
		currency = pair.BaseCurrency
	}

	return s.assetService.FreezeAssetWithTx(tx, order.UserID, currency, "ERC20", amount)
}

// processTradeSettlementWithTx processes trade settlement within a transaction
func (s *OrderService) processTradeSettlementWithTx(tx *gorm.DB, trade *models.Trade) error {
	// Get trading pair
	var pair models.TradingPair
	if err := tx.Where("symbol = ?", trade.Symbol).First(&pair).Error; err != nil {
		return err
	}

	// Update buyer assets (give base currency, take quote currency)
	var buyerBaseAsset models.UserAsset
	if err := tx.Where("user_id = ? AND currency = ? AND chain = ?",
		trade.BuyerID, pair.BaseCurrency, "ERC20").First(&buyerBaseAsset).Error; err != nil {
		return err
	}

	buyerBaseAsset.Available = buyerBaseAsset.Available.Add(trade.Quantity)
	if err := tx.Save(&buyerBaseAsset).Error; err != nil {
		return err
	}

	var buyerQuoteAsset models.UserAsset
	if err := tx.Where("user_id = ? AND currency = ? AND chain = ?",
		trade.BuyerID, pair.QuoteCurrency, "ERC20").First(&buyerQuoteAsset).Error; err != nil {
		return err
	}

	buyerQuoteAsset.Frozen = buyerQuoteAsset.Frozen.Sub(trade.Amount.Add(trade.BuyerFee))
	if err := tx.Save(&buyerQuoteAsset).Error; err != nil {
		return err
	}

	// Update seller assets (take base currency, give quote currency)
	var sellerBaseAsset models.UserAsset
	if err := tx.Where("user_id = ? AND currency = ? AND chain = ?",
		trade.SellerID, pair.BaseCurrency, "ERC20").First(&sellerBaseAsset).Error; err != nil {
		return err
	}

	sellerBaseAsset.Frozen = sellerBaseAsset.Frozen.Sub(trade.Quantity.Add(trade.SellerFee))
	if err := tx.Save(&sellerBaseAsset).Error; err != nil {
		return err
	}

	var sellerQuoteAsset models.UserAsset
	if err := tx.Where("user_id = ? AND currency = ? AND chain = ?",
		trade.SellerID, pair.QuoteCurrency, "ERC20").First(&sellerQuoteAsset).Error; err != nil {
		return err
	}

	sellerQuoteAsset.Available = sellerQuoteAsset.Available.Add(trade.Amount)
	if err := tx.Save(&sellerQuoteAsset).Error; err != nil {
		return err
	}

	return nil
}
