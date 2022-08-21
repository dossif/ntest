package http

import (
	"bytes"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
	"ntest/src/dns"
	"strings"
	"time"
)

type Test struct {
	Api      *http.Client
	Url      *url.URL
	Ip       *net.IPAddr
	Timeout  time.Duration
	Interval time.Duration
	Domain   string
	Method   string
	Body     string
}

func NewTest(bind string, host string, timeout int, interval int, method string, domain string, body string, ns string) *Test {
	bindIp, err := dns.ResolveAddr(bind, "")
	if err != nil {
		log.Fatalf("failed to resolve bind: %v", err)
	}
	hostUrl, err := url.Parse(host)
	if err != nil {
		log.Fatalf("failed to parse host: %v", err)
	}
	hostIp, err := dns.ResolveAddr(hostUrl.Host, ns)
	if err != nil {
		log.Fatalf("failed to resolve host: %v", err)
	}
	if hostUrl.Scheme == "" {
		hostUrl.Scheme = "http"
	}
	transport := http.Transport{
		Dial: (&net.Dialer{
			LocalAddr: &net.TCPAddr{IP: bindIp.IP, Port: 0, Zone: ""},
		}).Dial,
	}
	client := http.Client{
		Transport: &transport,
		Timeout:   4 * time.Second,
	}
	return &Test{
		Api:      &client,
		Url:      hostUrl,
		Ip:       hostIp,
		Timeout:  time.Duration(timeout) * time.Millisecond,
		Interval: time.Duration(interval) * time.Millisecond,
		Domain:   domain,
		Method:   method,
		Body:     body,
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
			cf := new(log.TextFormatter)
			cf.FullTimestamp = true
			log.SetFormatter(cf)
			
			lg := log.Fields{
				"seq":    seq,
				"dest":   fmt.Sprintf("%v (%v)", t.Url, t.Ip.IP),
				"method": t.Method,
			}
			seq = seq + 1
			
			req, err := http.NewRequest(t.Method, t.Url.String(), bytes.NewReader([]byte(t.Body)))
			if err != nil {
				log.Errorf("failed to create http request: %v", err)
			}
			rCtx, _ := context.WithTimeout(ctx, t.Timeout)
			req = req.WithContext(rCtx)
			resp, err := t.Api.Do(req)
			if err != nil {
				if !strings.Contains(err.Error(), "context canceled") {
					log.WithFields(lg).Errorf("http error: %v", err)
				}
			} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				log.WithFields(lg).Warnf("http %v error", resp.Status)
			} else if resp.StatusCode >= 500 && resp.StatusCode < 600 {
				log.WithFields(lg).Errorf("http %v error", resp.Status)
			} else {
				log.WithFields(lg).Infof("http ok: %v", resp.Status)
			}
		}
	}
}
