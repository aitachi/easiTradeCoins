package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/easitradecoins/backend/internal/database"
	"github.com/easitradecoins/backend/internal/models"
	"github.com/shopspring/decimal"
)

// StopOrderMonitor monitors stop-loss and take-profit orders
type StopOrderMonitor struct {
	orderService *OrderService
	tickInterval time.Duration
	mutex        sync.RWMutex
	stopChan     chan struct{}
	running      bool
}

// NewStopOrderMonitor creates a new stop order monitor
func NewStopOrderMonitor(orderService *OrderService, tickInterval time.Duration) *StopOrderMonitor {
	return &StopOrderMonitor{
		orderService: orderService,
		tickInterval: tickInterval,
		stopChan:     make(chan struct{}),
	}
}

// Start starts the monitor
func (m *StopOrderMonitor) Start() {
	m.mutex.Lock()
	if m.running {
		m.mutex.Unlock()
		return
	}
	m.running = true
	m.mutex.Unlock()

	go m.monitorLoop()
}

// Stop stops the monitor
func (m *StopOrderMonitor) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopChan)
}

// monitorLoop is the main monitoring loop
func (m *StopOrderMonitor) monitorLoop() {
	ticker := time.NewTicker(m.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkStopOrders()
		case <-m.stopChan:
			return
		}
	}
}

// checkStopOrders checks all pending stop orders
func (m *StopOrderMonitor) checkStopOrders() {
	// Get all pending stop orders
	var stopOrders []models.Order
	if err := database.DB.Where(
		"type IN (?, ?, ?, ?) AND status = ? AND is_triggered = ?",
		models.OrderTypeStopLoss,
		models.OrderTypeTakeProfit,
		models.OrderTypeStopLimit,
		models.OrderTypeTrailingStop,
		models.OrderStatusPending,
		false,
	).Find(&stopOrders).Error; err != nil {
		fmt.Printf("Error fetching stop orders: %v\n", err)
		return
	}

	// Check each stop order
	for _, order := range stopOrders {
		m.checkAndTriggerOrder(&order)
	}
}

// checkAndTriggerOrder checks if an order should be triggered
func (m *StopOrderMonitor) checkAndTriggerOrder(order *models.Order) {
	// Get current market price
	currentPrice, err := m.getCurrentPrice(order.Symbol)
	if err != nil {
		fmt.Printf("Error getting current price for %s: %v\n", order.Symbol, err)
		return
	}

	shouldTrigger := false
	var updatedOrder *models.Order

	switch order.Type {
	case models.OrderTypeStopLoss:
		shouldTrigger = m.checkStopLoss(order, currentPrice)

	case models.OrderTypeTakeProfit:
		shouldTrigger = m.checkTakeProfit(order, currentPrice)

	case models.OrderTypeStopLimit:
		shouldTrigger = m.checkStopLimit(order, currentPrice)

	case models.OrderTypeTrailingStop:
		shouldTrigger, updatedOrder = m.checkTrailingStop(order, currentPrice)
	}

	// Update trailing stop if needed
	if updatedOrder != nil {
		database.DB.Save(updatedOrder)
	}

	// Trigger the order if conditions are met
	if shouldTrigger {
		m.triggerOrder(order, currentPrice)
	}
}

// checkStopLoss checks if a stop-loss order should be triggered
func (m *StopOrderMonitor) checkStopLoss(order *models.Order, currentPrice decimal.Decimal) bool {
	if order.StopPrice == nil {
		return false
	}

	if order.Side == models.OrderSideSell {
		// Sell stop-loss triggers when price drops below stop price
		return currentPrice.LessThanOrEqual(*order.StopPrice)
	} else {
		// Buy stop-loss triggers when price rises above stop price
		return currentPrice.GreaterThanOrEqual(*order.StopPrice)
	}
}

// checkTakeProfit checks if a take-profit order should be triggered
func (m *StopOrderMonitor) checkTakeProfit(order *models.Order, currentPrice decimal.Decimal) bool {
	if order.TakeProfitPrice == nil {
		return false
	}

	if order.Side == models.OrderSideSell {
		// Sell take-profit triggers when price rises to target
		return currentPrice.GreaterThanOrEqual(*order.TakeProfitPrice)
	} else {
		// Buy take-profit triggers when price drops to target
		return currentPrice.LessThanOrEqual(*order.TakeProfitPrice)
	}
}

// checkStopLimit checks if a stop-limit order should be triggered
func (m *StopOrderMonitor) checkStopLimit(order *models.Order, currentPrice decimal.Decimal) bool {
	if order.StopPrice == nil {
		return false
	}

	// Similar to stop-loss, but converts to limit order when triggered
	if order.Side == models.OrderSideSell {
		return currentPrice.LessThanOrEqual(*order.StopPrice)
	} else {
		return currentPrice.GreaterThanOrEqual(*order.StopPrice)
	}
}

// checkTrailingStop checks and updates trailing stop order
func (m *StopOrderMonitor) checkTrailingStop(order *models.Order, currentPrice decimal.Decimal) (bool, *models.Order) {
	if order.TrailingDelta == nil || order.StopPrice == nil {
		return false, nil
	}

	if order.Side == models.OrderSideSell {
		// For sell orders, adjust stop price upward if market price increases
		newStopPrice := currentPrice.Sub(*order.TrailingDelta)

		// Only update if new stop price is higher
		if newStopPrice.GreaterThan(*order.StopPrice) {
			order.StopPrice = &newStopPrice
			order.UpdateTime = time.Now()
			return false, order // Return updated order but don't trigger yet
		}

		// Trigger if price drops to stop price
		if currentPrice.LessThanOrEqual(*order.StopPrice) {
			return true, nil
		}
	} else {
		// For buy orders, adjust stop price downward if market price decreases
		newStopPrice := currentPrice.Add(*order.TrailingDelta)

		// Only update if new stop price is lower
		if newStopPrice.LessThan(*order.StopPrice) {
			order.StopPrice = &newStopPrice
			order.UpdateTime = time.Now()
			return false, order
		}

		// Trigger if price rises to stop price
		if currentPrice.GreaterThanOrEqual(*order.StopPrice) {
			return true, nil
		}
	}

	return false, nil
}

// triggerOrder converts a stop order to a market/limit order
func (m *StopOrderMonitor) triggerOrder(order *models.Order, currentPrice decimal.Decimal) {
	// Mark as triggered
	now := time.Now()
	order.IsTriggered = true
	order.TriggerTime = &now

	// Convert to appropriate order type
	switch order.Type {
	case models.OrderTypeStopLoss, models.OrderTypeTakeProfit, models.OrderTypeTrailingStop:
		// Convert to market order
		order.Type = models.OrderTypeMarket
		order.Price = currentPrice

	case models.OrderTypeStopLimit:
		// Convert to limit order with specified price
		order.Type = models.OrderTypeLimit
		// Keep the original limit price
	}

	// Update order in database
	if err := database.DB.Save(order).Error; err != nil {
		fmt.Printf("Error triggering order %s: %v\n", order.ID, err)
		return
	}

	// Submit the triggered order to matching engine
	if _, _, err := m.orderService.CreateOrder(order); err != nil {
		fmt.Printf("Error submitting triggered order %s: %v\n", order.ID, err)
	}
}

// getCurrentPrice gets the current market price for a symbol
func (m *StopOrderMonitor) getCurrentPrice(symbol string) (decimal.Decimal, error) {
	// Try to get the last trade price
	var lastTrade models.Trade
	if err := database.DB.Where("symbol = ?", symbol).
		Order("trade_time DESC").
		First(&lastTrade).Error; err != nil {
		// If no trades, try to get from order book (best bid/ask)
		bids, asks := m.orderService.GetOrderBookDepth(symbol, 1)
		if len(bids) > 0 && len(asks) > 0 {
			// Use mid price
			midPrice := bids[0].Price.Add(asks[0].Price).Div(decimal.NewFromInt(2))
			return midPrice, nil
		} else if len(bids) > 0 {
			return bids[0].Price, nil
		} else if len(asks) > 0 {
			return asks[0].Price, nil
		}
		return decimal.Zero, fmt.Errorf("no price data available for %s", symbol)
	}

	return lastTrade.Price, nil
}

// CreateStopLossOrder creates a stop-loss order
func CreateStopLossOrder(userID uint, symbol string, side models.OrderSide, quantity, stopPrice decimal.Decimal) *models.Order {
	triggerCondition := "<="
	if side == models.OrderSideBuy {
		triggerCondition = ">="
	}

	return &models.Order{
		UserID:           userID,
		Symbol:           symbol,
		Side:             side,
		Type:             models.OrderTypeStopLoss,
		Quantity:         quantity,
		StopPrice:        &stopPrice,
		TriggerCondition: triggerCondition,
		Status:           models.OrderStatusPending,
		TimeInForce:      models.TimeInForceGTC,
		IsTriggered:      false,
	}
}

// CreateTakeProfitOrder creates a take-profit order
func CreateTakeProfitOrder(userID uint, symbol string, side models.OrderSide, quantity, takeProfitPrice decimal.Decimal) *models.Order {
	triggerCondition := ">="
	if side == models.OrderSideBuy {
		triggerCondition = "<="
	}

	return &models.Order{
		UserID:          userID,
		Symbol:          symbol,
		Side:            side,
		Type:            models.OrderTypeTakeProfit,
		Quantity:        quantity,
		TakeProfitPrice: &takeProfitPrice,
		TriggerCondition: triggerCondition,
		Status:          models.OrderStatusPending,
		TimeInForce:     models.TimeInForceGTC,
		IsTriggered:     false,
	}
}

// CreateTrailingStopOrder creates a trailing stop order
func CreateTrailingStopOrder(userID uint, symbol string, side models.OrderSide, quantity, initialStopPrice, trailingDelta decimal.Decimal) *models.Order {
	triggerCondition := "<="
	if side == models.OrderSideBuy {
		triggerCondition = ">="
	}

	return &models.Order{
		UserID:           userID,
		Symbol:           symbol,
		Side:             side,
		Type:             models.OrderTypeTrailingStop,
		Quantity:         quantity,
		StopPrice:        &initialStopPrice,
		TrailingDelta:    &trailingDelta,
		TriggerCondition: triggerCondition,
		Status:           models.OrderStatusPending,
		TimeInForce:      models.TimeInForceGTC,
		IsTriggered:      false,
	}
}
