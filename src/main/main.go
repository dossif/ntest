package main

import (
	"fmt"
	goping "github.com/digineo/go-ping"
	"log"
	"net"
	"time"
)

func main() {
	bind := net.ParseIP("0.0.0.0")
	dest, _ := net.ResolveIPAddr("ip4", "1.1.1.1")
	pinger, err := goping.New(bind.String(), "")
	if err != nil {
		log.Fatalf("failed to create new pinger: %v", err)
	}
	defer pinger.Close()
	tt, err := pinger.PingAttempts(dest, time.Second*5, 1)
	if err != nil {
		log.Fatalf("filed to ping %v: %v", dest, err)
	}
	fmt.Println(tt.String())
}
