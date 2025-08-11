package simulator

import (
	"log"
	"sync"
)

type Account struct {
	mu        sync.Mutex
	Balance   float64
	Positions map[string]float64 // symbol -> 持仓数量
}

func NewAccount(balance float64) *Account {
	return &Account{
		Balance:   balance,
		Positions: make(map[string]float64),
	}
}

// 模拟下单，简单买卖，更新持仓和资金（不考虑手续费滑点）
func (a *Account) OnOrder(symbol string, side string, price float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch side {
	case "buy":
		a.Positions[symbol] += 1
		a.Balance -= price
		log.Printf("买入: %s 单价 %.4f, 当前持仓 %.2f, 余额 %.2f", symbol, price, a.Positions[symbol], a.Balance)
	case "sell":
		if a.Positions[symbol] > 0 {
			a.Positions[symbol] -= 1
			a.Balance += price
			log.Printf("卖出: %s 单价 %.4f, 当前持仓 %.2f, 余额 %.2f", symbol, price, a.Positions[symbol], a.Balance)
		}
	default:
		log.Printf("未知方向: %s", side)
	}
}
