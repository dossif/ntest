package ui

import (
	"context"
	"fmt"
	"pping/src/pinger"
	"pping/src/ui_cli"
	"sync"
)

type Ui interface {
	RenderUi(journal chan pinger.Ping) error
}

func NewUi(ctx context.Context, wg *sync.WaitGroup, name string) (Ui, error) {
	switch name {
	case "ui_cli":
		return ui_cli.NewUi(ctx, wg)
	default:
		return nil, fmt.Errorf("bad ui")
	}
}
