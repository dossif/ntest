// Package signal derives a cancellable context.Context that is cancelled
// when the process receives SIGINT or SIGTERM, giving the running test a
// chance to shut down cleanly instead of being killed outright.
package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// ContextWithSignal returns a child of ctx that gets cancelled on the first
// SIGINT or SIGTERM.
func ContextWithSignal(ctx context.Context) context.Context {
	newCtx, cancel := context.WithCancel(ctx)
	// Buffered with capacity 1: signal.Notify never blocks on send, so an
	// unbuffered channel could silently drop a signal delivered before this
	// goroutine reaches the receive below (flagged by `go vet`).
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()
	return newCtx
}
