package matching

import (
	"sort"
	"sync"

	"github.com/easitradecoins/backend/internal/models"
	"github.com/shopspring/decimal"
)

// OrderBook represents the order book for a trading pair
type OrderBook struct {
	Symbol     string
	BuyLevels  map[string]*PriceLevel // price -> PriceLevel
	SellLevels map[string]*PriceLevel // price -> PriceLevel
	OrderMap   map[string]*models.Order // orderID -> Order
	mu         sync.RWMutex
}

// NewOrderBook creates a new order book
func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol:     symbol,
		BuyLevels:  make(map[string]*PriceLevel),
		SellLevels: make(map[string]*PriceLevel),
		OrderMap:   make(map[string]*models.Order),
	}
}

// AddOrder adds an order to the order book
func (ob *OrderBook) AddOrder(order *models.Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	priceKey := order.Price.String()

	var levels map[string]*PriceLevel
	if order.Side == models.OrderSideBuy {
		levels = ob.BuyLevels
	} else {
		levels = ob.SellLevels
	}

	// Get or create price level
	level, exists := levels[priceKey]
	if !exists {
		level = NewPriceLevel(order.Price)
		levels[priceKey] = level
	}

	level.AddOrder(order)
	ob.OrderMap[order.ID] = order
}

// RemoveOrder removes an order from the order book
func (ob *OrderBook) RemoveOrder(orderID string) bool {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	order, exists := ob.OrderMap[orderID]
	if !exists {
		return false
	}

	priceKey := order.Price.String()

	var levels map[string]*PriceLevel
	if order.Side == models.OrderSideBuy {
		levels = ob.BuyLevels
	} else {
		levels = ob.SellLevels
	}

	level, exists := levels[priceKey]
	if !exists {
		return false
	}

	if level.RemoveOrder(orderID) {
		delete(ob.OrderMap, orderID)

		// Remove empty price level
		if level.IsEmpty() {
			delete(levels, priceKey)
		}

		return true
	}

	return false
}

// GetOrder returns an order by ID
func (ob *OrderBook) GetOrder(orderID string) (*models.Order, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	order, exists := ob.OrderMap[orderID]
	return order, exists
}

// GetBestBid returns the highest buy price
func (ob *OrderBook) GetBestBid() (decimal.Decimal, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if len(ob.BuyLevels) == 0 {
		return decimal.Zero, false
	}

	var bestPrice decimal.Decimal
	first := true

	for priceStr := range ob.BuyLevels {
		price, _ := decimal.NewFromString(priceStr)
		if first || price.GreaterThan(bestPrice) {
			bestPrice = price
			first = false
		}
	}

	return bestPrice, true
}

// GetBestAsk returns the lowest sell price
func (ob *OrderBook) GetBestAsk() (decimal.Decimal, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if len(ob.SellLevels) == 0 {
		return decimal.Zero, false
	}

	var bestPrice decimal.Decimal
	first := true

	for priceStr := range ob.SellLevels {
		price, _ := decimal.NewFromString(priceStr)
		if first || price.LessThan(bestPrice) {
			bestPrice = price
			first = false
		}
	}

	return bestPrice, true
}

// GetDepth returns order book depth
func (ob *OrderBook) GetDepth(depth int) ([]PriceLevelInfo, []PriceLevelInfo) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	// Get buy levels (sorted descending)
	buyPrices := make([]decimal.Decimal, 0, len(ob.BuyLevels))
	for priceStr := range ob.BuyLevels {
		price, _ := decimal.NewFromString(priceStr)
		buyPrices = append(buyPrices, price)
	}
	sort.Slice(buyPrices, func(i, j int) bool {
		return buyPrices[i].GreaterThan(buyPrices[j])
	})

	bids := make([]PriceLevelInfo, 0, depth)
	for i := 0; i < len(buyPrices) && i < depth; i++ {
		level := ob.BuyLevels[buyPrices[i].String()]
		bids = append(bids, PriceLevelInfo{
			Price:  buyPrices[i],
			Volume: level.GetVolume(),
			Count:  level.GetOrderCount(),
		})
	}

	// Get sell levels (sorted ascending)
	sellPrices := make([]decimal.Decimal, 0, len(ob.SellLevels))
	for priceStr := range ob.SellLevels {
		price, _ := decimal.NewFromString(priceStr)
		sellPrices = append(sellPrices, price)
	}
	sort.Slice(sellPrices, func(i, j int) bool {
		return sellPrices[i].LessThan(sellPrices[j])
	})

	asks := make([]PriceLevelInfo, 0, depth)
	for i := 0; i < len(sellPrices) && i < depth; i++ {
		level := ob.SellLevels[sellPrices[i].String()]
		asks = append(asks, PriceLevelInfo{
			Price:  sellPrices[i],
			Volume: level.GetVolume(),
			Count:  level.GetOrderCount(),
		})
	}

	return bids, asks
}

// PriceLevelInfo represents aggregated price level information
type PriceLevelInfo struct {
	Price  decimal.Decimal `json:"price"`
	Volume decimal.Decimal `json:"volume"`
	Count  int             `json:"count"`
}
