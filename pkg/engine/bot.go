package engine

import (
	"log"
)

type TradingBot struct {
	engine     *Engine
	strategies map[string]*TradingStrategy
	c          chan interface{}
}

type TradingStrategy interface {
	ShouldBuy(sc *StatisticsCalculator, accountInformation *AccountInformation) bool
	ShouldSell(sc *StatisticsCalculator, accountInformation *AccountInformation) bool
	InvestmentAmount(accountInformation *AccountInformation) float64
}

func NewTradingBot(e *Engine) *TradingBot {
	return &TradingBot{
		engine:     e,
		strategies: make(map[string]*TradingStrategy),
		c:          make(chan interface{}),
	}
}
func (t *TradingBot) C() chan interface{} {
	return t.c
}
func (t *TradingBot) AddStrategyFor(symbol string, ts *TradingStrategy) {
	t.strategies[symbol] = ts
}

func (tb *TradingBot) HandleNewTicker(w *Worker, ticker interface{}) {
	t := ticker.(*OphrysTicker)

	strategy, hasStrategy := tb.strategies[t.Symbol]

	if !hasStrategy {
		return
	}

	stats, hasStats := tb.engine.stats.statisticsCalculators.Get(t.Symbol)

	if !hasStats {
		log.Printf("ERROR: No stats for symbol: %s", t.Symbol)
		return
	}

	if (*strategy).ShouldBuy(stats.(*StatisticsCalculator), tb.engine.cache.AccountInformation) {
		log.Printf("Buy: %s", t.Symbol)
	}

	if (*strategy).ShouldSell(stats.(*StatisticsCalculator), tb.engine.cache.AccountInformation) {
		log.Printf("Sell: %s", t.Symbol)
	}

}
