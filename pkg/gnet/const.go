package gnet

import "time"

const (
	DefaultReceiveBufSize      = 1024 * 8
	BigRecvBufMaxStarvedCycles = 100
)

type NetCallBack interface {
	OnConnect(c *Conn) error
	OnMessage(c *Conn, data *Msg) error
	OnClose(c *Conn)
}

type ConnOpts struct {
	Callback    NetCallBack
	ReadTimeout time.Duration
}
