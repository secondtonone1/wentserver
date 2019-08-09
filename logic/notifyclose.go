package logic

import (
	"fmt"
	"time"
	"wentby/config"
	"wentby/netmodel"
	"wentby/protocol"
)

func NotifyClientClose(se interface{}, param interface{}) error {

	notifyclose := new(protocol.MsgPacket)
	notifyclose.Head.Id = HELLOWORLD_RSP
	notifyclose.Head.Len = uint16(len("Server close your connection"))
	notifyclose.Body.Data = []byte("Server close your connection")
	session, ok := se.(*netmodel.Session)
	if !ok {
		return config.ErrTypeAssertain
	}
	err := session.AsyncSend(notifyclose)
	if err != nil {
		fmt.Println(config.ErrAsyncSendFailed.Error())
		return config.ErrAsyncSendFailed
	}

	time.Sleep(time.Second * time.Duration(20))
	err = session.AsyncSend(nil)
	if err != nil {
		fmt.Println(config.ErrAsyncSendFailed.Error())
		return config.ErrAsyncSendFailed
	}

	return nil

}
