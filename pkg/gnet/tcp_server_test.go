package gnet

import (
	"github.com/invokerw/demo/pkg/glog"
	"testing"
	"time"
)

var logger = &glog.LoggerImpl{}

type cb struct {
	ret bool
}

func (c2 cb) OnConnect(c *Conn) error {
	logger.Info("OnConnect", c.ID())
	return nil
}

func (c2 cb) OnMessage(c *Conn, data *Msg) error {
	logger.Info("OnMessage", c.ID(), data.Header().DebugString(), string(data.Payload()))
	if c2.ret {
		msg := &Msg{}
		msg.SetPayload([]byte("world"))
		_ = c.Send(msg)
	}
	return nil
}

func (c2 cb) OnClose(c *Conn) {
	logger.Info("OnClose", c.ID())
}

var _ NetCallBack = &cb{}

func TestTcpServer(t *testing.T) {
	go func() {
		time.Sleep(time.Second * 2)
		opts := ConnOpts{
			Callback: &cb{ret: false},
		}
		client, err := NewTcpClient("localhost:9999", logger, opts, 0)
		if err != nil {
			logger.Error(err)
			return
		}
		go client.Serve()
		for {
			time.Sleep(time.Second)
			msg := &Msg{}
			msg.SetPayload([]byte("hello"))
			err = client.Send(msg)
			if err != nil {
				logger.Error(err)
				return
			}
		}

	}()
	opts := ConnOpts{
		Callback: &cb{ret: true},
	}
	server, err := NewTcpServer("localhost:9999", logger, opts, nil)
	err = server.ListenAndServe()
	if err != nil {
		logger.Error(err)
	}
}
