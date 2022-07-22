package ui_cli

import (
	"context"
	"fmt"
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

func (ui Ui) RenderUi() error {
	defer ui.Wg.Done()
	for {
		select {
		case <-ui.Ctx.Done():
			return nil
		default:
			fmt.Println("print line")
			time.Sleep(time.Second * 1)
		}
	}
}
