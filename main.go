package main

import (
	"DevOps_Technical_challenge/internal"
	"DevOps_Technical_challenge/internal/types"
	"flag"
	"fmt"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"strconv"
	"strings"
	"time"
)

type SymbolSpread struct {
	types.SymbolStat
	Spread float64
}

const (
	BASEURL    = "https://api.binance.com"
	APIVERSION = "api/v3"
)

func init() {
	metrics.Registry.MustRegister(internal.PriceSpread)
}

func main() {
	var metricsPort int
	flag.IntVar(&metricsPort, "port", 8080, "The port prometheus binds to.")
	flag.Parse()

	zapLog, _ := zap.NewDevelopment()
	log := zapr.NewLogger(zapLog)

	altEndpoint := []string{
		"https://api1.binance.com",
		"https://api2.binance.com",
		"https://api3.binance.com",
	}

	server := internal.Server{
		BaseEndpoint: BASEURL,
		APIVersion:   APIVERSION,
		Log:          log,
		Port:         metricsPort,
	}

	if err := server.CheckConnection(); err != nil {
		server.Log.Error(err, fmt.Sprintf("unable to reach '%s'", server.BaseEndpoint))

		gotConnection := false
		for _, v := range altEndpoint {
			server.BaseEndpoint = v
			if err := server.CheckConnection(); err != nil {
				server.Log.Error(err, fmt.Sprintf("unable to reach '%s'", server.BaseEndpoint))
				continue
			}
			gotConnection = true
			break
		}

		if !gotConnection {
			return
		}
	}
	server.Log.V(1).Info(fmt.Sprintf("connection established to %s", server.BaseEndpoint))

	go func() {
		allSymbols, err := server.GetExchangeInfo()
		if err != nil {
			server.Log.Error(err, "failed to get symbol list")
			return
		}

		btcSymbols := internal.FilterByQuoteAsset(*allSymbols, "BTC") // Q1
		highestVolumeBTC := getHighestVolumeBTC(server, btcSymbols)

		usdtSymbol := internal.FilterByQuoteAsset(*allSymbols, "USDT") // Q2
		usdtStats, usdtTrades := getHighestTradesUSDT(server, usdtSymbol)

		totalNotionalValue(server, highestVolumeBTC) // Q3

		initialSpread := priceSpreadUSDT(server, usdtTrades, usdtStats) // Q4

		// Q5
		for {
			stats, trades := getHighestTradesUSDT(server, usdtSymbol)
			for i := 0; i < len(trades); i++ {
				symbol := trades[i]

				spread := priceSpreadUSDT(server, []string{symbol}, stats)
				symbolSpread, _ := spread[symbol]

				delta := symbolSpread.Spread - initialSpread[symbol].Spread
				internal.PriceSpread.WithLabelValues(symbol, fmt.Sprint(symbolSpread.Spread), fmt.Sprint(delta)).Set(symbolSpread.Spread)

				server.Log.Info(fmt.Sprintf("Delta of '%s': %v", symbol, delta), "ask_price", symbolSpread.AskPrice, "bid_price", symbolSpread.BidPrice, "delta", delta)
				initialSpread[symbol] = symbolSpread
			}
			time.Sleep(time.Second * 10)
		}
	}()

	server.Log.Info(fmt.Sprintf("starting server and serving metrics on port :%d", server.Port))
	if err := server.Start(); err != nil {
		server.Log.Error(err, "unable to start server")
		return
	}
}

func getHighestVolumeBTC(server internal.Server, btcSymbols []string) []string {
	btcStats, err := server.GetSymbolStats(btcSymbols)
	if err != nil {
		server.Log.Error(err, "unable to retrieve symbol stats")
	}

	symbolVolumes := internal.TopNSymbolVolume(btcStats, 5, false)
	server.Log.Info("highest volume over the last 24 hours in descending order", "result", fmt.Sprintf("[%s]", strings.Join(symbolVolumes, ", ")))
	return symbolVolumes
}

func getHighestTradesUSDT(server internal.Server, usdtSymbol []string) (map[string]types.SymbolStat, []string) {
	usdtStats, err := server.GetSymbolStats(usdtSymbol)
	if err != nil {
		server.Log.Error(err, "unable to retrieve symbol stats")
		return nil, nil
	}

	usdtTrades := internal.TopNSymbolTrades(usdtStats, 5, false)
	server.Log.Info("highest number of trades over the last 24 hours in descending order", "result", fmt.Sprintf("[%s]", strings.Join(usdtTrades, ", ")))

	return usdtStats, usdtTrades
}

func totalNotionalValue(server internal.Server, btcSymbols []string) {
	limit := 200

	for i := 0; i < len(btcSymbols); i++ {
		orderBook, err := server.GetOrderBook(btcSymbols[i], limit)
		if err != nil {
			server.Log.Error(err, "unable to retrieve order book", "symbol", btcSymbols[i])
		}

		var totalBids, totalAsks float64
		for j := 0; j < limit; j++ {
			if j < len(orderBook.Bids) {
				bidPrice, _ := strconv.ParseFloat(orderBook.Bids[j][0], 64)
				bidQty, _ := strconv.ParseFloat(orderBook.Bids[j][1], 64)
				totalBids += bidPrice * bidQty
			}

			if j < len(orderBook.Asks) {
				askPrice, _ := strconv.ParseFloat(orderBook.Asks[j][0], 64)
				askQty, _ := strconv.ParseFloat(orderBook.Asks[j][1], 64)
				totalAsks += askPrice * askQty
			}
		}

		server.Log.Info(fmt.Sprintf("total notional value of top %v %v bids: %v", limit, btcSymbols[i], totalBids))
		server.Log.Info(fmt.Sprintf("total notional value of top %v %v asks: %v", limit, btcSymbols[i], totalAsks))
	}
}

func priceSpreadUSDT(server internal.Server, symbolTrades []string, usdtStats map[string]types.SymbolStat) map[string]SymbolSpread {
	result := map[string]SymbolSpread{}

	for _, v := range symbolTrades {
		s, ok := usdtStats[v]
		if !ok {
			symbolStats, err := server.GetSymbolStats([]string{v})
			if err != nil {
				server.Log.Error(err, "unable to retrieve symbol stats")
				continue
			}

			exists := false
			s, exists = symbolStats[v]
			if !exists {
				server.Log.Error(err, "unable to retrieve symbol stats")
				continue
			}
		}

		spread := internal.GetSymbolSpread(s)
		server.Log.Info(fmt.Sprintf("Bid-Ask Spread of '%s': %v", v, spread), "ask_price", s.AskPrice, "bid_price", s.BidPrice)
		result[v] = SymbolSpread{
			SymbolStat: s,
			Spread:     spread,
		}

	}

	return result
}
