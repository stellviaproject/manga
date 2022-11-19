package main

import "fyne.io/fyne/v2"

type UIBase interface {
	OnClose()
	OnShow()
	CanClose() bool
	MakeUI() fyne.CanvasObject
}
