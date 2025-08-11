package strategy

//
//import (
//	"log"
//	"sync"
//	"time"
//
//	"trader/utils"
//)
//
//type Position struct {
//	EntryPrice float64
//	Size       float64
//	EntryTime  time.Time
//	EntryBar   int
//}
//
//type SMAStrategy struct {
//	mu            sync.Mutex
//	priceData     map[string][]float64
//	positions     map[string]Position
//	barCount      map[string]int
//	minHoldBars   int
//	stopLossPct   float64
//	takeProfitPct float64
//	orderHandler  func(symbol, side string, price float64)
//
//	signals map[string]map[string]bool
//}
//
//func NewSMAStrategy(minHoldBars int, stopLossPct, takeProfitPct float64, orderHandler func(string, string, float64)) *SMAStrategy {
//	return &SMAStrategy{
//		priceData:     make(map[string][]float64),
//		positions:     make(map[string]Position),
//		barCount:      make(map[string]int),
//		minHoldBars:   minHoldBars,
//		stopLossPct:   stopLossPct,
//		takeProfitPct: takeProfitPct,
//		orderHandler:  orderHandler,
//		signals:       make(map[string]map[string]bool),
//	}
//}
//
//func sma(data []float64, period int) float64 {
//	if len(data) < period {
//		return 0
//	}
//	sum := 0.0
//	for i := len(data) - period; i < len(data); i++ {
//		sum += data[i]
//	}
//	return sum / float64(period)
//}
//
//func (s *SMAStrategy) OnNewKline(symbol, interval string, k utils.Kline) {
//	key := symbol + "_" + interval
//
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	closes := s.priceData[key]
//	closes = append(closes, k.Close)
//	if len(closes) > 100 {
//		closes = closes[1:]
//	}
//	s.priceData[key] = closes
//
//	s.barCount[key]++
//
//	sma5 := sma(closes, 5)
//	sma20 := sma(closes, 20)
//	if sma5 == 0 || sma20 == 0 {
//		return
//	}
//
//	var currSignal bool
//	if sma5 > sma20 {
//		currSignal = true
//	} else {
//		currSignal = false
//	}
//
//	if _, ok := s.signals[symbol]; !ok {
//		s.signals[symbol] = make(map[string]bool)
//	}
//	s.signals[symbol][interval] = currSignal
//
//	pos := s.positions[key]
//
//	if pos.Size != 0 && s.barCount[key]-pos.EntryBar < s.minHoldBars {
//		return
//	}
//
//	allBull := true
//	allBear := true
//	for _, sig := range s.signals[symbol] {
//		if !sig {
//			allBull = false
//		}
//		if sig {
//			allBear = false
//		}
//	}
//
//	if allBull && pos.Size <= 0 {
//		log.Printf("策略信号：%s 多周期确认买入开多 %.4f", symbol, k.Close)
//		s.positions[key] = Position{
//			EntryPrice: k.Close,
//			Size:       1,
//			EntryTime:  time.Now(),
//			EntryBar:   s.barCount[key],
//		}
//		s.orderHandler(symbol, "buy", k.Close)
//		return
//	}
//
//	if allBear && pos.Size > 0 {
//		log.Printf("策略信号：%s 多周期确认卖出平多 %.4f", symbol, k.Close)
//		s.positions[key] = Position{}
//		s.orderHandler(symbol, "sell", k.Close)
//		return
//	}
//
//	if pos.Size > 0 {
//		stopLossPrice := pos.EntryPrice * (1 - s.stopLossPct)
//		takeProfitPrice := pos.EntryPrice * (1 + s.takeProfitPct)
//
//		if k.Close <= stopLossPrice {
//			log.Printf("策略信号：%s 止损卖出平多 %.4f", symbol, k.Close)
//			s.positions[key] = Position{}
//			s.orderHandler(symbol, "sell", k.Close)
//			return
//		}
//
//		if k.Close >= takeProfitPrice {
//			log.Printf("策略信号：%s 止盈卖出平多 %.4f", symbol, k.Close)
//			s.positions[key] = Position{}
//			s.orderHandler(symbol, "sell", k.Close)
//			return
//		}
//	}
//}
