package strategy

import (
	"time"
	"trader/utils"
)

// TrendFollow: 大周期判断趋势（EMA50/200），小周期用 ATR 回调入场
func TrendFollow(symbol string, klines4H, klines15M []utils.Kline, equity float64) *utils.Signal {
	// prepare closes for 4H
	var closes4 []float64
	for _, k := range klines4H {
		closes4 = append(closes4, k.Close)
	}
	ema50 := utils.EMA(closes4, 50)
	ema200 := utils.EMA(closes4, 200)
	if ema50 == nil || ema200 == nil {
		return nil
	}
	// use last values
	lastIdx := len(ema50) - 1
	if lastIdx < 0 || len(ema200)-1 < 0 {
		return nil
	}
	trendUp := ema50[lastIdx] > ema200[lastIdx]
	trendDown := ema50[lastIdx] < ema200[lastIdx]
	if !trendUp && !trendDown {
		return nil
	}

	// ATR on 15m
	atr := utils.ATR(klines15M, 14)
	if atr == 0 {
		return nil
	}

	// risk sizing: single trade risk 1% equity, stop = 2*ATR
	riskPerTrade := equity * 0.01
	size := riskPerTrade / (atr * 2) // base-asset amount

	// last close on 15m
	lastClose := klines15M[len(klines15M)-1].Close

	if trendUp {
		return &utils.Signal{
			Symbol:     symbol,
			Direction:  1,
			EntryPrice: lastClose,
			StopLoss:   lastClose - 2*atr,
			Size:       size,
			Time:       time.Now(),
		}
	}
	if trendDown {
		return &utils.Signal{
			Symbol:     symbol,
			Direction:  -1,
			EntryPrice: lastClose,
			StopLoss:   lastClose + 2*atr,
			Size:       size,
			Time:       time.Now(),
		}
	}
	return nil
}
