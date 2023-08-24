package gnet

import (
	"github.com/invokerw/demo/pkg/glog"
	"net"
	"time"
)

func NewTcpClient(addr string, logger glog.Logger, opts ConnOpts, dialTimeout time.Duration) (*Conn, error) {
	rawConn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return nil, err
	}
	conn := newConn(0, rawConn, logger, opts)
	return conn, nil
}
