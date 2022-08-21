package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"ntest/src/args"
	"ntest/src/http"
	"ntest/src/icmp"
	"ntest/src/signal"
	"ntest/src/tcp"
)

const (
	appName = "ntest"
)

var (
	appVersion = "0.0.0"
)

type Test interface {
	Execute(ctx context.Context) error
}

func NewTest(arg args.Arguments) Test {
	switch true {
	case arg.IcmpPing.Command.Happened():
		f := arg.IcmpPing.Flags
		return icmp.NewTest(*f.Bind, *f.Host, *f.Timeout, *f.Interval, *f.Warn, *f.Ns)
	case arg.TcpTest.Command.Happened():
		f := arg.TcpTest.Flags
		return tcp.NewTest(*f.Bind, *f.Host, *f.Port, *f.Timeout, *f.Interval, *f.Ns)
	case arg.HttpTest.Command.Happened():
		f := arg.HttpTest.Flags
		return http.NewTest(*f.Bind, *f.Host, *f.Timeout, *f.Interval, *f.Method, *f.Domain, *f.Body, *f.Ns)
	}
	return nil
}

func main() {
	arg := args.NewArgs(appName, appVersion)
	ctx := signal.ContextWithSignal(context.Background())
	t := NewTest(arg)
	err := t.Execute(ctx)
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to init test: %v", err))
	}
	//defer fmt.Println("exit main")
}
