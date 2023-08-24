package gnet

import (
	"errors"
)

type Msg struct {
	h       Header
	payload []byte
}

func (this *Msg) Size() uint32 {
	return this.h.Size() + this.h.PayloadSize
}

func (this *Msg) Header() *Header {
	return &this.h
}

func (this *Msg) Payload() []byte {
	return this.payload
}

func (this *Msg) SetPayload(payload []byte) {
	this.payload = payload
	this.h.PayloadSize = uint32(len(payload))
}

func (this *Msg) Marshal(buf []byte) (int, error) {
	if uint32(len(buf)) < this.Size() {
		return 0, errors.New("Msg:Marshal, the buffer is too small")
	}
	if int(this.h.PayloadSize) != len(this.payload) {
		return 0, errors.New("Msg:Marshal, h.PayloadSize is out of sync")
	}

	n, err := this.h.Marshal(buf)
	if err != nil {
		return n, err
	}
	if this.h.PayloadSize > 0 {
		n += copy(buf[n:], this.payload)
	}
	return n, nil
}

func (this *Msg) Unmarshal(buf []byte) (int, error) {
	n, err := this.h.Unmarshal(buf)
	if err != nil {
		return n, err
	}
	if uint32(len(buf)) < this.Size() {
		return 0, ErrAgain
	}
	if this.h.PayloadSize > 0 {
		s := int(this.h.PayloadSize)
		// TODO WEIFEI use a pool
		this.payload = make([]byte, s, s)
		n += copy(this.payload, buf[n:])
	}
	return n, nil
}
