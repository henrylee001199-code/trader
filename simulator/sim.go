package simulator

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"

	"trader/utils"
)

type Position struct {
	Symbol     string
	Size       float64 // base asset quantity, e.g., BTC amount
	EntryPrice float64
	StopLoss   float64
	Direction  int // 1=long, -1=short (spot we mostly use long)
	OpenedAt   time.Time
}

type Account struct {
	mu         sync.Mutex // 保护账户状态（Cash、Positions、LastPrices）
	fileMu     sync.Mutex // 保护文件写入（trade/equity csv）
	Initial    float64
	Cash       float64 // free USD
	Positions  map[string]*Position
	LastPrices map[string]float64

	tradeLog  *csv.Writer
	equityLog *csv.Writer
	logFile   *os.File
	eqFile    *os.File
}

func NewAccount(initEquity float64) *Account {
	a := &Account{
		Initial:    initEquity,
		Cash:       initEquity,
		Positions:  make(map[string]*Position),
		LastPrices: make(map[string]float64),
	}

	tf, err := os.OpenFile("trades.csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	a.logFile = tf
	a.tradeLog = csv.NewWriter(tf)
	if stat, _ := tf.Stat(); stat.Size() == 0 {
		a.tradeLog.Write([]string{"time", "event", "symbol", "direction", "price", "size", "pl_usd", "cash_after"})
		a.tradeLog.Flush()
	}

	ef, err := os.OpenFile("equity.csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	a.eqFile = ef
	a.equityLog = csv.NewWriter(ef)
	if stat, _ := ef.Stat(); stat.Size() == 0 {
		a.equityLog.Write([]string{"time", "equity", "cash", "unrealized_pnl"})
		a.equityLog.Flush()
	}

	return a
}

func (a *Account) Close() {
	// flush and close files safely
	a.fileMu.Lock()
	if a.tradeLog != nil {
		a.tradeLog.Flush()
	}
	if a.logFile != nil {
		_ = a.logFile.Close()
	}
	if a.equityLog != nil {
		a.equityLog.Flush()
	}
	if a.eqFile != nil {
		_ = a.eqFile.Close()
	}
	a.fileMu.Unlock()
}

func (a *Account) Execute(sig *utils.Signal) {
	if sig == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.Positions[sig.Symbol]; ok {
		// 已有持仓，当前简单策略不加仓
		return
	}

	requiredUSD := sig.EntryPrice * sig.Size
	if requiredUSD > a.Cash {
		// 资金不足，记录拒单（使用文件锁）
		a.writeTradeLog("reject_insufficient_cash", sig.Symbol, sig.Direction, sig.EntryPrice, sig.Size, 0)
		return
	}

	// 扣除现金并建立仓位
	a.Cash -= requiredUSD
	pos := &Position{
		Symbol:     sig.Symbol,
		Size:       sig.Size,
		EntryPrice: sig.EntryPrice,
		StopLoss:   sig.StopLoss,
		Direction:  sig.Direction,
		OpenedAt:   sig.Time,
	}
	a.Positions[sig.Symbol] = pos

	a.writeTradeLog("open", sig.Symbol, sig.Direction, sig.EntryPrice, sig.Size, 0)
	// 写净值日志（在持锁环境下，writeEquityLogLocked 会使用 fileMu 内部）
	a.writeEquityLogLocked()
	fmt.Printf("[SIM OPEN] %s size=%.6f @ %.2f | SL=%.2f | Cash=%.2f\n",
		sig.Symbol, sig.Size, sig.EntryPrice, sig.StopLoss, a.Cash)
}

func (a *Account) OnPriceUpdate(symbol string, price float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.LastPrices[symbol] = price

	// 检查止损（spot 多头逻辑）
	for sym, pos := range a.Positions {
		if pos.Direction == 1 {
			if price <= pos.StopLoss {
				a.closePositionLocked(sym, price, "stop_loss")
			}
		} else if pos.Direction == -1 {
			if price >= pos.StopLoss {
				a.closePositionLocked(sym, price, "stop_loss")
			}
		}
	}

	// 写净值日志
	a.writeEquityLogLocked()
}

func (a *Account) closePositionLocked(symbol string, exitPrice float64, reason string) {
	pos, ok := a.Positions[symbol]
	if !ok {
		return
	}
	pl := (exitPrice - pos.EntryPrice) * pos.Size * float64(pos.Direction)
	// 回补现价*size 到现金
	a.Cash += exitPrice * pos.Size
	delete(a.Positions, symbol)

	a.writeTradeLog(fmt.Sprintf("close_%s", reason), symbol, pos.Direction, exitPrice, pos.Size, pl)
	fmt.Printf("[SIM CLOSE] %s closed @ %.2f | size=%.6f | P&L=%.2f | Cash=%.2f\n",
		symbol, exitPrice, pos.Size, pl, a.Cash)
}

func (a *Account) writeTradeLog(event, symbol string, direction int, price, size, pl float64) {
	a.fileMu.Lock()
	defer a.fileMu.Unlock()

	if a.tradeLog == nil {
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	dir := "LONG"
	if direction < 0 {
		dir = "SHORT"
	}
	record := []string{
		now,
		event,
		symbol,
		dir,
		fmt.Sprintf("%.8f", price),
		fmt.Sprintf("%.8f", size),
		fmt.Sprintf("%.8f", pl),
		fmt.Sprintf("%.8f", a.Cash),
	}
	_ = a.tradeLog.Write(record)
	a.tradeLog.Flush()
}

func (a *Account) writeEquityLogLocked() {
	// caller must hold a.mu
	a.fileMu.Lock()
	defer a.fileMu.Unlock()

	if a.equityLog == nil {
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	unreal := a.computeUnrealizedLocked()
	eq := a.Cash + unreal
	rec := []string{
		now,
		fmt.Sprintf("%.8f", eq),
		fmt.Sprintf("%.8f", a.Cash),
		fmt.Sprintf("%.8f", unreal),
	}
	_ = a.equityLog.Write(rec)
	a.equityLog.Flush()
}

func (a *Account) computeUnrealizedLocked() float64 {
	// caller must hold a.mu
	unreal := 0.0
	for sym, pos := range a.Positions {
		price, ok := a.LastPrices[sym]
		if !ok || price == 0 {
			continue
		}
		unreal += (price - pos.EntryPrice) * pos.Size * float64(pos.Direction)
	}
	return unreal
}

func (a *Account) GetEquity() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	unreal := a.computeUnrealizedLocked()
	return a.Cash + unreal
}
