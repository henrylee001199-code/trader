package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"trader/utils"
)

// FetchKlines 从 Binance REST 拉取历史K线
func FetchKlines(symbol, interval string, limit int) ([]utils.Kline, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=%d", symbol, interval, limit)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var raw [][]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	var klines []utils.Kline
	for _, k := range raw {
		open, _ := strconv.ParseFloat(k[1].(string), 64)
		high, _ := strconv.ParseFloat(k[2].(string), 64)
		low, _ := strconv.ParseFloat(k[3].(string), 64)
		closeP, _ := strconv.ParseFloat(k[4].(string), 64)
		volume, _ := strconv.ParseFloat(k[5].(string), 64)
		openTime := int64(0)
		switch v := k[0].(type) {
		case float64:
			openTime = int64(v)
		case int64:
			openTime = v
		}
		klines = append(klines, utils.Kline{
			Time:   time.UnixMilli(openTime),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closeP,
			Volume: volume,
		})
	}
	return klines, nil
}
