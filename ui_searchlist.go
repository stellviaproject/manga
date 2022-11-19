package main

import (
	"image/color"
	"log"
	"sync"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type UISearchList struct {
	mutex      *sync.Mutex
	list       []*UIManga
	listWidget *widget.List
}

func NewUISearchList() *UISearchList {
	return &UISearchList{
		mutex: new(sync.Mutex),
		list:  make([]*UIManga, 0, 26),
	}
}

func (ui *UISearchList) SetList(listWidget *widget.List) {
	ui.listWidget = listWidget
}

func (ui *UISearchList) SetLen(length int) {
	log.Println("[UISearch.SetLen] setting length to search list...")
	ui.mutex.Lock()
	if length == 0 {
		ui.list = make([]*UIManga, 0, 26)
	} else if len(ui.list) < length {
		ui.list = append(ui.list, make([]*UIManga, length-len(ui.list))...)
	} else if len(ui.list) > length {
		i := 0
		for i < len(ui.list) {
			u := ui.list[i]
			if _, ok := controller.search.searchList.Index(u.mangaURL); !ok {
				ui.list = append(ui.list[:i], ui.list[i+1:]...)
			} else {
				i++
			}
		}
	}
	log.Println("[UISearch.SetLen] length set to search list successfully...")
	ui.mutex.Unlock()
	ui.listWidget.Refresh()
}

func (ui *UISearchList) Len() int {
	//log.Println("[UISearch.Len] getting search list length...")
	ui.mutex.Lock()
	length := len(ui.list)
	/*if ln := controller.search.searchList.Len(); ln > length {
		ui.list = append(ui.list, make([]*UIManga, ln-length)...)
	}*/
	ui.mutex.Unlock()
	//log.Println("[UISearch.Len] search list length successfully...")
	return length
}

func (ui *UISearchList) Create() fyne.CanvasObject {
	portrait := canvas.NewRectangle(color.NRGBA{128, 128, 128, 128})
	portrait.SetMinSize(fyne.NewSize(120, 160))
	nameLabel := widget.NewLabel("")
	descriptionLabel := widget.NewLabel("")
	chapterLabel := widget.NewLabel("")

	descriptionLabel.Wrapping = fyne.TextWrapWord
	nameLabel.TextStyle.Bold = true
	//nameLabel.Wrapping, , chapterLabel.Wrapping, fyne.TextWrapBreak, fyne.TextWrapBreak
	checkDownload := widget.NewCheck("", nil)
	downloadThis := widget.NewButtonWithIcon("", theme.DownloadIcon(), nil)
	viewThis := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), nil)
	//infoContainer := container.New(layout.NewMaxLayout(), descriptionLabel)
	desCo := container.NewVScroll(descriptionLabel)
	optCo := container.New(layout.NewVBoxLayout(), viewThis, checkDownload, downloadThis)
	infoCo := container.New(layout.NewBorderLayout(nameLabel, chapterLabel, nil, nil), nameLabel, chapterLabel, desCo)
	imgCo := container.New(layout.NewHBoxLayout(), portrait)
	return container.New(layout.NewBorderLayout(nil, nil, imgCo, optCo), imgCo, optCo, infoCo)
}

func (ui *UISearchList) Update(index int, co fyne.CanvasObject) {
	if ui.list[index] == nil {
		ui.list[index] = NewUIManga(index)
	}
	ui.list[index].UpDate(co)
	if index == ui.Len()-1 {
		controller.search.Next()
	}
}
