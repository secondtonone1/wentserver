# wentserver

## 简介
wentserver is a network frame, it supports tcp, websocket, http comunication

wentserver是一个用golang封装的基本的网络框架，支持TCP, WEBSOCKET, HTTP类型的通信。封装并不复杂，这么做有两个目的，一是使他人可以基于此框架二次封装和改造，二是可作为基本的通信框架用于游戏，互联网等应用中。

## 源码介绍
### config介绍
config 文件夹下实现了config.go,这里配置了tcpserver和webservr的端口，地址，以及服务器错误码定义。
### httplogic 
httplogic 里面实现的都是http请求的回调函数，reghttphandler.go文件里添加消息回调函数和路由路径。回调函数单独写成go文件，比如usrinfo.go就提供了用户信息的回调函数。
``` golang
func UsrInfoReq(w http.ResponseWriter, r *http.Request) {

	//打印请求的方法

	fmt.Println("method", r.Method)

	if r.Method == "GET" {
		w.Write([]byte("server receive get method, message is hello"))
	} else {

		//否则走打印输出post接受的参数username和password

		fmt.Println(r.PostFormValue("username"))

		fmt.Println(r.PostFormValue("password"))
		fmt.Println("server receive post method, message is hello")
	}

}

func RegUsrInfo(pattern string) {

	http.HandleFunc(pattern, UsrInfoReq)
}
```
用户可以自己写自己的功能保存成XXX.go然后注册在reghttphandler.go中
``` golang
func RegHttpServerHandlers() {
    RegUsrInfo("/info")
    RegPlayGame("/playgame")
}
```
### weblogic
weblogic 里面实现的都是websocket请求的回调函数，regwebhandler.go文件里添加websocket的回调函数和路由路径。回调函数单独写成go文件，比如helloworld.go就是处理helloworld请求的，helloworld.go里提供了helloworld的回调函数。regwebhandler.go中注册这些回调函数和路径就可以了。写法和http类似，读者可以去看看例子。
### webserver
webserver.go 提供了WtWebServer的实现
``` golang
type WtWebServer struct {
}

func (wb *WtWebServer) RegWebHandler() {
	weblogic.RegWebServerHandlers()
	httplogic.RegHttpServerHandlers()
}

func (wb *WtWebServer) ListenAndServe() error {
	address := config.SERVER_IP + ":" + strconv.Itoa(config.WEBSERVER_PORT)
	err := http.ListenAndServe(address, nil)
	return err
}

func (wb *WtWebServer) Start() {
	wb.RegWebHandler()
	err := wb.ListenAndServe()
	if err != nil {
		fmt.Println(config.ErrWebListenFailed.Error())
		return
	}
}

func NewWtWebServer() *WtWebServer {
	return &WtWebServer{}
}
```
RegWebHandler可以注册websocket和http两种协议的请求。
ListenAndServe用来监听和处理请求。这些统统封装在Start中。所以用户只要调用wtserver.Start()就可以完成服务器启动和监听处理请求。
### logic 
logic里主要是TCP请求的回调函数，比如处理helloworld请求就将回调函数写在helloworld.go中。处理oncehello请求就写在oncehello.go中。
然后通过reghandler.go注册消息。
packetid.go 里填写自己的消息id，每当需要添加新的请求和逻辑处理时就定义一个消息id填在packetid.go中，将回调函数成单独的go文件，然后在reghandler.go中注册即可。以下是使用例子
packetid 定义
``` golang
const (
	CLOSECLT_NOTIFY = 0
	HELLOWORLD_REQ  = 1
	HELLOWORLD_RSP  = 2
	ONCEHELLO_REQ   = 3
	ONCEHELLO_RESP  = 4
)
```
然后分别实现对应的请求处理，举例Helloworld的请求，回调函数写在helloworld.go中。
``` golang
func RegHelloWorldReq() {
	var HelloworldReq netmodel.CallBackFunc = func(se interface{}, param interface{}) error {
		msgpacket, ok := param.(*protocol.MsgPacket)
		if !ok {
			return config.ErrTypeAssertain
		}

		session, ok := se.(*netmodel.Session)
		if !ok {
			return config.ErrTypeAssertain
		}
		fmt.Println("Server recieve from ", session.RawConn().RemoteAddr().String())
		fmt.Println("Server Recv Msg is ", string(msgpacket.Body.Data))
		helloworldrsp := new(protocol.MsgPacket)
		helloworldrsp.Head.Id = HELLOWORLD_RSP
		helloworldrsp.Head.Len = uint16(len("server recive msg hello world!"))
		helloworldrsp.Body.Data = []byte("server recive msg hello world!")
		err := session.AsyncSend(helloworldrsp)
		if err != nil {
			fmt.Println("Handle Msg HelloworldReq failed")
			return config.ErrHelloWorldReqFailed
		}
		return nil
	}

	netmodel.MsgHandler.RegMsgHandler(HELLOWORLD_REQ, HelloworldReq)
}
```
然后再reghandler.go中注册消息
``` golang
func RegServerHandlers() {
	RegHelloWorldReq()
	RegOnceHelloReq()
}
```
这就是消息注册和处理的流程。
notifyclose.go是个特殊的处理功能，就是服务器发现对方长时间未通信要主动断开连接，所以先发送消息通知对方断开，然后再主动断开
其实这是我再三考虑的一个问题，这种场景在游戏和互联网领域都需要，也就是心跳保活机制，对方如果长时间不通信就必须断开，防止过多
的僵尸链接。这种断开是从logic层断开，比较合理。之前设计的处理方式为收取信息超时检测，如果长时间未收取就会返回错误并断开连接。
但是考虑这种检测是在session层面，而且存在多goroutine访问一个connection的情况，为避免安全隐患和效率问题，就舍弃了该做法。
但是session层仍保留了该超时检测的api，读者可以用，而且提供了goroutine安全模式和不安全(不加锁)模式。
###protocol
protocol主要是提供了字节读写和消息包的设计。
消息设计如下
``` golang
/*
-----------------------------------------------
               msgpacket
-----------------------------------------------
      msghead     |  msgbody
-----------------------------------------------
id      |   len   |   data
-----------------------------------------------
*/
type MsgHead struct {
	Id  uint16
	Len uint16
}
type MsgBody struct {
	Data []byte
}
type MsgPacket struct {
	Head MsgHead
	Body MsgBody
}
```
一个消息由MsgHead和MsgBody构成, MsgHead包括2字节Id, 2字节Len。MsgData包括[]byte类型的Data。
protocol.go提供了解析消息，读取消息，构造消息的功能。其实就是提供了TCP字节流和MsgPacket之间的转换。
stream.go实现了大端小端读写。
### netmodel
netmodel提供了网络层的基本功能。
client.go 提供了客户端连接和收发消息的api
server.go 提供了服务器处理消息和监听消息的api
msghandler.go提供了消息注册和处理的底层api
在msghandler.go中MsgHandler为一个单例对象，所以多goroutine访问存在安全问题，因此提供了goroutine安全的注册和处理函数。
但是一个优秀的设计不应该是为了使用者的安全稳定而权衡了性能，我的设计思路是应该极大的信任开发者。尽可能规避锁和chan的使用。我使
用了不加锁的消息注册和处理函数。前提是这些操作只在一个goroutine中完成。所以回顾server.go，我在主线程里完成了消息注册。
然后消息处理留给一个goroutine来处理，这样保证了安全性，也节省了锁的开销。
主线程中调用logic注册消息
``` golang
func main() {
	logic.RegServerHandlers()
	wt, err := netmodel.NewServer()
	if err != nil {
		return
	}
	defer wt.Close()
	wt.AcceptLoop()
}
```
在专门负责接收信息的goroutine中处理消息。
``` golang
func (se *Session) recvLoop() {
	defer se.Close()
	defer func() {
		fmt.Println("recv goroutine exit!")
	}()
	var packet interface{}
	var err error
	for {

		select {
		case <-se.stopedChan:
			return
		case <-se.asyncStop:
			return
		default:
			{
				packet, err = se.protocol.ReadPacket(se.conn)
				if packet == nil || err != nil {
					fmt.Println("Read packet error ", err.Error())
					return
				}

				//handle msg packet
				hdres := MsgHandler.HandleMsgPacket(packet, se)
				if hdres != nil {
					fmt.Println(hdres.Error())
					return
				}
			}

		}

	}
}
```
session.go
session.go就是网络层的会话功能，主要提供了消息收发的功能。一个session对应两个goroutine，一个goroutine用来接收消息，然后处理消息，
同时支持异步发送消息。另一个goroutine只用来发送，不做任何消息处理。我也建议使用者在发送goroutine中不要附加session的访问和处理。
因为两个goroutine同时处理一个session可能会造成隐患。
session的基本结构
``` golang
type Session struct {
	conn           net.Conn
	closed         int32
	stopedChan     <-chan struct{}
	protocol       protocol.ProtocolInter
	asyncStop      chan struct{}
	sendChan       chan interface{}
	lock           sync.Mutex
	sendChanClosed int32
}
```
conn为accept返回的连接。
closed表示连接是否关闭
asyncStop 用来在发送协程和接收协程中同步信号，一方通知另一方协程退出。
stopedChan 是外层server通知所有协程退出的信号
protocol  是protocol包的协议解析和处理对象。
sendChan  是用来在接收协程和发送协程之间同步数据的chan，里面存储了要发送给客户端的数据。
接收协程通过调用Asyncsend，由接收协程填入消息，然后发送协程取出消息并发送给客户端。
lock 是互斥锁，为了权衡一部分用户的互斥操作。
sendChanClosed 表示sendChan是否关闭。理论上接收协程和发送协程都退出后，sendChan会自动被回收，
我手动回收罢了。

Start()函数开辟发送协程和接收协程
``` golang
func (se *Session) Start() {
	if atomic.CompareAndSwapInt32(&se.closed, -1, 0) {
		atomic.CompareAndSwapInt32(&se.sendChanClosed, -1, 0)
		go se.sendLoop()
		go se.recvLoop()
	}
}
```
两个原子操作关闭
``` golang
// Close the session, destory other resource.
func (se *Session) Close() error {
	if atomic.CompareAndSwapInt32(&se.closed, 0, 1) {
		se.conn.Close()
		close(se.asyncStop)
	}
	return nil
}

func (se *Session) CloseSendChan() error {
	if atomic.CompareAndSwapInt32(&se.sendChanClosed, 0, 1) {
		close(se.sendChan)
		fmt.Println("send goroutine exit!")
	}
	return nil
}
```
提供了安全和不加锁两种模式的超时设置
``` golang
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
```
这两个函数最初是用来检测对方是否长时间不发送消息的，后来我不使用这个功能，将超时检测放到logic层面
下面是session的发送协程的功能
``` golang
func (se *Session) sendLoop() {
	defer se.Close()
	for {
		select {
		case <-se.stopedChan:
			return
		case <-se.asyncStop:
			return
		case packet, ok := <-se.sendChan:
			{
				if !ok {
					return
				}
				if packet == nil {
					return
				}
				err := se.protocol.WritePacket(se.conn, packet)
				if err != nil {
					return
				}
			}
		}
	}
}
```
发送协程随机选择可执行的条件执行，如果有关闭信号，则退出并Close。或者从sendChan中读取数据发送给客户端。
这里要提出一点，就是即使接受协程关闭，发送协程得到消息，但是随机选择了sendChan分支也没关系，因为该分支也会检测
连接的有效性。
下面是接收协程的功能
``` golang
func (se *Session) recvLoop() {
	defer se.Close()
	defer se.CloseSendChan()
	var packet interface{}
	var err error
	for {

		select {
		case <-se.stopedChan:
			return
		case <-se.asyncStop:
			return
		default:
			{
				packet, err = se.protocol.ReadPacket(se.conn)
				if packet == nil || err != nil {
					fmt.Println("Read packet error ", err.Error())
					return
				}

				//handle msg packet
				hdres := MsgHandler.HandleMsgPacket(packet, se)
				if hdres != nil {
					fmt.Println(hdres.Error())
					return
				}
			}

		}

	}
}
```
接受协程检测没有关闭信号，就会等待接受客户端发送数据。考虑到sendChan在接收协程中会被写入数据，golang中如果对一个关闭的协程执行写操作，会导致panic，所以sendChan的关闭放在接收协程中。另外存在这样一种情况，就是接收协成阻塞在ReadPacket中，而发送协程关闭退出了，那么接收协成也会返回，因为connection会在接收协程退出时关闭，而发送协程会检测到此关闭，从ReadPacket中返回。

异步发送函数AsyncSend()
``` golang
func (se *Session) AsyncSend(packet interface{}) error {
	select {
	case <-se.asyncStop:
		return config.ErrAsyncSendStop
	case <-se.stopedChan:
		return config.ErrAsyncSendStop
	default:
		if packet == nil {
			se.Close()
			return nil
		}
		se.sendChan <- packet
		return nil
	}
}
```
该异步发送函数首先检测连接是否关闭，然后在将packet写入sendChan。异步发送函数必须在接收协程中调用。
## 使用案例
### 客户端单线程
单线程100次收发
``` golang
package main

import (
	"fmt"
	"wentby/netmodel"
	"wentby/protocol"
)

func main() {
	cs, err := netmodel.Dial("tcp4", "127.0.0.1:10006")
	if err != nil {
		return
	}
	var i int16
	for i = 0; i < 100; i++ {
		packet := new(protocol.MsgPacket)
		packet.Head.Id = 1
		packet.Head.Len = 5
		packet.Body.Data = []byte("Hello")
		cs.Send(packet)
		packetrsp, err := cs.Recv()
		if err != nil {
			fmt.Println("receive error")
			return
		}

		datarsp := packetrsp.(*protocol.MsgPacket)
		fmt.Println("packet id is", datarsp.Head.Id)
		fmt.Println("packet len is", datarsp.Head.Len)
		fmt.Println("packet data is", string(datarsp.Body.Data))
		time.Sleep(time.Millisecond * time.Duration(10))
	}
	fmt.Println("circle times are ", i)
}
```
### 服务端
``` golang
package main

import (
	"wentby/logic"
	"wentby/netmodel"
)

func main() {
	logic.RegServerHandlers()
	wt, err := netmodel.NewServer()
	if err != nil {
		return
	}
	defer wt.Close()
	wt.AcceptLoop()
}
```

### 客户端多goroutine
``` golang
func main() {
	for i := 0; i < 100; i++ {
		go func() {
			cs, err := netmodel.Dial("tcp4", "127.0.0.1:10006")
			if err != nil {
				return
			}
			var i int16
			for i = 0; i < 100; i++ {
				packet := new(protocol.MsgPacket)
				packet.Head.Id = 1
				packet.Head.Len = 5
				packet.Body.Data = []byte("Hello")
				cs.Send(packet)
				packetrsp, err := cs.Recv()
				if err != nil {
					fmt.Println("receive error")
					return
				}

				datarsp := packetrsp.(*protocol.MsgPacket)
				fmt.Println("packet id is", datarsp.Head.Id)
				fmt.Println("packet len is", datarsp.Head.Len)
				fmt.Println("packet data is", string(datarsp.Body.Data))
				time.Sleep(time.Millisecond * time.Duration(10))
			}
			fmt.Println("circle times are ", i)
		}()

		time.Sleep(time.Second * time.Duration(1))
	}
	for {
		time.Sleep(time.Second * time.Duration(15))
	}
}
```
服务器同上就可以

### http客户端
``` golang
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func httpGet() {
	resp, err := http.Get("http://localhost:9998/info")
	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	fmt.Println(string(body))
}

func httpPost() {
	resp, err := http.Post("http://localhost:9998/info?username=Jenny&password=12345",
		"application/x-www-form-urlencoded",
		strings.NewReader("name=zack"))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	fmt.Println(string(body))
}

func httpPostForm() {
	resp, err := http.PostForm("http://localhost:9998/info",
		url.Values{"username": {"Zack"}, "password": {"123"}})

	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	fmt.Println(string(body))

}

func main() {
	httpGet()
	httpPost()
	httpPostForm()
}

```
### http服务端
``` golang
package main

import "wentby/webserver"

func main() {
	webserver := webserver.NewWtWebServer()
	webserver.Start()
}
```

### 客户端websocket
``` golang
package main

import (
	"fmt"
	"wentby/config"

	"golang.org/x/net/websocket"
)

var url = "ws://127.0.0.1:9998/"
var origin = "http://127.0.0.1:9998/"

func main() {
	conn, err := websocket.Dial(url, "", origin)
	if err != nil {
		fmt.Println(config.ErrWebSocketDail.Error())
	}

	_, err = conn.Write([]byte("Hello !"))
	if err != nil {
		fmt.Println(config.ErrWebSocketWrite)
		return
	}
	readdata := make([]byte, 1024)
	readlen, err := conn.Read(readdata)
	if err != nil {
		fmt.Println(config.ErrWebSocketRead.Error())
		return
	}

	fmt.Println("client recieve msg is : ", string(readdata[:readlen]))
}
```
### websocket服务器
``` golang
package main

import "wentby/webserver"

func main() {
	webserver := webserver.NewWtWebServer()
	webserver.Start()
}
```


个人公众号

![https://github.com/secondtonone1/blogsbackup/blob/master/blogs/source/_posts/golang01/wxgzh.jpg](https://github.com/secondtonone1/blogsbackup/blob/master/blogs/source/_posts/golang01/wxgzh.jpg)
