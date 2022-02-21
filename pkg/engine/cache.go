package engine

import (
	"ophrys/pkg/adt"
)

type Cache struct {
	AccountInformation *AccountInformation
	LastDepths         *adt.ConcurrentMap
	LastTickers        *adt.ConcurrentMap
	ordersIndex        *adt.ConcurrentMap
	ordersBySymbol     *adt.ConcurrentMap
}

func NewCache() *Cache {
	return &Cache{
		LastDepths:  adt.NewConcurrentMap(),
		LastTickers: adt.NewConcurrentMap(),
	}
}

func (c *Cache) UpdateLastDepth(lastDepth *OphrysDepth) {
	c.LastDepths.Put(lastDepth.Symbol, lastDepth)
}

func (c *Cache) UpdateLastTicker(lastTicker *OphrysTicker) {
	c.LastTickers.Put(lastTicker.Symbol, lastTicker)
}

func (c *Cache) GetLastTicker(symbol string) *OphrysTicker {
	ticker, ok := c.LastTickers.Get(symbol)

	if !ok {
		return nil
	}

	return ticker.(*OphrysTicker)
}

func (c *Cache) GetLastDepth(symbol string) *OphrysDepth {
	depth, ok := c.LastDepths.Get(symbol)

	if !ok {
		return nil
	}

	return depth.(*OphrysDepth)
}

func (c *Cache) UpdateAccountInformation(accountInformation *AccountInformation) {
	c.AccountInformation = accountInformation
}

func (c *Cache) UpdateOrder(o *Order) {
	if !c.ordersIndex.Has(o.OrderId) {
		c.ordersIndex.Put(o.OrderId, o)

		orderIds, hasSymbol := c.ordersBySymbol.Get(o.Symbol)

		if !hasSymbol {
			orderIds = make([]int, 0)
			c.ordersBySymbol.Put(o.Symbol, orderIds)
		}

		c.ordersBySymbol.Put(o.Symbol, append(orderIds.([]int), o.OrderId))
	}

	order, _ := c.ordersIndex.Get(o.OrderId)

	order.(*Order).Status = o.Status

}
