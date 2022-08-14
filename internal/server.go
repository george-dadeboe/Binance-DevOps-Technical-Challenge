package internal

import (
	"DevOps_Technical_challenge/internal/types"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"net/http"
	"net/url"
)

const (
	PING             = "ping"
	EXCHANGEINFO     = "exchangeInfo"
	SYMBOLTICKER24HR = "ticker/24hr"
	ORDERBOOK        = "depth"
)

type Server struct {
	BaseEndpoint string
	APIVersion   string
	Log          logr.Logger
	Port         int
}

func (server Server) CheckConnection() error {
	if _, err := MakeRequest(fmt.Sprintf("%s/%s/%s", server.BaseEndpoint, server.APIVersion, PING), http.MethodGet); err != nil {
		return err
	}

	return nil
}

func (server Server) GetExchangeInfo() (*types.SymbolList, error) {
	response, err := MakeRequest(fmt.Sprintf("%s/%s/%s", server.BaseEndpoint, server.APIVersion, EXCHANGEINFO), http.MethodGet)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var result types.SymbolList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (server Server) GetSymbolStats(symbols []string) (map[string]types.SymbolStat, error) {
	data, err := json.Marshal(symbols)
	if err != nil {
		return nil, err
	}

	query := url.QueryEscape(string(data))

	res, err := MakeRequest(fmt.Sprintf("%s/%s/%s?symbols=%s", server.BaseEndpoint, server.APIVersion, SYMBOLTICKER24HR, query), http.MethodGet)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var symbolStatList []types.SymbolStat
	if err := json.Unmarshal(body, &symbolStatList); err != nil {
		return nil, err
	}

	result := make(map[string]types.SymbolStat)
	for _, v := range symbolStatList {
		result[v.Name] = v
	}

	return result, nil
}

func (server Server) GetOrderBook(symbol string, limit int) (*types.OrderBook, error) {
	res, err := MakeRequest(fmt.Sprintf("%s/%s/%s?symbol=%s&limit=%v", server.BaseEndpoint, server.APIVersion, ORDERBOOK, symbol, limit), http.MethodGet)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var orderBook types.OrderBook
	if err := json.Unmarshal(body, &orderBook); err != nil {
		return nil, err
	}

	return &orderBook, nil
}

func (server Server) Start() error {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%v", server.Port), nil); err != nil {
		return err
	}
	return nil
}
