package strategy

import (
	"sync"
)

type SimpleMA struct {
	shortPeriod int
	longPeriod  int

	prices []float64

	mu sync.Mutex
}

func NewSimpleMA(short, long int) *SimpleMA {
	return &SimpleMA{
		shortPeriod: short,
		longPeriod:  long,
		prices:      make([]float64, 0, long),
	}
}

// 添加新价格，触发信号返回买卖操作，空字符串表示无信号
func (s *SimpleMA) OnNewPrice(price float64) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.prices = append(s.prices, price)
	if len(s.prices) > s.longPeriod {
		s.prices = s.prices[len(s.prices)-s.longPeriod:]
	}

	if len(s.prices) < s.longPeriod {
		return ""
	}

	shortMA := average(s.prices[len(s.prices)-s.shortPeriod:])
	longMA := average(s.prices)

	// 简单交叉判断（可优化状态控制避免频繁信号）
	if shortMA > longMA {
		return "buy"
	} else if shortMA < longMA {
		return "sell"
	}
	return ""
}

func average(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}
