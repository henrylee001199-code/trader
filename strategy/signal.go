package strategy

//
//import (
//	"log"
//	"sync"
//	"trader/utils"
//)
//
//// Position 仓位信息
//type Position struct {
//	EntryPrice float64
//	Size       float64 // 正为多，负为空
//}
//
//// Strategy 管理多币种多周期状态
//type Strategy struct {
//	mu sync.Mutex
//
//	// symbol_interval -> k线序列
//	priceData map[string][]utils.Kline
//
//	// symbol_interval -> 持仓信息
//	positions map[string]Position
//
//	// 模拟账户下单回调函数
//	orderHandler func(symbol string, side string, price float64)
//}
//
//func NewStrategy(orderHandler func(string, string, float64)) *Strategy {
//	return &Strategy{
//		priceData:    make(map[string][]utils.Kline),
//		positions:    make(map[string]Position),
//		orderHandler: orderHandler,
//	}
//}
//
//func (s *Strategy) OnNewKline(symbol, interval string, k utils.Kline) {
//	key := symbol + "_" + interval
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	// 保存K线，简单示例只保留最近20根
//	data := s.priceData[key]
//	if len(data) >= 20 {
//		data = data[1:]
//	}
//	data = append(data, k)
//	s.priceData[key] = data
//
//	// 至少要有两根K线才判断
//	if len(data) < 2 {
//		return
//	}
//
//	// 简单MA交叉示例，取近2根均价
//	prevClose := data[len(data)-2].Close
//	currClose := data[len(data)-1].Close
//
//	pos := s.positions[key]
//
//	// 简单策略示例：
//	// 价格上涨且无仓位，买入开多
//	if currClose > prevClose && pos.Size <= 0 {
//		log.Printf("策略信号：%s %s 买入开多 %.4f", symbol, interval, currClose)
//		s.positions[key] = Position{EntryPrice: currClose, Size: 1}
//		s.orderHandler(symbol, "buy", currClose)
//		return
//	}
//
//	// 价格下跌且有多仓，卖出平仓
//	if currClose < prevClose && pos.Size > 0 {
//		log.Printf("策略信号：%s %s 卖出平多 %.4f", symbol, interval, currClose)
//		s.positions[key] = Position{}
//		s.orderHandler(symbol, "sell", currClose)
//		return
//	}
//}
