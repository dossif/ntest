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
	// pinger
	pngCtx := context.WithValue(mainCtx, "name", "func1")
	pngCtx, pngCancel := context.WithTimeout(mainCtx, time.Second*999)
	defer pngCancel()
	journal := make(chan pinger.Ping, 100)
	ping, err := pinger.NewPinger(pngCtx, &wg, net.ParseIP("0.0.0.0"), &log.Logger{})
	if err != nil {
		log.Fatalf("failed to create pinger: %v", err)
	}
	go func() { ping.Ping("1.1.1.1", journal) }()
	wg.Add(1)
	// ui
	uiCtx := context.WithValue(mainCtx, "name", "func1")
	uiCtx, uiCancel := context.WithTimeout(mainCtx, time.Second*999)
	defer uiCancel()
	nui, err := ui.NewUi(uiCtx, &wg, "ui_cli")
	if err != nil {
		log.Fatalf("failed to create ui: %v", err)
	}
	go func() { err = nui.RenderUi(journal) }()
	wg.Add(1)
	// wait until cancel main context
	<-mainCtx.Done()
	fmt.Println("wait all wg done")
	wg.Wait()
	fmt.Println(fmt.Sprintf("main: %v", mainCtx.Err()))
}
