package main

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type UIAppTabs struct {
	items   map[UIBase]*container.TabItem
	stack   []UIBase
	appTabs *container.AppTabs
}

func NewUIAppTabs() *UIAppTabs {
	return &UIAppTabs{
		items: map[UIBase]*container.TabItem{},
		stack: []UIBase{},
	}
}

func (ui *UIAppTabs) MakeUI() {
	controller.uiErrStack.MakeUI()
	ui.appTabs = container.NewAppTabs()
}

func (ui *UIAppTabs) Add(tittle string, uiItem UIBase, changeTo bool) *container.TabItem {
	var fc fyne.CanvasObject
	if uiItem.CanClose() {
		closeButton := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
			/*if len(ui.stack) > 0 {
				base := ui.stack[len(ui.stack)-1]
				ui.Show(base)
				ui.stack = ui.stack[:len(ui.stack)-1]
			}*/
			ui.Close(uiItem)
			uiItem.OnClose()
		})
		controls := container.New(layout.NewBorderLayout(nil, nil, closeButton, nil), closeButton, controller.uiErrStack.centerCo)
		fc = container.New(layout.NewBorderLayout(controls, nil, nil, nil), controls, uiItem.MakeUI())
	} else {
		fc = container.New(layout.NewBorderLayout(controller.uiErrStack.centerCo, nil, nil, nil), controller.uiErrStack.centerCo, uiItem.MakeUI())
	}
	item := container.NewTabItem(tittle, fc)
	ui.appTabs.Append(item)
	if changeTo {
		ui.appTabs.Select(item)
	}
	ui.items[uiItem] = item
	/*if changeTo {
		ui.stack = append(ui.stack, uiItem)
	}*/
	return item
}

func (ui *UIAppTabs) Show(uiItem UIBase) {
	uiItem.OnShow()
	ui.appTabs.Select(ui.items[uiItem])
}

func (ui *UIAppTabs) Close(uiItem UIBase) {
	ui.appTabs.Remove(ui.items[uiItem])
}
