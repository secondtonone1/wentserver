package netmodel

import (
	"sync"
	"wentmin/common"
	"wentmin/protocol"
)

type MsgHandlerInter interface {
	HandleMsgPacket(param interface{}) error
	RegMsgHandler(param interface{}) error
}

type CallBackFunc func(session interface{}, param interface{}) error
type MsgHandlerImpl struct {
	cbfuncs map[uint16]CallBackFunc
	rwlock  sync.RWMutex
}

func (mh *MsgHandlerImpl) HandleMsgPacket(param interface{}, se interface{}) error {

	var (
		msgpacket *protocol.MsgPacket
		callback  CallBackFunc
		ok        bool
		session   *Session
	)
	if msgpacket, ok = param.(*protocol.MsgPacket); !ok {
		return common.ErrTypeAssertain
	}

	if session, ok = se.(*Session); !ok {
		return common.ErrTypeAssertain
	}

	if callback, ok = mh.cbfuncs[msgpacket.Head.Id]; !ok {
		//不存在
		return common.ErrMsgHandlerReg
	}

	return callback(session, param)
}

func (mh *MsgHandlerImpl) RegMsgHandler(cbid uint16, param interface{}) error {
	var (
		callback CallBackFunc
		ok       bool
	)

	if callback, ok = param.(CallBackFunc); !ok {
		return common.ErrParamCallBack
	}

	mh.cbfuncs[cbid] = callback
	return nil
}

//goroutine safe
func (mh *MsgHandlerImpl) SafeHandleMsgPacket(param interface{}, se interface{}) error {
	mh.rwlock.RLock()
	defer mh.rwlock.RUnlock()

	var (
		msgpacket *protocol.MsgPacket
		callback  CallBackFunc
		ok        bool
		session   *Session
	)
	if msgpacket, ok = param.(*protocol.MsgPacket); !ok {
		return common.ErrTypeAssertain
	}

	if session, ok = se.(*Session); !ok {
		return common.ErrTypeAssertain
	}

	if callback, ok = mh.cbfuncs[msgpacket.Head.Id]; !ok {
		//不存在
		return common.ErrMsgHandlerReg
	}

	return callback(session, param)
}

//goroutine safe
func (mh *MsgHandlerImpl) SafeRegMsgHandler(cbid uint16, param interface{}) error {
	mh.rwlock.Lock()
	defer mh.rwlock.Unlock()
	var (
		callback CallBackFunc
		ok       bool
	)

	if callback, ok = param.(CallBackFunc); !ok {
		return common.ErrParamCallBack
	}

	mh.cbfuncs[cbid] = callback
	return nil
}

var ins *MsgHandlerImpl
var once sync.Once

func GetMsgHandlerIns() *MsgHandlerImpl {
	once.Do(func() {
		ins = &MsgHandlerImpl{cbfuncs: make(map[uint16]CallBackFunc)}
	})
	return ins
}
