//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

package matching

import (
	"errors"
	"sync"
	"time"

	"github.com/easitradecoins/backend/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MatchingEngine handles order matching
type MatchingEngine struct {
	orderBooks map[string]*OrderBook // symbol -> OrderBook
	mu         sync.RWMutex
	tradeChan  chan *models.Trade
}

// NewMatchingEngine creates a new matching engine
func NewMatchingEngine() *MatchingEngine {
	return &MatchingEngine{
		orderBooks: make(map[string]*OrderBook),
		tradeChan:  make(chan *models.Trade, 10000),
	}
}

// GetOrCreateOrderBook gets or creates an order book for a symbol
func (me *MatchingEngine) GetOrCreateOrderBook(symbol string) *OrderBook {
	me.mu.Lock()
	defer me.mu.Unlock()

	ob, exists := me.orderBooks[symbol]
	if !exists {
		ob = NewOrderBook(symbol)
		me.orderBooks[symbol] = ob
	}

	return ob
}

// ProcessOrder processes a new order
func (me *MatchingEngine) ProcessOrder(order *models.Order) ([]*models.Trade, error) {
	if order == nil {
		return nil, errors.New("order is nil")
	}

	// Validate order
	if err := me.validateOrder(order); err != nil {
		return nil, err
	}

	ob := me.GetOrCreateOrderBook(order.Symbol)

	var trades []*models.Trade

	// Match market order or limit order
	if order.Type == models.OrderTypeMarket {
		trades = me.matchMarketOrder(ob, order)
	} else {
		trades = me.matchLimitOrder(ob, order)
	}

	// If order is not fully filled and not IOC/FOK, add to order book
	if order.Status == models.OrderStatusPending || order.Status == models.OrderStatusPartial {
		if order.TimeInForce == models.TimeInForceGTC {
			ob.AddOrder(order)
		} else if order.TimeInForce == models.TimeInForceIOC {
			// IOC: cancel remaining
			order.Status = models.OrderStatusCancelled
		} else if order.TimeInForce == models.TimeInForceFOK {
			// FOK: if not fully filled, cancel
			if !order.FilledQty.Equal(order.Quantity) {
				order.Status = models.OrderStatusCancelled
				// Rollback trades (in real implementation, use database transaction)
				trades = nil
			}
		}
	}

	// Send trades to channel
	for _, trade := range trades {
		me.tradeChan <- trade
	}

	return trades, nil
}

// matchLimitOrder matches a limit order
func (me *MatchingEngine) matchLimitOrder(ob *OrderBook, order *models.Order) []*models.Trade {
	var trades []*models.Trade

	order.Status = models.OrderStatusPending

	// Match buy order
	if order.Side == models.OrderSideBuy {
		bestAsk, hasAsk := ob.GetBestAsk()

		for hasAsk && bestAsk.LessThanOrEqual(order.Price) {
			level := ob.SellLevels[bestAsk.String()]
			if level == nil {
				break
			}

			makerOrder := level.GetFirstOrder()
			if makerOrder == nil {
				break
			}

			// Execute trade
			trade := me.executeTrade(order, makerOrder, bestAsk)
			if trade != nil {
				trades = append(trades, trade)
			}

			// Update maker order
			if makerOrder.FilledQty.Equal(makerOrder.Quantity) {
				makerOrder.Status = models.OrderStatusFilled
				ob.RemoveOrder(makerOrder.ID)
			} else {
				makerOrder.Status = models.OrderStatusPartial
				level.UpdateVolume()
			}

			// Check if taker order is filled
			if order.FilledQty.Equal(order.Quantity) {
				order.Status = models.OrderStatusFilled
				break
			}

			// Get next best ask
			bestAsk, hasAsk = ob.GetBestAsk()
		}
	} else {
		// Match sell order
		bestBid, hasBid := ob.GetBestBid()

		for hasBid && bestBid.GreaterThanOrEqual(order.Price) {
			level := ob.BuyLevels[bestBid.String()]
			if level == nil {
				break
			}

			makerOrder := level.GetFirstOrder()
			if makerOrder == nil {
				break
			}

			// Execute trade
			trade := me.executeTrade(makerOrder, order, bestBid)
			if trade != nil {
				trades = append(trades, trade)
			}

			// Update maker order
			if makerOrder.FilledQty.Equal(makerOrder.Quantity) {
				makerOrder.Status = models.OrderStatusFilled
				ob.RemoveOrder(makerOrder.ID)
			} else {
				makerOrder.Status = models.OrderStatusPartial
				level.UpdateVolume()
			}

			// Check if taker order is filled
			if order.FilledQty.Equal(order.Quantity) {
				order.Status = models.OrderStatusFilled
				break
			}

			// Get next best bid
			bestBid, hasBid = ob.GetBestBid()
		}
	}

	return trades
}

// matchMarketOrder matches a market order
func (me *MatchingEngine) matchMarketOrder(ob *OrderBook, order *models.Order) []*models.Trade {
	var trades []*models.Trade

	order.Status = models.OrderStatusPending

	// Market buy order
	if order.Side == models.OrderSideBuy {
		bestAsk, hasAsk := ob.GetBestAsk()

		for hasAsk && order.FilledQty.LessThan(order.Quantity) {
			level := ob.SellLevels[bestAsk.String()]
			if level == nil {
				break
			}

			makerOrder := level.GetFirstOrder()
			if makerOrder == nil {
				break
			}

			// Execute trade at maker's price
			trade := me.executeTrade(order, makerOrder, makerOrder.Price)
			if trade != nil {
				trades = append(trades, trade)
			}

			// Update maker order
			if makerOrder.FilledQty.Equal(makerOrder.Quantity) {
				makerOrder.Status = models.OrderStatusFilled
				ob.RemoveOrder(makerOrder.ID)
			} else {
				makerOrder.Status = models.OrderStatusPartial
				level.UpdateVolume()
			}

			// Check if taker order is filled
			if order.FilledQty.Equal(order.Quantity) {
				order.Status = models.OrderStatusFilled
				break
			}

			// Get next best ask
			bestAsk, hasAsk = ob.GetBestAsk()
		}
	} else {
		// Market sell order
		bestBid, hasBid := ob.GetBestBid()

		for hasBid && order.FilledQty.LessThan(order.Quantity) {
			level := ob.BuyLevels[bestBid.String()]
			if level == nil {
				break
			}

			makerOrder := level.GetFirstOrder()
			if makerOrder == nil {
				break
			}

			// Execute trade at maker's price
			trade := me.executeTrade(makerOrder, order, makerOrder.Price)
			if trade != nil {
				trades = append(trades, trade)
			}

			// Update maker order
			if makerOrder.FilledQty.Equal(makerOrder.Quantity) {
				makerOrder.Status = models.OrderStatusFilled
				ob.RemoveOrder(makerOrder.ID)
			} else {
				makerOrder.Status = models.OrderStatusPartial
				level.UpdateVolume()
			}

			// Check if taker order is filled
			if order.FilledQty.Equal(order.Quantity) {
				order.Status = models.OrderStatusFilled
				break
			}

			// Get next best bid
			bestBid, hasBid = ob.GetBestBid()
		}
	}

	// Market order must be filled or cancelled
	if !order.FilledQty.Equal(order.Quantity) {
		order.Status = models.OrderStatusCancelled
	}

	return trades
}

// executeTrade executes a trade between two orders
func (me *MatchingEngine) executeTrade(buyOrder, sellOrder *models.Order, price decimal.Decimal) *models.Trade {
	// Calculate trade quantity
	buyRemaining := buyOrder.Quantity.Sub(buyOrder.FilledQty)
	sellRemaining := sellOrder.Quantity.Sub(sellOrder.FilledQty)

	var tradeQty decimal.Decimal
	if buyRemaining.LessThan(sellRemaining) {
		tradeQty = buyRemaining
	} else {
		tradeQty = sellRemaining
	}

	if tradeQty.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	// Calculate trade amount
	tradeAmount := tradeQty.Mul(price)

	// Calculate fees (0.1% for simplicity)
	feeRate := decimal.NewFromFloat(0.001)
	buyerFee := tradeAmount.Mul(feeRate)
	sellerFee := tradeQty.Mul(feeRate)

	// Update orders
	buyOrder.FilledQty = buyOrder.FilledQty.Add(tradeQty)
	buyOrder.FilledAmount = buyOrder.FilledAmount.Add(tradeAmount)
	buyOrder.Fee = buyOrder.Fee.Add(buyerFee)
	buyOrder.UpdateTime = time.Now()

	if buyOrder.FilledQty.GreaterThan(decimal.Zero) {
		buyOrder.AvgPrice = buyOrder.FilledAmount.Div(buyOrder.FilledQty)
	}

	sellOrder.FilledQty = sellOrder.FilledQty.Add(tradeQty)
	sellOrder.FilledAmount = sellOrder.FilledAmount.Add(tradeAmount)
	sellOrder.Fee = sellOrder.Fee.Add(sellerFee)
	sellOrder.UpdateTime = time.Now()

	if sellOrder.FilledQty.GreaterThan(decimal.Zero) {
		sellOrder.AvgPrice = sellOrder.FilledAmount.Div(sellOrder.FilledQty)
	}

	// Create trade record
	trade := &models.Trade{
		ID:          uuid.New().String(),
		Symbol:      buyOrder.Symbol,
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		BuyerID:     buyOrder.UserID,
		SellerID:    sellOrder.UserID,
		Price:       price,
		Quantity:    tradeQty,
		Amount:      tradeAmount,
		BuyerFee:    buyerFee,
		SellerFee:   sellerFee,
		TradeTime:   time.Now(),
	}

	return trade
}

// CancelOrder cancels an order
func (me *MatchingEngine) CancelOrder(symbol, orderID string) error {
	ob := me.GetOrCreateOrderBook(symbol)

	order, exists := ob.GetOrder(orderID)
	if !exists {
		return errors.New("order not found")
	}

	if order.Status == models.OrderStatusFilled || order.Status == models.OrderStatusCancelled {
		return errors.New("order cannot be cancelled")
	}

	ob.RemoveOrder(orderID)
	order.Status = models.OrderStatusCancelled
	order.UpdateTime = time.Now()

	return nil
}

// validateOrder validates an order
func (me *MatchingEngine) validateOrder(order *models.Order) error {
	if order.Symbol == "" {
		return errors.New("symbol is required")
	}

	if order.Quantity.LessThanOrEqual(decimal.Zero) {
		return errors.New("quantity must be positive")
	}

	if order.Type == models.OrderTypeLimit && order.Price.LessThanOrEqual(decimal.Zero) {
		return errors.New("price must be positive for limit order")
	}

	if order.Side != models.OrderSideBuy && order.Side != models.OrderSideSell {
		return errors.New("invalid order side")
	}

	return nil
}

// GetOrderBook returns an order book for a symbol
func (me *MatchingEngine) GetOrderBook(symbol string) (*OrderBook, bool) {
	me.mu.RLock()
	defer me.mu.RUnlock()

	ob, exists := me.orderBooks[symbol]
	return ob, exists
}

// GetTradeChan returns the trade channel
func (me *MatchingEngine) GetTradeChan() <-chan *models.Trade {
	return me.tradeChan
}
