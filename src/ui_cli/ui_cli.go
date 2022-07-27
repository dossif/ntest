package ui_cli

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"pping/src/pinger"
	"sync"
	"time"
)

type Ui struct {
	Ctx context.Context
	Wg  *sync.WaitGroup
}

func NewUi(ctx context.Context, wg *sync.WaitGroup) (*Ui, error) {
	return &Ui{
		Ctx: ctx,
		Wg:  wg,
	}, nil
}

func (ui *Ui) RenderUi(journal chan pinger.Ping) error {
	defer ui.Wg.Done()
	defer fmt.Println("exit from ui")
	for {
		select {
		case <-ui.Ctx.Done():
			return nil
		default:
			ping := <-journal
			cf := new(log.TextFormatter)
			cf.FullTimestamp = true
			log.SetFormatter(cf)
			lg := log.Fields{
				"seq":  ping.Sequence,
				"dest": fmt.Sprintf("%v (%v)", ping.Host, ping.Ip),
				"rtt":  ping.Rtt,
			}
			if ping.Status == true {
				log.WithFields(lg).Info("ping is successful")
			} else {
				log.WithFields(lg).Warn(ping.Error.Error())
			}
			time.Sleep(time.Second * 1)
		}
	}
}
