package pinger

import (
	"context"
	"fmt"
	goping "github.com/digineo/go-ping"
	"log"
	"net"
	"sync"
	"time"
)

// github.com/digineo/go-ping

type Ping struct {
	Status  bool
	Time    time.Duration
	Message string
}

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

func (p *Pinger) Ping(ctx context.Context, wg *sync.WaitGroup, dest string, journal chan Ping) {
	defer wg.Done()
	destIp, _ := net.ResolveIPAddr("ip4", dest)
	for {
		select {
		case <-ctx.Done():
			fmt.Println(fmt.Sprintf("pinger: %v", ctx.Err()))
			return
		default:
			var ping Ping
			tt, err := p.Api.PingAttempts(destIp, time.Second*5, 1)
			if err != nil {
				ping.Status = false
				ping.Time = time.Second * 0
				ping.Message = err.Error()
			} else {
				ping.Status = true
				ping.Time = tt
				ping.Message = "ok"
			}
		}
	}
}
