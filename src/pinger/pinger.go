package pinger

import (
	goping "github.com/digineo/go-ping"
	"log"
	"net"
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

func (p *Pinger) Ping(dest string, journal chan Ping) error {
	destIp, _ := net.ResolveIPAddr("ip4", dest)
	for true {
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

		journal <- ping
		time.Sleep(time.Second * 1)
	}
	return nil
}
