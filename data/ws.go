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

// StringOrFloat64 支持字符串或数字类型的JSON字段解析
type StringOrFloat64 float64

func (s *StringOrFloat64) UnmarshalJSON(data []byte) error {
	// 尝试解析成 float64
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*s = StringOrFloat64(f)
		return nil
	}
	// 尝试解析成字符串再转 float64
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}
	*s = StringOrFloat64(f)
	return nil
}

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

func (w *WSClient) Recv() <-chan utils.KlineWithSymbol {
	return w.recv
}

type rawStream struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

type klineInner struct {
	StartTime int64           `json:"t"`
	CloseTime int64           `json:"T"`
	Symbol    string          `json:"s"`
	Interval  string          `json:"i"`
	Open      StringOrFloat64 `json:"o"`
	Close     StringOrFloat64 `json:"c"`
	High      StringOrFloat64 `json:"h"`
	Low       StringOrFloat64 `json:"l"`
	Volume    StringOrFloat64 `json:"v"`
	IsFinal   bool            `json:"x"`
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
			EvtType string     `json:"e"`
			E       int64      `json:"E"`
			K       klineInner `json:"k"`
		}
		if err := json.Unmarshal(rs.Data, &inner); err != nil {
			log.Println("unmarshal inner err:", err)
			continue
		}

		if inner.EvtType != "kline" {
			continue
		}

		k := inner.K
		kline := utils.Kline{
			Time:   time.UnixMilli(k.StartTime),
			Open:   float64(k.Open),
			High:   float64(k.High),
			Low:    float64(k.Low),
			Close:  float64(k.Close),
			Volume: float64(k.Volume),
		}

		select {
		case w.recv <- utils.KlineWithSymbol{
			Symbol:   strings.ToUpper(k.Symbol),
			Interval: k.Interval,
			Kline:    kline,
		}:
		case <-w.done:
			return
		}
	}
}
