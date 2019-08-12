package logic

import (
	"fmt"
	"wentserver/config"
	"wentserver/netmodel"
	wentproto "wentserver/proto"
	"wentserver/protocol"

	"github.com/gogo/protobuf/proto"
)

func RegPlayerInfoReq() {
	var PlayerInfoReq netmodel.CallBackFunc = func(se interface{}, param interface{}) error {
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
		inforeq := &wentproto.CSPlayerInfo{}

		err := proto.Unmarshal(msgpacket.Body.Data, inforeq)
		if err != nil {
			return config.ErrProtobuffUnMarshal
		}

		playerinforsp := new(protocol.MsgPacket)
		playerinforsp.Head.Id = PLAYERINFO_RSP

		playerinfos := wentproto.PlayerInfo{
			Accountid:   1,
			Accountname: inforeq.Accountname,
		}

		inforsp := &wentproto.SCPlayerInfo{
			Playerinfo: &playerinfos,
		}
		rspdata, err := proto.Marshal(inforsp)
		if err != nil {
			return config.ErrProtobuffMarshal
		}

		playerinforsp.Head.Len = uint16(len(rspdata))
		playerinforsp.Body.Data = rspdata
		err = session.AsyncSend(playerinforsp)
		if err != nil {
			fmt.Println("Handle Msg HelloworldReq failed")
			return config.ErrHelloWorldReqFailed
		}
		return nil
	}

	netmodel.MsgHandler.RegMsgHandler(PLAYERINFO_REQ, PlayerInfoReq)
}
