package strategy

type RSIStrategy struct {
	period      int
	gains       []float64
	losses      []float64
	avgGain     float64
	avgLoss     float64
	prevClose   float64
	initialized bool
}

func NewRSIStrategy(period int) *RSIStrategy {
	return &RSIStrategy{
		period: period,
		gains:  make([]float64, 0, period),
		losses: make([]float64, 0, period),
	}
}

func (r *RSIStrategy) OnNewPrice(price float64) string {
	if !r.initialized {
		// 初始化前，收集数据
		if r.prevClose != 0 {
			change := price - r.prevClose
			if change > 0 {
				r.gains = append(r.gains, change)
				r.losses = append(r.losses, 0)
			} else {
				r.gains = append(r.gains, 0)
				r.losses = append(r.losses, -change)
			}
		}
		r.prevClose = price

		if len(r.gains) < r.period {
			return "" // 未收集足够数据
		}
		// 计算初始平均涨跌幅
		var sumGain, sumLoss float64
		for i := 0; i < r.period; i++ {
			sumGain += r.gains[i]
			sumLoss += r.losses[i]
		}
		r.avgGain = sumGain / float64(r.period)
		r.avgLoss = sumLoss / float64(r.period)
		r.initialized = true
		return ""
	}

	// 计算当前涨跌幅
	change := price - r.prevClose
	var gain, loss float64
	if change > 0 {
		gain = change
		loss = 0
	} else {
		gain = 0
		loss = -change
	}

	// 平滑平均计算
	r.avgGain = (r.avgGain*(float64(r.period-1)) + gain) / float64(r.period)
	r.avgLoss = (r.avgLoss*(float64(r.period-1)) + loss) / float64(r.period)

	r.prevClose = price

	// 计算RSI
	var rsi float64
	if r.avgLoss == 0 {
		rsi = 100
	} else {
		rs := r.avgGain / r.avgLoss
		rsi = 100 - (100 / (1 + rs))
	}

	// 简单买卖信号，阈值可调
	if rsi < 30 {
		return "买入"
	} else if rsi > 70 {
		return "卖出"
	}
	return ""
}
