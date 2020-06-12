package protocol

import (
	"fmt"
	"io"
	"net"
	"wentmin/common"
	"wentmin/components"
)

// PacketReader is used to unmarshal a complete packet from buff
type ReaderInter interface {
	// Read data from conn and build a complete packet.
	ReadPacket(conn net.Conn) (interface{}, error)
}

// PacketWriter is used to marshal packet into buff
type WriterInter interface {
	WritePacket(net.Conn, interface{}) error
}

// PacketProtocol just a composite interface
type ProtocolInter interface {
	ReaderInter
	WriterInter
}

type ProtocolImpl struct {
}

func (pi *ProtocolImpl) ParaseHead(packet interface{}, buff []byte) (interface{}, error) {
	if len(buff) < 4 {
		return nil, common.ErrBuffLenLess
	}
	msgpacket, ok := packet.(*MsgPacket)
	if !ok {
		fmt.Println("it's not msgpacket type")
		return nil, common.ErrTypeAssertain
	}
	stream := NewBigEndianStream(buff)
	var err error
	if msgpacket.Head.Id, err = stream.ReadUint16(); err != nil {
		return nil, common.ErrParaseMsgHead
	}

	if msgpacket.Head.Len, err = stream.ReadUint16(); err != nil {
		return nil, common.ErrParaseMsgHead
	}

	return msgpacket, nil
}

func (pi *ProtocolImpl) ReadPacket(conn net.Conn) (interface{}, error) {
	buff := make([]byte, 4)
	_, err := io.ReadAtLeast(conn, buff[:4], 4)
	if err != nil {
		fmt.Println(err.Error())
		return nil, common.ErrReadAtLeast
	}

	var msgpacket *MsgPacket = new(MsgPacket)
	value, err := pi.ParaseHead(msgpacket, buff[:4])

	msgpacket, ok := value.(*MsgPacket)
	if !ok {
		fmt.Println("it's not msgpacket type")
		return nil, common.ErrTypeAssertain
	}

	if components.MaxMsgLen < msgpacket.Head.Len {
		return nil, common.ErrMsgLenLarge
	}

	if uint16(len(msgpacket.Body.Data)) < msgpacket.Head.Len {
		msgpacket.Body.Data = make([]byte, msgpacket.Head.Len)
	}

	if _, err = io.ReadFull(conn, msgpacket.Body.Data[:msgpacket.Head.Len]); err != nil {
		fmt.Println("err is ", err.Error())
		return nil, common.ErrReadAtLeast
	}

	return msgpacket, nil
}

func (pi *ProtocolImpl) WritePacket(conn net.Conn, packet interface{}) error {
	var msgpacket *MsgPacket = packet.(*MsgPacket)
	if msgpacket == nil {
		return common.ErrPacketEmpty
	}
	msglen := 4 + msgpacket.Head.Len
	buff := make([]byte, msglen)
	stream := NewBigEndianStream(buff[:])
	if err := stream.WriteUint16(msgpacket.Head.Id); err != nil {
		return common.ErrWritePacketFailed
	}

	if err := stream.WriteUint16(msgpacket.Head.Len); err != nil {
		return common.ErrWritePacketFailed
	}

	if err := stream.WriteBuff(msgpacket.Body.Data); err != nil {
		return common.ErrWritePacketFailed
	}
	wn, err := conn.Write(buff)
	if err != nil {
		return common.ErrConnWriteFailed
	}
	fmt.Println("write bytes ", wn)
	return nil
}
