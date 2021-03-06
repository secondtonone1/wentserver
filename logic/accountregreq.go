package logic

import (
	"fmt"
	"wentserver/config"
	"wentserver/netmodel"
	wentproto "wentserver/proto"
	"wentserver/protocol"

	"github.com/gogo/protobuf/proto"
)

func RegAccountRegReq() {
	var AccountRegReq netmodel.CallBackFunc = func(se interface{}, param interface{}) error {
		msgpacket, ok := param.(*protocol.MsgPacket)
		if !ok {
			return config.ErrTypeAssertain
		}

		session, ok := se.(*netmodel.Session)
		if !ok {
			return config.ErrTypeAssertain
		}

		fmt.Println("Server recieve from ", session.RawConn().RemoteAddr().String())
		fmt.Println("Server Recv MsgID is ", msgpacket.Head.Id)

		inforeq := &wentproto.CSAccountReg{}

		err := proto.Unmarshal(msgpacket.Body.Data, inforeq)
		if err != nil {
			return config.ErrProtobuffUnMarshal
		}

		_, geter := GetAccountManagerIns().GetAccount(inforeq.Accountname)
		if geter == nil {

			regrsp := new(protocol.MsgPacket)
			regrsp.Head.Id = ACCOUNTREG_RSP

			inforsp := &wentproto.SCAccountReg{
				Errid: ERR_ACCOUNTHASEXIST,
			}
			rspdata, err := proto.Marshal(inforsp)
			if err != nil {
				return config.ErrProtobuffMarshal
			}

			regrsp.Head.Len = uint16(len(rspdata))
			regrsp.Body.Data = rspdata
			err = session.AsyncSend(regrsp)
			return nil
		}

		regrsp := new(protocol.MsgPacket)
		regrsp.Head.Id = ACCOUNTREG_RSP

		uid, uiderr := GetGenuidIns().generateuid()
		if uiderr != nil {
			return config.ErrAccountRegFailed
		}

		actinfo := wentproto.AccountInfo{
			Accountid:   uid,
			Accountname: inforeq.Accountname,
		}

		GetAccountManagerIns().RegAccount(inforeq.Accountname, &actinfo)
		inforsp := &wentproto.SCAccountReg{
			Accountinfo: &actinfo,
		}
		rspdata, err := proto.Marshal(inforsp)
		if err != nil {
			return config.ErrProtobuffMarshal
		}

		regrsp.Head.Len = uint16(len(rspdata))
		regrsp.Body.Data = rspdata
		err = session.AsyncSend(regrsp)
		if err != nil {
			fmt.Println("Handle Msg HelloworldReq failed")
			return config.ErrHelloWorldReqFailed
		}
		return nil
	}

	netmodel.GetMsgHandlerIns().RegMsgHandler(ACCOUNTREG_REQ, AccountRegReq)
}
