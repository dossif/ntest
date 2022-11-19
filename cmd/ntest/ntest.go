package main

import (
	"context"
	"github.com/dossif/ntest/internal/args"
	"github.com/dossif/ntest/internal/http"
	"github.com/dossif/ntest/internal/icmp"
	"github.com/dossif/ntest/internal/signal"
	"github.com/dossif/ntest/internal/tcp"
	log "github.com/sirupsen/logrus"
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
		log.Fatalf("failed to init test: %v", err)
	}
	//defer fmt.Println("exit main")
}
