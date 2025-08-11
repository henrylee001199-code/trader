package utils

import "time"

// Kline 是单根K线数据结构
type Kline struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// KlineWithSymbol 带币种和周期的K线，用于跨包传递
type KlineWithSymbol struct {
	Symbol   string
	Interval string
	Kline    Kline
}
