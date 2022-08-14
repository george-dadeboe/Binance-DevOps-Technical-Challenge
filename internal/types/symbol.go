package types

type Symbol struct {
	Name       string `json:"symbol"`
	QuoteAsset string `json:"quoteAsset"`
}

type SymbolList struct {
	Timezone   string   `json:"timezone"`
	ServerTime int64    `json:"serverTime"`
	Symbols    []Symbol `json:"symbols"`
}

type SymbolStat struct {
	Name     string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	AskPrice string `json:"askPrice"`
	Volume   string `json:"volume"`
	Count    int    `json:"count"`
}
