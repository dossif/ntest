package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"pping/src/pinger"
	"pping/src/ui"
	"sync"
	"syscall"
	"time"
)

// handle signals and return context
func contextWithSignal(ctx context.Context) context.Context {
	newCtx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-signals:
			cancel()
		}
	}()
	return newCtx
}

func main() {
	// create waitgroup
	wg := sync.WaitGroup{}
	// creat main context
	mainCtx := context.Background()
	mainCtx = contextWithSignal(mainCtx)
	mainCtx, mainCancel := context.WithTimeout(mainCtx, time.Second*999)
	defer mainCancel()
	// create child context
	childCtx := context.WithValue(mainCtx, "name", "func1")
	childCtx, childCancel := context.WithTimeout(childCtx, time.Second*999)
	defer childCancel()
	// exec long function with context and waitgroup
	journal := make(chan pinger.Ping)
	ping, err := pinger.NewPinger(net.ParseIP("0.0.0.0"), &log.Logger{})
	if err != nil {
		log.Fatalf("failed to create pinger: %v", err)
	}
	if err != nil {
		log.Fatalf("failed to create tui: %v", err)
	}
	wg.Add(1)
	go func() { ping.Ping(childCtx, &wg, "1.1.1.1", journal) }()
	wg.Add(1)
	nui, err := ui.NewUi(childCtx, &wg, "ui_cli")
	err = nui.RenderUi()
	// wait until cancel main context
	<-mainCtx.Done()
	fmt.Println("wait all wg done")
	wg.Wait()
	fmt.Println(fmt.Sprintf("main: %v", mainCtx.Err()))
}
