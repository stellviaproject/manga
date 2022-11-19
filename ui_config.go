package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

//Interfaces de Usuario

type UIConfig struct {
	proxy *UIProxy
}

func NewUIConfig() *UIConfig {
	return &UIConfig{
		proxy: NewUIProxy(),
	}
}

func (ui *UIConfig) CanClose() bool {
	return true
}

func (ui *UIConfig) OnClose() {
	controller.config.SetProxy(ui.proxy.user, ui.proxy.password, ui.proxy.url, ui.proxy.useSystemProxy)
}

func (ui *UIConfig) OnShow() {}

func (ui *UIConfig) MakeUI() fyne.CanvasObject {
	cfg := container.NewGridWithRows(2,
		ui.proxy.MakeUI(),
	)
	return cfg
}
