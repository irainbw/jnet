package base

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type Codec interface {
	Decode(Session Session) (IMessage, error)
	Encode(msg IMessage) ([]byte, error)
}

// package head : uint32 msgId + uint32  dataLen
// --------------
// | head | data |
// --------------

type PacketParser struct {
	PacketHeadLen int
	MaxPacketLen  int
	ByteOrder     binary.ByteOrder
}

func (p *PacketParser) Decode(Session Session) (IMessage, error) {
	//异常情况返回err
	CurBuf := Session.Read()
	if len(CurBuf) < p.PacketHeadLen {
		return nil, nil
	}
	msg := &Message{}
	dataBuff := bytes.NewReader(CurBuf[:p.PacketHeadLen])
	if err := binary.Read(dataBuff, p.ByteOrder, &msg.ID); err != nil {
		return nil, err
	}
	if err := binary.Read(dataBuff, p.ByteOrder, &msg.DataLen); err != nil {
		return nil, err
	}
	if int(msg.DataLen) > p.MaxPacketLen {
		return nil, errors.New("message too long")
	}
	fullPkgLen := p.PacketHeadLen + int(msg.DataLen)
	if len(CurBuf) < fullPkgLen {
		return nil, nil
	}
	msg.SetData(CurBuf[p.PacketHeadLen:fullPkgLen])
	Session.Next(fullPkgLen)
	return msg, nil
}

func (p *PacketParser) Encode(msg IMessage) ([]byte, error) {
	dataBuff := bytes.NewBuffer([]byte{})
	if err := binary.Write(dataBuff, p.ByteOrder, msg.GetMsgID()); err != nil {
		return nil, err
	}
	if err := binary.Write(dataBuff, p.ByteOrder, msg.GetDataLen()); err != nil {
		return nil, err
	}
	if err := binary.Write(dataBuff, p.ByteOrder, msg.GetData()); err != nil {
		return nil, err
	}
	return dataBuff.Bytes(), nil
}
