package tui

// https://github.com/gizak/termui

import (
	tui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
)

func StartTui() {
	err := tui.Init()
	if err != nil {
		log.Fatalf("failed to create ui interface: %v", err)
	}
	defer tui.Close()
	p := widgets.NewParagraph()
	p.Text = "Hello World!"
	p.SetRect(0, 0, 25, 5)
	
	tui.Render(p)
	
	for e := range tui.PollEvents() {
		if e.Type == tui.KeyboardEvent {
			break
		}
	}
}
