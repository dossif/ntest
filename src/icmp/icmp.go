package icmp

import (
	"context"
	"fmt"
	"github.com/digineo/go-ping"
	log "github.com/sirupsen/logrus"
	"net"
	"ntest/src/dns"
	"time"
)

type Test struct {
	Api      ping.Pinger
	Host     string
	Ip       net.IPAddr
	Timeout  time.Duration
	Interval time.Duration
	Warn     time.Duration
}

func NewTest(bind string, host string, timeout int, interval int, warn int) *Test {
	bindIp, err := dns.ResolveAddr(bind)
	if err != nil {
		log.Fatalf("failed to resolve bind %v", err)
	}
	hostIp, err := dns.ResolveAddr(host)
	if err != nil {
		log.Fatalf("failed to resolve host %v", err)
	}
	api, err := ping.New(bindIp.String(), "")
	if err != nil {
		log.Fatalf("failed to create new pinger: %v", err)
	}
	return &Test{
		Api:      *api,
		Host:     host,
		Ip:       *hostIp,
		Timeout:  time.Duration(timeout) * time.Millisecond,
		Interval: time.Duration(interval) * time.Millisecond,
		Warn:     time.Duration(warn) * time.Millisecond,
	}
}

func (t *Test) Execute(ctx context.Context) error {
	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()
	var seq int
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			rtt, err := t.Api.Ping(&t.Ip, t.Timeout)
			seq = seq + 1
			cf := new(log.TextFormatter)
			cf.FullTimestamp = true
			log.SetFormatter(cf)
			lg := log.Fields{
				"seq":  seq,
				"dest": fmt.Sprintf("%v (%v)", t.Host, t.Ip.IP),
				"rtt":  rtt.Round(time.Millisecond),
			}
			if err != nil {
				log.WithFields(lg).Errorf("ping error: %v", err)
			} else if rtt > t.Warn {
				log.WithFields(lg).Warnf("warn threshold %v exceed", t.Warn)
			} else {
				log.WithFields(lg).Infof("ping ok")
			}
		}
	}
}
