package strategy

type SimpleMA struct {
	shortPeriod int
	longPeriod  int
	prices      []float64
	shortMA     float64
	longMA      float64
}

func NewSimpleMA(shortPeriod, longPeriod int) *SimpleMA {
	return &SimpleMA{
		shortPeriod: shortPeriod,
		longPeriod:  longPeriod,
		prices:      make([]float64, 0, longPeriod),
	}
}

// 计算简单移动平均
func movingAverage(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range prices {
		sum += p
	}
	return sum / float64(len(prices))
}

// OnNewPrice 接收最新价格，计算短期和长期均线，判断交叉，返回交易信号
func (s *SimpleMA) OnNewPrice(price float64) string {
	// 添加新价格
	s.prices = append(s.prices, price)
	// 保持长度不超过longPeriod
	if len(s.prices) > s.longPeriod {
		s.prices = s.prices[1:]
	}

	if len(s.prices) < s.longPeriod {
		// 数据不足，暂不产生信号
		return ""
	}

	// 计算短期和长期均线
	s.shortMA = movingAverage(s.prices[len(s.prices)-s.shortPeriod:])
	s.longMA = movingAverage(s.prices)

	// 判断均线交叉
	// 简单示例：如果短期均线高于长期均线，返回买入信号；反之卖出信号
	if s.shortMA > s.longMA {
		return "买入开多"
	}
	if s.shortMA < s.longMA {
		return "卖出平多"
	}
	return ""
}
