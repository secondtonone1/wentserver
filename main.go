package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wentmin/components"

	"wentmin/netmodel"

	"github.com/astaxie/beego/logs"
)

func main() {
	logs.Debug("server port is %d", components.ServerPort)
	wt, err := netmodel.NewTcpServer()
	if err != nil {
		panic("new tcp server failed")
	}
	go wt.AcceptLoop()
	stopsignal := make(chan os.Signal) // 接收系统中断信号
	var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGINT}
	signal.Notify(stopsignal, shutdownSignals...)
	select {
	case sign := <-stopsignal:
		fmt.Println("catch stop signal, ", sign)
		wt.Close()
	}
	wt.WaitClose()
	time.Sleep(time.Second * 5)
}
