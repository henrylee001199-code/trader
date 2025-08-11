package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"trader/data"
)

func main() {
	client := data.NewWSClient()
	go client.Start()

	// 等待系统信号优雅退出
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	client.Stop()
}
