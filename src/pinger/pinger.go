package pinger

import (
	goping "github.com/digineo/go-ping"
	"log"
	"net"
	"time"
)

// github.com/digineo/go-ping

type Pinger struct {
	Api *goping.Pinger
	Log *log.Logger
}

func NewPinger(bind net.IP, logger *log.Logger) (Pinger, error) {
	pinger, err := goping.New(bind.String(), "")
	if err != nil {
		log.Fatalf("failed to create new pinger: %v", err)
	}
	return Pinger{
		Api: pinger,
		Log: logger,
	}, nil
}

func (p *Pinger) Ping(dest string, journal chan string) error {
	destIp, _ := net.ResolveIPAddr("ip4", dest)
	for true {
		tt, err := p.Api.PingAttempts(destIp, time.Second*5, 1)
		if err != nil {
			log.Fatalf("filed to ping %v: %v", dest, err)
		}
		journal <- tt.String()
		time.Sleep(time.Second * 1)
	}
	return nil
}
