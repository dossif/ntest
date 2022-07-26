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
	Host     string
	Ip       string
	Sequence int
	Status   bool
	Rtt      time.Duration
	Error    error
}

type Pinger struct {
	Api *goping.Pinger
	Log *log.Logger
	Ctx context.Context
	Wg  *sync.WaitGroup
}

func NewPinger(ctx context.Context, wg *sync.WaitGroup, bind net.IP, logger *log.Logger) (Pinger, error) {
	pinger, err := goping.New(bind.String(), "")
	if err != nil {
		log.Fatalf("failed to create new pinger: %v", err)
	}
	return Pinger{
		Api: pinger,
		Log: logger,
		Ctx: ctx,
		Wg:  wg,
	}, nil
}

func (p *Pinger) Ping(dest string, journal chan Ping) {
	defer p.Wg.Done()
	defer fmt.Println("exit from pinger")
	destIp, _ := net.ResolveIPAddr("ip4", dest)
	var seq int
	for {
		select {
		case <-p.Ctx.Done():
			fmt.Println(fmt.Sprintf("pinger: %v", p.Ctx.Err()))
			close(journal)
			return
		default:
			rtt, err := p.Api.PingAttempts(destIp, time.Second*5, 1)
			var ping Ping
			ping.Host = dest
			ping.Ip = destIp.String()
			ping.Sequence = seq
			ping.Rtt = rtt
			ping.Error = err
			if err != nil {
				ping.Status = false
			} else {
				ping.Status = true
			}
			select {
			case journal <- ping:
			}
			seq = seq + 1
			time.Sleep(time.Second * 1)
		}
	}
}
