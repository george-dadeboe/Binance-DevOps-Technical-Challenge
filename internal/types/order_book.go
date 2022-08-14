package types

type OrderBook struct {
	LastUpdateId int        `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}
