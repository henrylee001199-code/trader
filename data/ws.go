package data

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"trader/utils"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	conn *websocket.Conn
	recv chan utils.KlineWithSymbol

	done chan struct{}
	wg   sync.WaitGroup
}

func NewWSClient(url string) (*WSClient, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	client := &WSClient{
		conn: c,
		recv: make(chan utils.KlineWithSymbol, 100),
		done: make(chan struct{}),
	}
	return client, nil
}

func (w *WSClient) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.readLoop()
	}()
}

func (w *WSClient) Stop() {
	close(w.done)
	_ = w.conn.Close()
	w.wg.Wait()
	close(w.recv)
}

// 公开获取行情数据通道
func (w *WSClient) Recv() <-chan utils.KlineWithSymbol {
	return w.recv
}

type rawStream struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

type klineInner struct {
	StartTime int64  `json:"t"` // 开盘时间
	CloseTime int64  `json:"T"` // 收盘时间
	Symbol    string `json:"s"`
	Interval  string `json:"i"`
	Open      string `json:"o"`
	Close     string `json:"c"`
	High      string `json:"h"`
	Low       string `json:"l"`
	Volume    string `json:"v"`
	IsFinal   bool   `json:"x"`
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func (w *WSClient) readLoop() {
	for {
		select {
		case <-w.done:
			return
		default:
		}

		_, msg, err := w.conn.ReadMessage()
		if err != nil {
			log.Println("read websocket message err:", err)
			return
		}

		var rs rawStream
		if err := json.Unmarshal(msg, &rs); err != nil {
			log.Println("unmarshal raw err:", err)
			continue
		}

		var inner struct {
			E json.Number `json:"E"`
			K klineInner  `json:"k"`
		}
		if err := json.Unmarshal(rs.Data, &inner); err != nil {
			log.Println("unmarshal inner err:", err)
			continue
		}

		_, err = inner.E.Int64()
		if err != nil {
			log.Println("parse event time failed:", err)
			continue
		}

		k := inner.K
		open := parseFloat(k.Open)
		high := parseFloat(k.High)
		low := parseFloat(k.Low)
		closeP := parseFloat(k.Close)
		vol := parseFloat(k.Volume)
		sym := strings.ToUpper(k.Symbol)
		interval := k.Interval

		kline := utils.Kline{
			Time:   time.UnixMilli(k.StartTime),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closeP,
			Volume: vol,
		}

		select {
		case w.recv <- utils.KlineWithSymbol{
			Symbol:   sym,
			Interval: interval,
			Kline:    kline,
		}:
		case <-w.done:
			return
		}
	}
}
