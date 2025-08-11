package utils

import (
	"math"
)

// EMA 计算（简单实现，seed 为第一个值）
func EMA(values []float64, period int) []float64 {
	n := len(values)
	if n < period || period <= 0 {
		return nil
	}
	result := make([]float64, n)
	multiplier := 2.0 / float64(period+1)
	// seed with simple average of first `period` values
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	result[period-1] = sum / float64(period)
	// for earlier indexes (< period-1) we keep as zero (not used)
	for i := period; i < n; i++ {
		result[i] = (values[i]-result[i-1])*multiplier + result[i-1]
	}
	// shift so last index corresponds (user should use last element which is at index n-1)
	// For simplicity, fill first elements by copying earliest computed ema
	for i := 0; i < period-1; i++ {
		result[i] = result[period-1]
	}
	return result
}

// ATR 计算（返回当前 ATR 值，基于 period）
func ATR(klines []Kline, period int) float64 {
	if len(klines) < period+1 {
		return 0
	}
	trs := make([]float64, 0, len(klines)-1)
	for i := 1; i < len(klines); i++ {
		h := klines[i].High
		l := klines[i].Low
		pc := klines[i-1].Close
		tr := math.Max(h-l, math.Max(math.Abs(h-pc), math.Abs(l-pc)))
		trs = append(trs, tr)
	}
	if len(trs) < period {
		return 0
	}
	sum := 0.0
	for i := len(trs) - period; i < len(trs); i++ {
		sum += trs[i]
	}
	return sum / float64(period)
}
