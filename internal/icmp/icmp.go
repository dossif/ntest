package icmp

import (
	"context"
	"fmt"
	"github.com/digineo/go-ping"
	"github.com/dossif/ntest/internal/dns"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type Test struct {
	Api      ping.Pinger
	Host     string
	Ip       net.IPAddr
	Timeout  time.Duration
	Interval time.Duration
	Warn     time.Duration
	Ns       string
}

func NewTest(bind string, host string, timeout int, interval int, warn int, ns string) *Test {
	bindIp, err := dns.ResolveAddr(bind, "")
	if err != nil {
		log.Fatalf("failed to resolve bind: %v", err)
	}
	hostIp, err := dns.ResolveAddr(host, ns)
	if err != nil {
		log.Fatalf("failed to resolve host: %v", err)
	}
	api, err := ping.New(bindIp.String(), "")
	if err != nil {
		log.Fatalf("failed to create new icmp pinger: %v", err)
	}
	return &Test{
		Api:      *api,
		Host:     host,
		Ip:       *hostIp,
		Timeout:  time.Duration(timeout) * time.Millisecond,
		Interval: time.Duration(interval) * time.Millisecond,
		Warn:     time.Duration(warn) * time.Millisecond,
		Ns:       ns,
	}
}

func (t *Test) Execute(ctx context.Context) error {
	ticker := time.NewTicker(t.Interval)
	cf := new(log.TextFormatter)
	cf.FullTimestamp = true
	cf.TimestampFormat = "15:04:05"
	log.SetFormatter(cf)
	defer ticker.Stop()
	var seq int
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			func() {
				seq = seq + 1
				var dest string
				if t.Host == t.Ip.IP.String() {
					dest = t.Ip.IP.String()
				} else {
					dest = fmt.Sprintf("%v[%v]", t.Host, t.Ip.IP.String())
				}
				lg := log.Fields{
					"seq":  seq,
					"dest": dest,
				}
				rtt, err := t.Api.Ping(&t.Ip, t.Timeout)
				if err != nil {
					log.WithFields(lg).Errorf("icmp error: %v", err)
				} else if rtt > t.Warn {
					log.WithFields(lg).Warnf("icmp rtt %v: warn threshold %v exceed", rtt.Round(time.Millisecond), t.Warn)
				} else {
					log.WithFields(lg).Infof("icmp rtt %v", rtt.Round(time.Millisecond))
				}
			}()
		}
	}
}
