package internal

import (
	"DevOps_Technical_challenge/internal/types"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func MakeRequest(url, method string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	return (&http.Client{}).Do(req)
}

func BatchSymbols(symbols []types.Symbol, batchSize int) map[string][]string {
	result := make(map[string][]string)
	n := 1
	var temp []string

	for _, v := range symbols {
		temp = append(temp, v.Name)

		if len(temp) == batchSize {
			x := make([]string, len(temp))
			copy(x, temp)
			result[fmt.Sprintf("batch_%d", n)] = x
			temp = []string{}
			n += 1
		}
	}

	if len(temp) > 0 {
		result[fmt.Sprintf("batch_%d", n)] = temp
	}

	return result
}

func TopNSymbolVolume(symbolStats map[string]types.SymbolStat, n int, reverse bool) []string {
	var ranking []string
	for name, sym := range symbolStats {
		symVol, _ := strconv.ParseFloat(sym.Volume, 64)

		if len(ranking) < 1 {
			ranking = append(ranking, name)
			continue
		}

		for i, v := range ranking {
			s, _ := symbolStats[v]
			sv, _ := strconv.ParseFloat(s.Volume, 64)

			if symVol > sv {
				ranking = append(ranking[:i+1], ranking[i:]...)
				ranking[i] = name
				break
			}
		}
	}

	result := ranking[:n]
	if reverse {
		for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
			result[i], result[j] = result[j], result[i]
		}
	}

	return result
}

func TopNSymbolTrades(symbolStats map[string]types.SymbolStat, n int, reverse bool) []string {
	var ranking []string
	for name, sym := range symbolStats {
		if len(ranking) < 1 {
			ranking = append(ranking, name)
			continue
		}

		for i, v := range ranking {
			s, _ := symbolStats[v]

			if sym.Count > s.Count {
				ranking = append(ranking[:i+1], ranking[i:]...)
				ranking[i] = name
				break
			}
		}
	}

	var result []string
	if len(ranking) > n {
		result = ranking[:n]
	}

	if reverse {
		for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
			result[i], result[j] = result[j], result[i]
		}
	}

	return result
}

func FilterByQuoteAsset(symbols types.SymbolList, quoteAsset string) []string {
	unique := make(map[string]bool)
	var result []string

	for _, s := range symbols.Symbols {
		if strings.ToUpper(s.QuoteAsset) == strings.ToUpper(quoteAsset) {
			if _, ok := unique[s.Name]; !ok {
				unique[s.Name] = true
				result = append(result, s.Name)
			}
		}
	}

	return result
}

func GetSymbolSpread(symbol types.SymbolStat) float64 {
	askPrice, _ := strconv.ParseFloat(symbol.AskPrice, 64)
	bidPrice, _ := strconv.ParseFloat(symbol.BidPrice, 64)

	if askPrice <= 0 {
		return 0
	}

	return (askPrice - bidPrice) / askPrice
}
