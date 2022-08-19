package tcp

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"ntest/src/dns"
	"strconv"
	"time"
)

type Test struct {
	Api      *net.Dialer
	Host     string
	Ip       net.IPAddr
	Port     int
	Timeout  time.Duration
	Interval time.Duration
}

func NewTest(bind string, host string, port int, timeout int, interval int) *Test {
	bindIp, err := dns.ResolveAddr(bind)
	if err != nil {
		log.Fatalf("failed to resolve bind: %v", err)
	}
	hostIp, err := dns.ResolveAddr(host)
	if err != nil {
		log.Fatalf("failed to resolve host: %v", err)
	}
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: bindIp.IP, Port: 0, Zone: ""},
	}
	return &Test{
		Api:      dialer,
		Host:     host,
		Ip:       *hostIp,
		Port:     port,
		Timeout:  time.Duration(timeout) * time.Millisecond,
		Interval: time.Duration(interval) * time.Millisecond,
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
			seq = seq + 1
			addr := net.JoinHostPort(t.Ip.String(), strconv.Itoa(t.Port))
			conn, err := t.Api.DialContext(ctx, "tcp", addr)
			cf := new(log.TextFormatter)
			cf.FullTimestamp = true
			log.SetFormatter(cf)
			lg := log.Fields{
				"seq":  seq,
				"dest": fmt.Sprintf("%v (%v)", t.Host, t.Ip.IP),
				"port": t.Port,
			}
			if err != nil {
				log.WithFields(lg).Errorf("tcp error: %v", err)
			} else {
				log.WithFields(lg).Infof("tcp ok")
				defer func() { _ = conn.Close() }()
			}
		}
	}
}
