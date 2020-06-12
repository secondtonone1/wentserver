package netmodel

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"wentmin/common"
	"wentmin/components"
)

func NewTcpServer() (*WtServer, error) {
	address := "0.0.0.0:" + strconv.Itoa(components.ServerPort)
	listenert, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("listen failed !!!")
		return nil, common.ErrListenFailed
	}

	return &WtServer{listener: listenert, stopedChan: make(chan struct{}),
		once: &sync.Once{}, sessionGroup: &sync.WaitGroup{}, notifyMain: make(chan struct{})}, nil
}

type WtServer struct {
	listener     net.Listener
	stopedChan   chan struct{} //通知session关闭
	once         *sync.Once
	sessionGroup *sync.WaitGroup
	notifyMain   chan struct{}
}

//主协程主动关闭accept
func (wt *WtServer) Close() {
	wt.once.Do(func() {
		if wt.listener != nil {
			defer wt.listener.Close()
		}
	})
}

func (wt *WtServer) acceptLoop() error {

	tcpConn, err := wt.listener.Accept()
	if err != nil {
		fmt.Println("Accept error!, err is ", err.Error())
		return common.ErrAcceptFailed
	}

	newsess := NewSession(tcpConn, wt.stopedChan, wt.sessionGroup)
	fmt.Println("A client connected :" + tcpConn.RemoteAddr().String())
	newsess.Start()
	wt.sessionGroup.Add(1)
	return nil

}

func (wt *WtServer) AcceptLoop() {
	defer func() {
		fmt.Println("main io goroutin exit ")
		if err := recover(); err != nil {
			fmt.Println("server recover from err , err is ", err)
		}
		close(wt.stopedChan)
		wt.sessionGroup.Wait()
		close(wt.notifyMain)
	}()
	for {
		if err := wt.acceptLoop(); err != nil {
			fmt.Println("went server accept failed!! ")
			return
		}
	}
}

func (wt *WtServer) WaitClose() {
	_, ok := <-wt.notifyMain
	if !ok {
		fmt.Println("wt server closed successfully ")
	}
}
