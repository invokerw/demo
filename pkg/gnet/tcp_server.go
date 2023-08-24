package gnet

import (
	"fmt"
	"github.com/invokerw/demo/misc/gmath"
	"github.com/invokerw/demo/pkg/glog"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type TcpServerOption func(s *TcpServer)

type TcpServer struct {
	listener   net.Listener
	addr       string
	connIdStub atomic.Uint64
	connOpts   ConnOpts

	connMap map[uint64]*Conn

	wg sync.WaitGroup
	sync.RWMutex
	glog.Logger
}

func NewTcpServer(addr string, logger glog.Logger, opts ConnOpts, options ...TcpServerOption) (*TcpServer, error) {
	if opts.Callback == nil {
		return nil, fmt.Errorf("opts callback is nil")
	}
	s := &TcpServer{
		addr:     addr,
		connOpts: opts,
		connMap:  make(map[uint64]*Conn),
		Logger:   logger,
	}

	for _, option := range options {
		if option != nil {
			option(s)
		}
	}
	return s, nil
}

func (this *TcpServer) ListenAndServe() error {
	l, err := net.Listen("tcp", this.addr)
	if err != nil {
		return err
	}
	this.Lock()
	if this.listener != nil {
		_ = l.Close()
		this.Unlock()
		return fmt.Errorf("tcp server already started")
	}
	this.listener = l
	this.Unlock()

	defer func() {
		this.Lock()
		defer this.Unlock()
		if this.listener == nil {
			return
		}
		_ = this.listener.Close()
		this.listener = nil
	}()
	var retryDelay time.Duration
	for {
		rawConn, err := this.listener.Accept()
		if err != nil {
			if ne, ok := err.(interface {
				Temporary() bool
			}); ok && ne.Temporary() {
				retryDelay = gmath.Min(retryDelay+time.Millisecond*10, time.Second)
				<-time.After(retryDelay)
				continue
			} else {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return nil
				} else {
					return err
				}
			}
		}

		retryDelay = 0
		connID := this.connIdStub.Add(1)
		this.wg.Add(1)
		go func() {
			this.newConn(connID, rawConn)
		}()
	}
}

func (this *TcpServer) Shutdown() error {
	this.Lock()
	defer this.Unlock()
	if this.listener == nil {
		return nil
	}

	_ = this.listener.Close()
	this.listener = nil

	for _, conn := range this.connMap {
		_ = conn.Close()
	}
	this.connMap = nil
	return nil
}

func (this *TcpServer) removeConn(conn *Conn) {
	this.Lock()
	defer this.Unlock()
	delete(this.connMap, conn.ID())
}

func (this *TcpServer) newConn(connID uint64, rawConn net.Conn) {
	defer func() {
		this.wg.Done()
	}()
	conn := newConn(connID, rawConn, this.Logger, this.connOpts)
	conn.Serve()
}
