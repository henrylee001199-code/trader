package strategy

import (
	"log"
	"sync"
	"time"
)

// 引用 utils.Kline
import "trader/utils"

type Position struct {
	EntryPrice float64
	Size       float64 // 多仓正数，空仓负数，0表示空仓
	EntryTime  time.Time
}

type SMAStrategy struct {
	mu          sync.Mutex
	priceData   map[string][]float64 // key: symbol_interval, 只存Close价
	positions   map[string]Position
	minHoldBars int            // 最短持仓K线数，比如5根15m
	barCount    map[string]int // 当前K线数量计数，辅助最短持仓判断

	orderHandler func(symbol, side string, price float64)
}

func NewSMAStrategy(minHoldBars int, orderHandler func(string, string, float64)) *SMAStrategy {
	return &SMAStrategy{
		priceData:    make(map[string][]float64),
		positions:    make(map[string]Position),
		minHoldBars:  minHoldBars,
		barCount:     make(map[string]int),
		orderHandler: orderHandler,
	}
}

// 计算简单移动均线
func sma(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}
	sum := 0.0
	for i := len(data) - period; i < len(data); i++ {
		sum += data[i]
	}
	return sum / float64(period)
}

func (s *SMAStrategy) OnNewKline(symbol, interval string, k utils.Kline) {
	key := symbol + "_" + interval
	s.mu.Lock()
	defer s.mu.Unlock()

	// 保存收盘价
	closes := s.priceData[key]
	closes = append(closes, k.Close)
	if len(closes) > 100 { // 保留最近100根收盘价，防止内存无限增长
		closes = closes[1:]
	}
	s.priceData[key] = closes

	// 计数
	s.barCount[key]++

	// 计算SMA5和SMA20
	sma5 := sma(closes, 5)
	sma20 := sma(closes, 20)

	if sma5 == 0 || sma20 == 0 {
		// 数据不足，跳过
		return
	}

	pos := s.positions[key]

	// 买卖信号逻辑
	// 1. 最短持仓限制
	if pos.Size != 0 && s.barCount[key]-pos.EntryTimeBar() < s.minHoldBars {
		// 持仓未达到最短周期，跳过平仓判断
		return
	}

	// 2. SMA交叉判断
	if sma5 > sma20 && pos.Size <= 0 {
		// 多头开仓
		log.Printf("策略信号：%s %s 买入开多 %.4f", symbol, interval, k.Close)
		s.positions[key] = Position{
			EntryPrice: k.Close,
			Size:       1,
			EntryTime:  time.Now(),
		}
		s.orderHandler(symbol, "buy", k.Close)
		return
	}

	if sma5 < sma20 && pos.Size > 0 {
		// 多头平仓
		log.Printf("策略信号：%s %s 卖出平多 %.4f", symbol, interval, k.Close)
		s.positions[key] = Position{}
		s.orderHandler(symbol, "sell", k.Close)
		return
	}
}

// Position 添加辅助方法计算持仓时长条数
func (p *Position) EntryTimeBar() int {
	// 这里暂时用固定值，实际你可以用策略结构体里的barCount和EntryTime结合实现更精准逻辑
	return 0
}
