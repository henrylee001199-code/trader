package utils

type Kline struct {
	StartTime      int64   `json:"t"`
	CloseTime      int64   `json:"T"`
	Symbol         string  `json:"-"` // 这个字段自己赋值，JSON里没有
	Interval       string  `json:"i"`
	FirstTradeID   int64   `json:"f"`
	LastTradeID    int64   `json:"L"`
	Open           float64 `json:"o,string"`
	Close          float64 `json:"c,string"`
	High           float64 `json:"h,string"`
	Low            float64 `json:"l,string"`
	Volume         float64 `json:"v,string"`
	Trades         int64   `json:"n"`
	IsClosed       bool    `json:"x"`
	QuoteVolume    float64 `json:"q,string"`
	BuyVolume      float64 `json:"V,string"`
	BuyQuoteVolume float64 `json:"Q,string"`
}
