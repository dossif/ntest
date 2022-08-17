package icmpping

import (
	"context"
	"fmt"
	"github.com/digineo/go-ping"
	log "github.com/sirupsen/logrus"
	"net"
	"pping/src/dns"
	"time"
)

type Test struct {
	Api     ping.Pinger
	Host    string
	Ip      net.IPAddr
	Timeout time.Duration
}

//type Ping struct {
//	Host     string
//	Ip       string
//	Sequence int
//	Status   bool
//	Rtt      time.Duration
//	Error    error
//}
//
//type Pinger struct {
//	Api *goping.Pinger
//	Log *log.Logger
//	Ctx context.Context
//	Wg  *sync.WaitGroup
//}

func NewTest(bind string, host string, timeout int) *Test {
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
		Api:     *api,
		Host:    host,
		Ip:      *hostIp,
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
}

func (t *Test) Execute(ctx context.Context) error {
	var seq int
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			rtt, err := t.Api.Ping(&t.Ip, t.Timeout)
			seq = seq + 1
			cf := new(log.TextFormatter)
			cf.FullTimestamp = true
			log.SetFormatter(cf)
			lg := log.Fields{
				"seq":  seq,
				"dest": fmt.Sprintf("%v (%v)", t.Host, t.Ip.IP),
				"rtt":  rtt,
			}
			if err != nil {
				log.WithFields(lg).Errorf("ping error: %v", err)
			} else {
				log.WithFields(lg).Infof("ping ok")
			}
			time.Sleep(time.Second * 1)
		}
	}
}
