package data

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"trader/utils"
)

type WSClient struct {
	conn      *websocket.Conn
	ctx       context.Context
	cancel    context.CancelFunc
	klineCh   chan KlineWithSymbol
	wg        sync.WaitGroup
	symbols   []string
	intervals []string
}

// KlineWithSymbol 封装行情数据和标识
type KlineWithSymbol struct {
	Symbol   string
	Interval string
	Kline    utils.Kline
}

// NewWSClient 创建新的WebSocket客户端
func NewWSClient(symbols []string, intervals []string) *WSClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &WSClient{
		ctx:       ctx,
		cancel:    cancel,
		klineCh:   make(chan KlineWithSymbol, 100),
		symbols:   symbols,
		intervals: intervals,
	}
}

// Start 建立连接并开始接收数据
func (c *WSClient) Start() error {
	streams := []string{}
	for _, sym := range c.symbols {
		for _, interval := range c.intervals {
			streams = append(streams, strings.ToLower(sym)+"@kline_"+interval)
		}
	}

	u := url.URL{
		Scheme:   "wss",
		Host:     "stream.binance.com:9443",
		Path:     "/stream",
		RawQuery: "streams=" + strings.Join(streams, "/"),
	}

	log.Println("connecting to", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.conn = conn

	c.wg.Add(1)
	go c.readLoop()

	return nil
}

func (c *WSClient) readLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			log.Println("ws read loop exiting")
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			time.Sleep(time.Second)
			continue
		}

		var resp struct {
			Data struct {
				E string          `json:"e"`
				S string          `json:"s"`
				K json.RawMessage `json:"k"`
			} `json:"data"`
		}

		if err := json.Unmarshal(message, &resp); err != nil {
			log.Println("json unmarshal error:", err)
			continue
		}

		if resp.Data.E != "kline" {
			continue
		}

		var k utils.Kline
		if err := json.Unmarshal(resp.Data.K, &k); err != nil {
			log.Println("kline unmarshal error:", err)
			continue
		}

		c.klineCh <- KlineWithSymbol{
			Symbol:   resp.Data.S,
			Interval: k.Interval,
			Kline:    k,
		}
	}
}

// Recv 返回行情数据通道，供外部消费
func (c *WSClient) Recv() <-chan KlineWithSymbol {
	return c.klineCh
}

// Stop 关闭连接并退出
func (c *WSClient) Stop() {
	c.cancel()
	if c.conn != nil {
		c.conn.Close()
	}
	c.wg.Wait()
	close(c.klineCh)
}
