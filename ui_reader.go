package main

import (
	"fmt"
	"image/color"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type UIReader struct {
	list   *widget.List
	mangas []*UIMangaReader
}

func NewUIReader() *UIReader {
	return &UIReader{
		mangas: make([]*UIMangaReader, 0, 10),
	}
}

func (ui *UIReader) CanClose() bool {
	return false
}

func (ui *UIReader) OnClose() {}
func (ui *UIReader) OnShow() {}

func (ui *UIReader) MakeUI() fyne.CanvasObject {
	ui.list = widget.NewList(ui.Length, ui.CreateItem, ui.UpDateItem)
	return ui.list
}

func (ui *UIReader) Length() int {
	length, iLength := controller.reader.list.Length(), len(ui.mangas)
	if length > iLength {
		diff := length - iLength
		ui.mangas = append(ui.mangas, make([]*UIMangaReader, diff)...)
	}
	return len(ui.mangas)
}

func (ui *UIReader) CreateItem() fyne.CanvasObject {
	rect := canvas.NewRectangle(color.NRGBA{128, 128, 128, 128})
	rect.SetMinSize(fyne.NewSize(120, 160))
	information := widget.NewLabel("")
	open, remove, view := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {}), widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {}), widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {})
	options := container.New(layout.NewHBoxLayout(), open, remove, view)
	//right := container.New(layout.NewVBoxLayout(), information, options)
	all := container.New(layout.NewBorderLayout(nil, options, rect, nil), options, rect, information)
	return all
}

func (ui *UIReader) UpDateItem(index int, co fyne.CanvasObject) {
	if ui.mangas[index] == nil {
		ui.mangas[index] = NewUIMangaReader(index, co)
	} else {
		ui.mangas[index].UpDate(co)
	}
}

type UIMangaReader struct {
	image fyne.CanvasObject
	name  string
	count int
}

func NewUIMangaReader(index int, co fyne.CanvasObject) *UIMangaReader {
	read := controller.reader.list.Get(index)
	ui := &UIMangaReader{
		name:  read.GetName(),
		count: read.Count(),
	}
	image := read.GetImage()
	if image != nil {
		image := canvas.NewImageFromImage(image)
		image.SetMinSize(fyne.NewSize(120, 160))
		ui.image = image
	} else {
		rect := canvas.NewRectangle(color.NRGBA{128, 128, 128, 128})
		rect.SetMinSize(fyne.NewSize(120, 160))
		ui.image = rect
	}
	return ui
}

func (ui *UIMangaReader) UpDate(co fyne.CanvasObject) {
	all := co.(*fyne.Container)
	all.Objects[1] = ui.image
	all.Objects[2] = widget.NewLabel(ui.Information())
}

func (ui *UIMangaReader) Information() string {
	return fmt.Sprintf("%s Capítulos %d", ui.name, ui.count)
}
