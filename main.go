package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"trader/data"
)

func main() {
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	intervals := []string{"15m", "4h"}

	client := data.NewWSClient(symbols, intervals)
	err := client.Start()
	if err != nil {
		log.Fatal("ws start err:", err)
	}

	// 在策略里调用
	go func() {
		for k := range client.Recv() {
			log.Printf("行情: %s %s 收盘价 %.4f", k.Symbol, k.Interval, k.Kline.Close)
			// strat.OnNewKline(k.Symbol, k.Interval, k.Kline)
		}
	}()

	// 等待退出信号
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	client.Stop()
	log.Println("程序退出")
}
