package data

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"sync"
	"trader/strategy"
)

// Binance WebSocket Kline事件结构体，价格等字段用json.Number兼容数字或字符串
type KlineEvent struct {
	EventType string    `json:"e"` // 事件类型
	EventTime int64     `json:"E"` // 事件时间，毫秒
	Symbol    string    `json:"s"` // 交易对
	Kline     KlineData `json:"k"` // K线详细数据
}

type KlineData struct {
	StartTime           int64       `json:"t"`
	CloseTime           int64       `json:"T"`
	Symbol              string      `json:"s"`
	Interval            string      `json:"i"`
	FirstTradeID        int64       `json:"f"`
	LastTradeID         int64       `json:"L"`
	Open                json.Number `json:"o"`
	Close               json.Number `json:"c"`
	High                json.Number `json:"h"`
	Low                 json.Number `json:"l"`
	Volume              json.Number `json:"v"`
	NumberOfTrades      int64       `json:"n"`
	IsKlineClosed       bool        `json:"x"`
	QuoteVolume         json.Number `json:"q"`
	TakerBuyBaseVolume  json.Number `json:"V"`
	TakerBuyQuoteVolume json.Number `json:"Q"`
	Ignore              string      `json:"B"`
}

type WSClient struct {
	conn   *websocket.Conn
	stopCh chan struct{}

	mu         sync.Mutex
	strategies map[string]strategy.Strategy // 改成接口 Strategy
}

// NewWSClient 构造函数，初始化不同周期对应的复合策略
func NewWSClient() *WSClient {
	simpleMA15m := strategy.NewSimpleMA(5, 20)
	rsi15m := strategy.NewRSIStrategy(14)
	composite15m := strategy.NewCompositeStrategy(simpleMA15m, rsi15m)

	simpleMA4h := strategy.NewSimpleMA(5, 20)
	rsi4h := strategy.NewRSIStrategy(14)
	composite4h := strategy.NewCompositeStrategy(simpleMA4h, rsi4h)

	return &WSClient{
		stopCh: make(chan struct{}),
		strategies: map[string]strategy.Strategy{
			"15m": composite15m,
			"4h":  composite4h,
		},
	}
}

func (c *WSClient) Start() error {
	streams := "btcusdt@kline_15m/btcusdt@kline_4h/ethusdt@kline_15m/ethusdt@kline_4h/bnbusdt@kline_15m/bnbusdt@kline_4h"
	u := url.URL{
		Scheme:   "wss",
		Host:     "stream.binance.com:9443",
		Path:     "/stream",
		RawQuery: "streams=" + streams,
	}

	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	log.Printf("connected to %s", u.String())

	go c.readLoop()
	return nil
}

func (c *WSClient) readLoop() {
	defer c.conn.Close()

	for {
		select {
		case <-c.stopCh:
			return
		default:
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("read error: %v", err)
				return
			}

			c.handleMessage(msg)
		}
	}
}

func (c *WSClient) handleMessage(msg []byte) {
	// 外层结构 Binance WebSocket推送格式： { "stream": "...", "data": { ... } }
	var wrapper struct {
		Stream string          `json:"stream"`
		Data   json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(msg, &wrapper); err != nil {
		log.Printf("json unmarshal wrapper error: %v", err)
		return
	}

	var event KlineEvent
	if err := json.Unmarshal(wrapper.Data, &event); err != nil {
		log.Printf("json unmarshal inner error: %v", err)
		return
	}

	closePrice, err := event.Kline.Close.Float64()
	if err != nil {
		log.Printf("parse close price error: %v", err)
		return
	}

	log.Printf("行情: %s %s 收盘价 %.4f %v", event.Symbol, event.Kline.Interval, closePrice, event.Kline.IsKlineClosed)

	if event.Kline.IsKlineClosed {
		c.mu.Lock()
		strat, ok := c.strategies[event.Kline.Interval]
		c.mu.Unlock()

		if !ok {
			// 未知周期，忽略
			return
		}

		signal := strat.OnNewPrice(closePrice)
		if signal != "" {
			log.Printf("策略信号：%s %s %s %.4f", event.Symbol, event.Kline.Interval, signal, closePrice)
			// 这里扩展调用模拟或实盘下单
			c.simulateOrder(event.Symbol, signal, closePrice)
		}
	}
}

func (c *WSClient) simulateOrder(symbol, action string, price float64) {
	log.Printf("模拟下单: %s %s %.4f", symbol, action, price)
	// TODO: 这里可以调用真实下单接口或其它逻辑
}

func (c *WSClient) Stop() {
	close(c.stopCh)
	if c.conn != nil {
		c.conn.Close()
	}
}
