package main

import (
	//"image/color"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	//"fyne.io/fyne/v2/theme"
	//"fyne.io/fyne/v2/canvas"
	//"fyne.io/fyne/v2/internal/widget"
	//"fyne.io/fyne/v2/storage"
)

func main() {
	a := app.NewWithID("Manga")
	w := a.NewWindow("Manga")
	NewCCMain() //Inicializa la controladora
	w.SetOnClosed(controller.Close)
	w.SetPadded(true)
	w.Resize(fyne.NewSize(700, 500))
	w.SetContent(controller.MainUI())
	w.ShowAndRun()
}
