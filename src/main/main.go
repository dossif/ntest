package main

import (
	"context"
	"fmt"
	"pping/src/args"
	"pping/src/icmpping"
	"pping/src/signal"
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
		return icmpping.NewTest(*f.Bind, *f.Host, *f.Timeout)
	case arg.TcpTest.Command.Happened():
	}
	return nil
}

func main() {
	arg := args.NewArgs(appName, appVersion)
	ctx := signal.ContextWithSignal(context.Background())
	t := NewTest(arg)
	err := t.Execute(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to init test: %v", err))
	}
	//defer fmt.Println("exit main")
}
