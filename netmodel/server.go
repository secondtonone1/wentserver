package netmodel

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"wentserver/config"
)

func NewServer() (*WtServer, error) {
	address := config.SERVER_IP + ":" + strconv.Itoa(config.SERVER_PORT)
	listenert, err := net.Listen(config.SERVER_TYPE, address)
	if err != nil {
		fmt.Println("listen failed !!!")
		return nil, config.ErrListenFailed
	}

	return &WtServer{listener: listenert, stopedChan: make(chan struct{}), once: &sync.Once{}}, nil
}

type WtServer struct {
	listener   net.Listener
	stopedChan chan struct{}
	once       *sync.Once
}

func (wt *WtServer) Close() {
	wt.once.Do(func() {
		if wt.listener != nil {
			defer wt.listener.Close()
		}
		//send signal to all session
		close(wt.stopedChan)
	})

}

func (wt *WtServer) acceptLoop() error {
	tcpConn, err := wt.listener.Accept()
	if err != nil {
		fmt.Println("Accept error!")
		return config.ErrAcceptFailed
	}

	newsess := NewSession(tcpConn, wt.stopedChan)
	fmt.Println("A client connected :" + tcpConn.RemoteAddr().String())
	newsess.Start()
	return nil
}

func (wt *WtServer) AcceptLoop() {

	for {
		if err := wt.acceptLoop(); err != nil {
			fmt.Println("went server accept failed!!")
			return
		}
	}

}
