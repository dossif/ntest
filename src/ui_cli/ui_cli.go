package ui_cli

import (
	"context"
	"fmt"
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
			msg := <-journal
			fmt.Println(fmt.Sprintf("print line: %v", msg))
			time.Sleep(time.Second * 1)
		}
	}
}
