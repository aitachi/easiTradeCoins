//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

package matching

import (
	"container/list"
	"sync"

	"github.com/easitradecoins/backend/internal/models"
	"github.com/shopspring/decimal"
)

// PriceLevel represents orders at a specific price
type PriceLevel struct {
	Price      decimal.Decimal
	Volume     decimal.Decimal
	OrderCount int
	Orders     *list.List // list of *models.Order
	mu         sync.RWMutex
}

// NewPriceLevel creates a new price level
func NewPriceLevel(price decimal.Decimal) *PriceLevel {
	return &PriceLevel{
		Price:      price,
		Volume:     decimal.Zero,
		OrderCount: 0,
		Orders:     list.New(),
	}
}

// AddOrder adds an order to the price level
func (pl *PriceLevel) AddOrder(order *models.Order) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	pl.Orders.PushBack(order)
	pl.Volume = pl.Volume.Add(order.Quantity.Sub(order.FilledQty))
	pl.OrderCount++
}

// RemoveOrder removes an order from the price level
func (pl *PriceLevel) RemoveOrder(orderID string) bool {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	for e := pl.Orders.Front(); e != nil; e = e.Next() {
		order := e.Value.(*models.Order)
		if order.ID == orderID {
			pl.Orders.Remove(e)
			remainingQty := order.Quantity.Sub(order.FilledQty)
			pl.Volume = pl.Volume.Sub(remainingQty)
			pl.OrderCount--
			return true
		}
	}
	return false
}

// GetFirstOrder returns the first order in the queue
func (pl *PriceLevel) GetFirstOrder() *models.Order {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	if pl.Orders.Len() == 0 {
		return nil
	}

	front := pl.Orders.Front()
	if front == nil {
		return nil
	}

	return front.Value.(*models.Order)
}

// UpdateVolume recalculates total volume
func (pl *PriceLevel) UpdateVolume() {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	volume := decimal.Zero
	for e := pl.Orders.Front(); e != nil; e = e.Next() {
		order := e.Value.(*models.Order)
		remainingQty := order.Quantity.Sub(order.FilledQty)
		volume = volume.Add(remainingQty)
	}
	pl.Volume = volume
}

// IsEmpty returns true if there are no orders
func (pl *PriceLevel) IsEmpty() bool {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.Orders.Len() == 0
}

// GetVolume returns total volume at this price level
func (pl *PriceLevel) GetVolume() decimal.Decimal {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.Volume
}

// GetOrderCount returns number of orders at this price level
func (pl *PriceLevel) GetOrderCount() int {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.OrderCount
}
