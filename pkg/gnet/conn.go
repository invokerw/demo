package gnet

import (
	"fmt"
	"github.com/invokerw/demo/pkg/glog"
	"net"
	"time"
)

type Conn struct {
	id      uint64
	rawConn net.Conn

	receiveBuff             []byte
	receiveBufNumBytes      int
	receiveBufStarvedCycles int

	opts ConnOpts
	glog.Logger
}

func newConn(id uint64, rawConn net.Conn, logger glog.Logger, opts ConnOpts) *Conn {
	c := &Conn{
		id:          id,
		rawConn:     rawConn,
		receiveBuff: make([]byte, DefaultReceiveBufSize),
		opts:        opts,
		Logger:      logger,
	}
	return c
}

func (this *Conn) ID() uint64 {
	return this.id
}

func (this *Conn) Close() error {
	return this.rawConn.Close()
}

func (this *Conn) Send(msg *Msg) error {
	var err error
	n := msg.Size()
	data := make([]byte, n, n)
	_, err = msg.Marshal(data)
	if err != nil {
		return err
	}
	_, err = this.rawConn.Write(data)
	return err
}

func (this *Conn) Serve() {
	var err error
	err = this.opts.Callback.OnConnect(this)
	if err != nil {
		this.Error("OnConnect error: %v", err)
		return
	}
	defer func() {
		if err != nil {
			this.Error("Serve error: %v", err)
		}
		this.opts.Callback.OnClose(this)
	}()
	readTimeout := this.opts.ReadTimeout
	if readTimeout <= 0 {
		readTimeout = time.Second * 10
	}
	var num, num2, off int
loop:
	for {
		if this.receiveBufNumBytes == len(this.receiveBuff) {
			err = fmt.Errorf("receiveBufNumBytes error: %d", this.receiveBufNumBytes)
		}
		if this.receiveBufStarvedCycles >= BigRecvBufMaxStarvedCycles && len(this.receiveBuff) > DefaultReceiveBufSize {
			newBuf := make([]byte, DefaultReceiveBufSize)
			copy(newBuf, this.receiveBuff[off:this.receiveBufNumBytes])
			this.receiveBuff = newBuf
			this.receiveBufStarvedCycles = 0
		}

		deadline := time.Now().Add(readTimeout)
		err = this.rawConn.SetReadDeadline(deadline)
		if err != nil {
			return
		}
		// TODO WEIFEI buff 优化
		num, err = this.rawConn.Read(this.receiveBuff[this.receiveBufNumBytes:])
		this.receiveBufNumBytes += num
		if err != nil {
			return
		}
		if this.receiveBufNumBytes <= DefaultReceiveBufSize {
			this.receiveBufStarvedCycles++
		} else {
			this.receiveBufStarvedCycles = 0
		}
		msg := &Msg{}
		for off < this.receiveBufNumBytes {
			num2, err = msg.Unmarshal(this.receiveBuff[off:this.receiveBufNumBytes])
			if err != nil {
				if err != ErrAgain {
					return
				}
				if off == 0 {
					if this.receiveBufNumBytes == len(this.receiveBuff) {
						newBuf := make([]byte, len(this.receiveBuff)*2)
						copy(newBuf, this.receiveBuff)
						this.receiveBuff = newBuf
						this.receiveBufStarvedCycles = 0
					}
				} else {
					if this.receiveBufNumBytes == len(this.receiveBuff) {
						copy(this.receiveBuff, this.receiveBuff[off:this.receiveBufNumBytes])
						this.receiveBufNumBytes -= off
						off = 0
					}
				}
				continue loop
			}
			err = this.opts.Callback.OnMessage(this, msg)
			if err != nil {
				return
			}
			off += num2
		}

		if off != this.receiveBufNumBytes {
			const format = "off should equal to receiveBufNumBytes here. off: %d, receiveBufNumBytes: %d"
			err = fmt.Errorf(format, off, this.receiveBufNumBytes)
			return
		}
		this.receiveBufNumBytes = 0
		off = 0
	}
}
