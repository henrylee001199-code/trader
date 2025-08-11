package strategy

// Strategy 定义策略接口
type Strategy interface {
	OnNewPrice(price float64) string
}

// CompositeStrategy 聚合多个策略
type CompositeStrategy struct {
	strategies []Strategy
}

// NewCompositeStrategy 创建复合策略
func NewCompositeStrategy(strats ...Strategy) *CompositeStrategy {
	return &CompositeStrategy{
		strategies: strats,
	}
}

// OnNewPrice 依次调用每个子策略的 OnNewPrice 方法，
// 如果有策略返回非空信号，就立即返回该信号
func (c *CompositeStrategy) OnNewPrice(price float64) string {
	for _, strat := range c.strategies {
		signal := strat.OnNewPrice(price)
		if signal != "" {
			return signal
		}
	}
	return ""
}
