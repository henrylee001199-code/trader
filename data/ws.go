package data

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"time"
	"trader/strategy"
)

// Binance WebSocket Kline事件结构体，字段价格、成交量用json.Number兼容数字或字符串
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
}

func NewWSClient() *WSClient {
	return &WSClient{
		stopCh: make(chan struct{}),
	}
}

func (c *WSClient) Start() error {
	// 示例订阅BTC/ETH/BNB 15m和4h K线
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

	// 打印示例：交易对、周期、收盘价、是否收盘
	closePrice, _ := event.Kline.Close.Float64()
	log.Printf("行情: %s %s 收盘价 %.4f %s %v", event.Symbol, event.Kline.Interval, closePrice, time.UnixMilli(event.Kline.CloseTime).Format(time.RFC3339), event.Kline.IsKlineClosed)
}

func (c *WSClient) Stop() {
	close(c.stopCh)
	if c.conn != nil {
		c.conn.Close()
	}
}

// 全局策略实例，举例15m和4h两个周期
var (
	simpleMA15m = strategy.NewSimpleMA(5, 20)
	simpleMA4h  = strategy.NewSimpleMA(5, 20)
)

func handleMessage(msg []byte) {
	var event KlineEvent
	err := json.Unmarshal(msg, &event)
	if err != nil {
		log.Printf("json unmarshal error: %v", err)
		return
	}

	// 转换收盘价字符串为float64
	closePrice, err := event.Kline.Close.Float64()
	if err != nil {
		log.Printf("parse close price error: %v", err)
		return
	}

	log.Printf("行情: %s %s 收盘价 %.4f %v", event.Symbol, event.Kline.Interval, closePrice, event.Kline.IsKlineClosed)

	if event.Kline.IsKlineClosed {
		var signal string
		switch event.Kline.Interval {
		case "15m":
			signal = simpleMA15m.OnNewPrice(closePrice)
		case "4h":
			signal = simpleMA4h.OnNewPrice(closePrice)
		}

		if signal != "" {
			log.Printf("策略信号：%s %s %s %.4f", event.Symbol, event.Kline.Interval, signal, closePrice)
			// 这里可以接入模拟下单或实盘下单逻辑
		}
	}
}
