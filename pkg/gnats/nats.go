package gnats

import (
	"github.com/invokerw/demo/pkg/glog"
	"github.com/nats-io/nats.go"
)

type GNats struct {
	glog.Logger
	conn *nats.Conn
}

func NewGNats(logger glog.Logger, url string) (*GNats, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &GNats{Logger: logger, conn: conn}, nil
}

func (g *GNats) Publish(subject string, data []byte) error {
	return g.conn.Publish(subject, data)
}

func (g *GNats) Subscribe(subject string, handler func(msg *nats.Msg)) error {
	_, err := g.conn.Subscribe(subject, handler)
	return err
}

func (g *GNats) Close() {
	g.conn.Close()
}
