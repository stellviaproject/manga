package main

import (
	"sync"
	"time"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ErrorCallBack func()

type ErrorItem struct {
	err                     error
	callback                ErrorCallBack
	waitForError, autoClose bool
	condition               *sync.Cond
}

type UIErrStack struct {
	centerCo *fyne.Container
	stack    []*ErrorItem
	mutex    *sync.Mutex
}

func NewUIErrStack() *UIErrStack {
	return &UIErrStack{
		stack: make([]*ErrorItem, 0),
		mutex: &sync.Mutex{},
	}
}

func (ui *UIErrStack) MakeUI() {
	button := widget.NewButtonWithIcon("", theme.ConfirmIcon(), ui.Confirm)
	message := widget.NewLabel("")
	button.Hide()
	message.Hide()
	horizontal := container.New(layout.NewHBoxLayout(), message, button)
	ui.centerCo = container.New(layout.NewCenterLayout(), horizontal)
}

func (ui *UIErrStack) Confirm() {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	top := ui.stack[len(ui.stack)-1]
	ui.stack = ui.stack[0 : len(ui.stack)-1]

	top.condition.L.Lock()
	if top.waitForError {
		top.condition.Signal()
	}
	top.condition.L.Unlock()
	for len(ui.stack) > 0 {
		curr := ui.stack[len(ui.stack)-1]
		switch curr.err.(type) {
		case InternetConnectionError:
			switch top.err.(type) {
			case InternetConnectionError:
				curr.condition.L.Lock()
				if curr.waitForError {
					curr.condition.Signal()
				}
				curr.condition.L.Unlock()
				ui.stack = ui.stack[:len(ui.stack)-1]
			}
		}
	}

	if top.callback != nil {
		top.callback()
	}

	horizontal := ui.centerCo.Objects[0].(*fyne.Container)

	horizontal.Objects[0].Hide()
	horizontal.Objects[1].Hide()

	if len(ui.stack) > 0 {
		horizontal.Objects[0].(*widget.Label).SetText(ui.stack[len(ui.stack)-1].err.Error())
		ui.centerCo.Refresh()
		go func() {
			time.Sleep(time.Second * 1)
			horizontal.Objects[0].Show()
			horizontal.Objects[1].Show()
		}()
	}
}

func (ui *UIErrStack) NotifyError(err error, callback ErrorCallBack, waitForError, autoClose bool) {
	ui.mutex.Lock()
	item := &ErrorItem{
		err:          err,
		callback:     callback,
		waitForError: waitForError,
		autoClose:    autoClose,
		condition:    sync.NewCond(&sync.Mutex{}),
	}
	ui.stack = append(ui.stack, item)
	if len(ui.stack) == 1 {
		horizontal := ui.centerCo.Objects[0].(*fyne.Container)
		label := horizontal.Objects[0].(*widget.Label)
		label.SetText(item.err.Error())
		label.Show()
		button := horizontal.Objects[1].(*widget.Button)
		button.Show()
		ui.centerCo.Refresh()
	}
	ui.mutex.Unlock()

	if item.autoClose {
		go func() {
			time.Sleep(time.Second * 3)
			ui.Confirm()
		}()
	} else {
		item.condition.L.Lock()
		if waitForError {
			item.condition.Wait()
		}
		item.condition.L.Unlock()
	}
}
