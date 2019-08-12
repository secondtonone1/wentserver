package main

import (
	"fmt"
	"wentserver/config"
	"wentserver/logic"
	"wentserver/netmodel"
	wentproto "wentserver/proto"
	"wentserver/protocol"

	"github.com/gogo/protobuf/proto"
)

func main() {

	cs, err := netmodel.Dial("tcp4", "127.0.0.1:10006")
	if err != nil {
		return
	}
	packet := new(protocol.MsgPacket)
	packet.Head.Id = logic.PLAYERINFO_REQ
	csplayerinfo := &wentproto.CSPlayerInfo{
		Accountname: "Zack",
	}

	//protobuf编码
	pData, err := proto.Marshal(csplayerinfo)
	if err != nil {
		fmt.Println(config.ErrProtobuffMarshal.Error())
		return
	}
	packet.Head.Len = (uint16)(len(pData))
	packet.Body.Data = pData
	cs.Send(packet)
	packetrsp, err := cs.Recv()
	if err != nil {
		fmt.Println("receive error")
		return
	}

	datarsp := packetrsp.(*protocol.MsgPacket)
	fmt.Println("packet id is", datarsp.Head.Id)
	fmt.Println("packet len is", datarsp.Head.Len)
	scplayerinfo := &wentproto.SCPlayerInfo{}
	error2 := proto.Unmarshal(datarsp.Body.Data, scplayerinfo)
	if error2 != nil {
		fmt.Println(config.ErrProtobuffUnMarshal.Error())
		return
	}
	fmt.Println("scplayerinfo.Playerinfo.Accountid is ", scplayerinfo.Playerinfo.Accountid)
	fmt.Println("scplayerinfo.Playerinfo.Accountname is ", scplayerinfo.Playerinfo.Accountname)

}
