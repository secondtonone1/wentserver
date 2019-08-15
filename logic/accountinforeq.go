package logic

import (
	"fmt"
	"wentserver/config"
	"wentserver/netmodel"
	wentproto "wentserver/proto"
	"wentserver/protocol"

	"github.com/gogo/protobuf/proto"
)

func RegAccountInfoReq() {
	var AccountInfoReq netmodel.CallBackFunc = func(se interface{}, param interface{}) error {
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
		inforeq := &wentproto.CSAccountInfo{}

		err := proto.Unmarshal(msgpacket.Body.Data, inforeq)
		if err != nil {
			return config.ErrProtobuffUnMarshal
		}

		actinfo, err := GetAccountManagerIns().GetAccount(inforeq.Accountname)
		if err != nil {
			actinforsp := new(protocol.MsgPacket)
			actinforsp.Head.Id = ACCOUNTINFO_RSP
			inforsp := &wentproto.SCAccountInfo{
				Errid: ERR_ACTNOTEXIST,
			}

			rspdata, err := proto.Marshal(inforsp)
			if err != nil {
				return config.ErrProtobuffMarshal
			}

			actinforsp.Head.Len = uint16(len(rspdata))
			actinforsp.Body.Data = rspdata
			err = session.AsyncSend(actinforsp)
			if err != nil {
				fmt.Println("Handle Msg HelloworldReq failed")
				return config.ErrHelloWorldReqFailed
			}

			return nil
		}

		actinfopl, ok := actinfo.(*wentproto.AccountInfo)
		if !ok {
			return config.ErrTypeAssertain
		}
		actinforsp := new(protocol.MsgPacket)
		actinforsp.Head.Id = ACCOUNTINFO_RSP

		inforsp := &wentproto.SCAccountInfo{
			Accountinfo: actinfopl,
		}
		rspdata, err := proto.Marshal(inforsp)
		if err != nil {
			return config.ErrProtobuffMarshal
		}

		actinforsp.Head.Len = uint16(len(rspdata))
		actinforsp.Body.Data = rspdata
		err = session.AsyncSend(actinforsp)
		if err != nil {
			fmt.Println("Handle Msg HelloworldReq failed")
			return config.ErrHelloWorldReqFailed
		}
		return nil
	}

	netmodel.GetMsgHandlerIns().RegMsgHandler(ACCOUNTINFO_REQ, AccountInfoReq)
}
