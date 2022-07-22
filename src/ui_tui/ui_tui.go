package ui_tui

type Ui struct {
	Region string
}

func NewUi() (*Ui, error) {
	return &Ui{
		Region: "",
	}, nil
}

func (ui Ui) RenderUi() error {
	return nil
}
