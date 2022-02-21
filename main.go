package main

import (
	"flag"
	"fmt"
	"log"
	"ophrys/pkg/api"
	"ophrys/pkg/engine"
	"ophrys/pkg/market"
	"ophrys/pkg/storage"
	"os"
	"os/signal"

	"gonum.org/v1/gonum/stat"
)

type BasicTradingStrategy struct {
}

func (bst *BasicTradingStrategy) ShouldBuy(sc *engine.StatisticsCalculator, accountInformation *engine.AccountInformation) bool {
	shortLastPriceMean := sc.CalculationFor(10, "lastPriceMean")
	midLastPriceMean := sc.CalculationFor(100, "lastPriceMean")
	longLastPriceMean := sc.CalculationFor(1000, "lastPriceMean")

	log.Printf("ShouldBuy: %f, %f, %f", longLastPriceMean-midLastPriceMean, longLastPriceMean-shortLastPriceMean, midLastPriceMean-shortLastPriceMean)
	return longLastPriceMean-midLastPriceMean > 0 && longLastPriceMean-shortLastPriceMean > 0 && midLastPriceMean-shortLastPriceMean > 0

}

func (bst *BasicTradingStrategy) ShouldSell(sc *engine.StatisticsCalculator, accountInformation *engine.AccountInformation) bool {
	shortLastPriceMean := sc.CalculationFor(10, "lastPriceMean")
	midLastPriceMean := sc.CalculationFor(100, "lastPriceMean")
	longLastPriceMean := sc.CalculationFor(1000, "lastPriceMean")

	log.Printf("ShouldSell: %f, %f, %f", longLastPriceMean-midLastPriceMean, longLastPriceMean-shortLastPriceMean, midLastPriceMean-shortLastPriceMean)
	return longLastPriceMean-midLastPriceMean < 0 && longLastPriceMean-shortLastPriceMean < 0 && midLastPriceMean-shortLastPriceMean < 0
}

func (bst *BasicTradingStrategy) InvestmentAmount(accountInformation *engine.AccountInformation) float64 {
	return 0
}

func main() {
	secretKeyPtr := flag.String("secretKey", "", "Market Client Secret Key.")
	apiKeyPtr := flag.String("apiKey", "", "Market Client API Key")
	flag.Parse()
	fmt.Printf("secretKeyPtr: %s, apiKeyPtr: %s\n", *secretKeyPtr, *apiKeyPtr)

	var binanceClient engine.MarketClient = market.NewBinanceClient("https://api.binance.com", *apiKeyPtr, *secretKeyPtr)
	listenKey, err := binanceClient.(*market.BinanceClient).ListenKey()
	if err != nil {
		log.Panicln("Error obtaining listenKey from Binance: %s", err)
	}

	var pgStorage engine.Storage = storage.NewPostgresStorage("localhost", 5432, "ophrys", "ophrys", "ophrys")
	var binanceProvider engine.Provider = market.NewBinanceProvider("stream.binance.com", 9443)
	var httpApi engine.API = api.NewHttpAPI(9000)

	e := engine.NewEngine(&pgStorage)
	e.EngageAPI(&httpApi)
	e.EngageProvider(&binanceProvider)
	e.EngageMarketClient(&binanceClient)

	e.AddCalculationBuckets(10, 100, 1000)

	e.AddCalculation("lastPriceMean", func(tickers []interface{}) float64 {
		var lastPrices []float64
		for _, ticker := range tickers {
			lastPrices = append(lastPrices, ticker.(*engine.OphrysTicker).LastPrice)
		}
		return stat.Mean(lastPrices, nil)
	})

	e.AddCalculation("lastPriceStdDev", func(tickers []interface{}) float64 {
		var lastPrices []float64
		for _, ticker := range tickers {
			lastPrices = append(lastPrices, ticker.(*engine.OphrysTicker).LastPrice)
		}
		return stat.StdDev(lastPrices, nil)
	})

	var ts engine.TradingStrategy = &BasicTradingStrategy{}

	e.AddStrategyFor("ETHUSDT", &ts)

	e.TurnOn()
	(*e.GetProvider("BinanceProvider")).Subscribe("ethusdt")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
	e.TurnOff()

}
