package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"trader/data"
	"trader/simulator"
	"trader/strategy"
	"trader/utils"
)

func main() {
	account := simulator.NewAccount(1000)
	wsURL := "wss://stream.binance.com:9443/stream?streams=" +
		"btcusdt@kline_15m/btcusdt@kline_4h/" +
		"ethusdt@kline_15m/ethusdt@kline_4h/" +
		"bnbusdt@kline_15m/bnbusdt@kline_4h"

	client, err := data.NewWSClient(wsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()
	client.Start()

	// 策略实例，orderHandler为模拟下单回调
	strat := strategy.NewSMAStrategy(5, func(symbol, side string, price float64) {
		log.Printf("模拟下单: %s %s %.4f", symbol, side, price)
		account.OnPriceUpdate(symbol, price)
	})

	go func() {
		for kline := range client.Recv() {
			strat.OnNewKline(kline.Symbol, kline.Interval, utils.Kline{
				Time:  kline.Kline.Time,
				Close: kline.Kline.Close,
			})
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	account.Close()
}
