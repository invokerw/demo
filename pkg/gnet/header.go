package gnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const HeaderSize = 12
const MagicID = 0x12345678

var ErrAgain = errors.New("again")

const (
	Offset_Magic       = 0
	Offset_MsgID       = 4
	Offset_PayloadSize = 8
)

type Header struct {
	MsgId       uint32
	PayloadSize uint32
}

func (this *Header) DebugString() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "MsgID: %v", this.MsgId)
	fmt.Fprintf(&buf, ", PayloadSize: %d", this.PayloadSize)
	return buf.String()
}

func (this *Header) Reset() {
	this.MsgId = 0
	this.PayloadSize = 0
}

func (this *Header) Marshal(buf []byte) (int, error) {
	if len(buf) < HeaderSize {
		return 0, errors.New("the buffer is too small")
	}
	binary.LittleEndian.PutUint32(buf[Offset_Magic:], MagicID)
	binary.LittleEndian.PutUint32(buf[Offset_MsgID:], this.MsgId)
	binary.LittleEndian.PutUint32(buf[Offset_PayloadSize:], this.PayloadSize)
	return HeaderSize, nil

}

func (this *Header) Unmarshal(buf []byte) (int, error) {
	n := len(buf)
	if n < HeaderSize {
		return 0, ErrAgain
	}
	magicID := binary.LittleEndian.Uint32(buf[Offset_Magic:])
	if magicID != MagicID {
		return 0, errors.New("magicID does not match")
	}
	this.MsgId = binary.LittleEndian.Uint32(buf[Offset_MsgID:])
	this.PayloadSize = binary.LittleEndian.Uint32(buf[Offset_PayloadSize:])
	return HeaderSize, nil
}

func (this *Header) Size() uint32 {
	return HeaderSize
}
