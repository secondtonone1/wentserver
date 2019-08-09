package logic

import (
	"fmt"
	"wentby/config"
	"wentby/netmodel"
	"wentby/protocol"
)

func RegOnceHelloReq() {
	var OnceHello netmodel.CallBackFunc = func(se interface{}, param interface{}) error {
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
		helloworldrsp.Head.Id = ONCEHELLO_RESP
		helloworldrsp.Head.Len = uint16(len("server recive once hello"))
		helloworldrsp.Body.Data = []byte("server recive once hello")
		err := session.AsyncSend(helloworldrsp)
		if err != nil {
			fmt.Println("Handle Msg HelloworldReq failed")
			return config.ErrHelloWorldReqFailed
		}

		NotifyClientClose(se, nil)
		return nil
	}

	netmodel.MsgHandler.RegMsgHandler(ONCEHELLO_REQ, OnceHello)
}
