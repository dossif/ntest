package main

import (
	"context"
	"fmt"
	args2 "pping/src/args"
	"pping/src/signal"
)

const (
	appName = "ntest"
)

var (
	appVersion = "0.0.0"
)

func main() {
	args := args2.NewArgs(appName, appVersion)
	_ = signal.ContextWithSignal(context.Background())
	switch true {
	case args.IcmpPing.Command.Happened():
		fmt.Println("ICMP PING")
	case args.TcpTest.Command.Happened():
		fmt.Println("TCP PING")
	default:
		panic("unknown command")
	}

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
