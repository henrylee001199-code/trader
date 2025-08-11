package utils

import "time"

type Signal struct {
	Symbol     string
	Direction  int // 1=多 -1=空
	EntryPrice float64
	StopLoss   float64
	Size       float64
	Time       time.Time
}
