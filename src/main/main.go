package main

import (
	"log"
	"net"
	"pping/src/pinger"
	"pping/src/tui"
)

func main() {
	journal := make(chan pinger.Ping)
	ping, err := pinger.NewPinger(net.ParseIP("0.0.0.0"), &log.Logger{})
	if err != nil {
		log.Fatalf("failed to create pinger: %v", err)
	}
	go func() { _ = ping.Ping("1.1.1.1", journal) }()
	go func() { tui.StartTui(journal) }()

	for {
		select {
		case <-journal:
			//fmt.Println(msg)
		}
	}
}
