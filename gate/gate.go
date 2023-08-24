package main

import (
	"flag"
	"github.com/invokerw/demo/g"
	"github.com/invokerw/demo/gate/internal/config"
	"github.com/invokerw/demo/pkg/glog"
	"github.com/invokerw/demo/pkg/gnats"
	"github.com/nats-io/nats.go"
	"github.com/zeromicro/go-zero/core/conf"
	"os"
	"os/signal"
	"syscall"
)

var configFile = flag.String("f", "etc/gate.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	logger := &glog.LoggerImpl{}
	agent, err := gnats.NewGNats(logger, c.NatsURL)
	if err != nil {
		logger.Error(err)
		return
	}
	defer agent.Close()
	err = agent.Subscribe(g.NATS_GATE_SUBJECT, func(msg *nats.Msg) {
		logger.Info(g.NATS_GATE_SUBJECT, "Subscribe receive:", string(msg.Data))
	})
	if err != nil {
		logger.Error(err)
		return
	}

	// Wait for signals
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-chSignal:
		logger.Infof("signal received: %v", s)
		_ = logger.SyncLogger()
		signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	}
	logger.Info("Server end")
}
