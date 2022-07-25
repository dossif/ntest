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
	Number  int
	Status  bool
	Time    time.Duration
	Message string
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
	var num int
	for {
		select {
		case <-p.Ctx.Done():
			fmt.Println(fmt.Sprintf("pinger: %v", p.Ctx.Err()))
			close(journal)
			return
		default:
			var ping Ping
			tt, err := p.Api.PingAttempts(destIp, time.Second*5, 1)
			if err != nil {
				ping.Number = num
				ping.Status = false
				ping.Time = time.Second * 0
				ping.Message = err.Error()
			} else {
				ping.Number = num
				ping.Status = true
				ping.Time = tt
				ping.Message = "ok"
			}
			select {
			case journal <- ping:
			}
			num = num + 1
			time.Sleep(time.Second * 1)
		}
	}
}
