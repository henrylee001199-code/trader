package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"trader/data"
	"trader/simulator"
)

func main() {
	account := simulator.NewAccount(1000)

	wsURL := "wss://stream.binance.com:9443/stream?streams=" +
		"btcusdt@kline_15m/btcusdt@kline_4h/" +
		"ethusdt@kline_15m/ethusdt@kline_4h/" +
		"bnbusdt@kline_15m/bnbusdt@kline_4h"

	client, err := data.NewWSClient(wsURL)
	if err != nil {
		log.Fatal("failed to connect ws:", err)
	}
	defer client.Stop()

	client.Start()

	go func() {
		for kline := range client.Recv() {
			// 示例调用账户更新价格
			account.OnPriceUpdate(kline.Symbol, kline.Kline.Close)
			// 这里可以接入你的策略信号处理逻辑
		}
	}()

	// 等待系统信号优雅退出
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	account.Close()
}
