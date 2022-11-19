package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type UISearch struct {
	avancedOptions *UIAvancedOptions
}

func NewUISearch() *UISearch {
	return &UISearch{avancedOptions: NewUIAvancedOptions()}
}

func (ui *UISearch) CanClose() bool {
	return false
}

func (ui *UISearch) OnClose() {}
func (ui *UISearch) OnShow()  {}

func (ui *UISearch) MakeUI() fyne.CanvasObject {
	controller.uiSearchButton.MakeUI()
	//listOptions := widget.NewRadioGroup([]string{"Todos", "Populares", "Nuevos", "Recientes"}, controller.uiOptions.OnSelect)
	mangaList := widget.NewList(controller.uiSearchList.Len, controller.uiSearchList.Create, controller.uiSearchList.Update) //Listado de Mangas
	//listOptions.Selected = "Todos"
	//optionsContainer := container.NewVScroll(container.New(layout.NewVBoxLayout() /*listOptions,*/, ui.avancedOptions.MakeUI()))
	workAreaContainer := container.New(layout.NewBorderLayout(nil, nil, nil /*optionsContainer*/, nil) /*optionsContainer,*/, mangaList)
	controller.uiSearchList.SetList(mangaList)
	return container.New(layout.NewBorderLayout(controller.uiSearchButton.container, nil, nil, nil), controller.uiSearchButton.container, workAreaContainer)
}
