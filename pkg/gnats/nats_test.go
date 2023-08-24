package gnats

import (
	"github.com/invokerw/demo/pkg/glog"
	"github.com/nats-io/nats.go"
	"testing"
	"time"
)

func TestGNats(t *testing.T) {
	agent, err := NewGNats(&glog.LoggerImpl{}, "nats://localhost:4222")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer agent.Close()
	err = agent.Subscribe("test", func(msg *nats.Msg) {
		t.Log("Subscribe receive:", string(msg.Data))
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	time.Sleep(time.Second)
	err = agent.Publish("test", []byte("test"))
	if err != nil {
		t.Fatal(err)
		return
	}
	time.Sleep(time.Second)
}
