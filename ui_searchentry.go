package main

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type UISearchEntry struct {
	container   *fyne.Container
	searchEntry *widget.Entry
}

func NewUISearchButton() *UISearchEntry {
	return &UISearchEntry{}
}

func (ui *UISearchEntry) MakeUI() {
	ui.searchEntry = widget.NewEntry() //Entrada de texto de busqueda
	ui.searchEntry.OnSubmitted = ui.OnSubmitted
	searchButton := widget.NewButtonWithIcon("", theme.SearchIcon(), controller.uiSearchButton.OnClick) //Lista de Opciones
	searchButton.Resize(fyne.NewSize(30, 30))
	configButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		controller.uiAppTabs.Add("Configuración", controller.uiConfig, true)
	})
	hButtons := container.New(layout.NewHBoxLayout(), configButton, searchButton)
	ui.container = container.New(layout.NewBorderLayout(nil, nil, hButtons, nil), hButtons, ui.searchEntry)
}

func (ui *UISearchEntry) OnClick() {
	controller.search.Search(ui.searchEntry.Text)
}

func (ui *UISearchEntry) OnSubmitted(text string) {
	controller.search.Search(ui.searchEntry.Text)
}
