package netmodel

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"wentmin/protocol"
)

type Session struct {
	conn         net.Conn
	closed       int32                  //session是否关闭，-1未开启，0未关闭，1关闭
	stopedChan   <-chan struct{}        //接受主协程退出通知
	protocol     protocol.ProtocolInter //字节序和自己处理器
	lock         sync.Mutex             //协程锁
	sessionGroup *sync.WaitGroup        //session的group，用于accept协程阻塞等待
}

func NewSession(connt net.Conn, stopchan <-chan struct{},
	sw *sync.WaitGroup) *Session {
	sess := &Session{
		conn:         connt,
		closed:       -1,
		stopedChan:   stopchan,
		protocol:     new(protocol.ProtocolImpl),
		sessionGroup: sw,
	}
	tcpConn := sess.conn.(*net.TCPConn)
	tcpConn.SetNoDelay(true)
	tcpConn.SetReadBuffer(64 * 1024)
	tcpConn.SetWriteBuffer(64 * 1024)
	return sess
}

func (se *Session) RawConn() *net.TCPConn {
	return se.conn.(*net.TCPConn)
}

func (se *Session) Start() {
	if atomic.CompareAndSwapInt32(&se.closed, -1, 0) {
		go se.recvLoop()
	}
}

// Close the session, destory other resource.
func (se *Session) Close() error {
	if atomic.CompareAndSwapInt32(&se.closed, 0, 1) {
		se.conn.Close()
		se.sessionGroup.Done()
	}
	return nil
}

//set read time out
//if u don't need to set read deadline, please not use it
func (se *Session) SetReadDeadline(delt time.Duration) {
	se.conn.SetReadDeadline(time.Now().Add(delt)) // timeout
}

//goroutine safe
func (se *Session) SafeSetReadDeadline(delt time.Duration) {
	se.lock.Lock()
	se.conn.SetReadDeadline(time.Now().Add(delt)) // timeout
	defer se.lock.Unlock()
}

func (se *Session) recvLoop() {
	defer se.Close()
	var packet interface{}
	var err error
	for {

		select {
		case <-se.stopedChan:
			return
		default:
			{
				packet, err = se.protocol.ReadPacket(se.conn)
				if packet == nil || err != nil {
					fmt.Println("Read packet error ", err.Error())
					return
				}

				//handle msg packet
				hdres := GetMsgHandlerIns().HandleMsgPacket(packet, se)
				if hdres != nil {
					fmt.Println(hdres.Error())
					return
				}
			}

		}

	}
}
