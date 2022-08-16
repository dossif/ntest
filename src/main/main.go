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
	Execute(ctx context.Context, arguments args.Arguments) error
}

func NewTest(arg args.Arguments) Test {
	switch true {
	case arg.IcmpPing.Command.Happened():
		return icmpping.NewTest()
	case arg.TcpTest.Command.Happened():
	}
	return nil
}

func main() {
	arg := args.NewArgs(appName, appVersion)
	ctx := signal.ContextWithSignal(context.Background())
	t := NewTest(arg)
	err := t.Execute(ctx, arg)
	if err != nil {
		panic(fmt.Sprintf("failed to init test: %v", err))
	}
	defer fmt.Println("exit main")

	//dest := os.Args[1]
	//// create waitgroup
	//wg := sync.WaitGroup{}
	//// creat main context
	//mainCtx := context.Background()
	//mainCtx = contextWithSignal(mainCtx)
	//mainCtx, mainCancel := context.WithTimeout(mainCtx, time.Second*999)
	//defer mainCancel()
	//// pinger
	//pngCtx := context.WithValue(mainCtx, "name", "func1")
	//pngCtx, pngCancel := context.WithTimeout(mainCtx, time.Second*999)
	//defer pngCancel()
	//journal := make(chan pinger.Ping)
	//ping, err := pinger.NewPinger(pngCtx, &wg, net.ParseIP("0.0.0.0"), &log.Logger{})
	//if err != nil {
	//	log.Fatalf("failed to create pinger: %v", err)
	//}
	//go func() { ping.Ping(dest, journal) }()
	//wg.Add(1)
	//// ui
	//uiCtx := context.WithValue(mainCtx, "name", "func1")
	//uiCtx, uiCancel := context.WithTimeout(mainCtx, time.Second*999)
	//defer uiCancel()
	//nui, err := ui.NewUi(uiCtx, &wg, "ui_cli")
	//if err != nil {
	//	log.Fatalf("failed to create ui: %v", err)
	//}
	//go func() { err = nui.RenderUi(journal) }()
	//wg.Add(1)
	//// wait until cancel main context
	//<-mainCtx.Done()
	//fmt.Println("wait all wg done")
	//wg.Wait()
	//fmt.Println(fmt.Sprintf("main: %v", mainCtx.Err()))
}
