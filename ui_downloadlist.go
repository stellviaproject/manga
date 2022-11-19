package main

import (
	"image/color"
	"sync"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type UIDownloadList struct {
	co              *fyne.Container
	downloads       []*UIDownload
	mutex           *sync.Mutex
	downloadList    *widget.List
	pauseResumeTool *widget.ToolbarAction
	refMutex        *sync.Mutex
}

func NewUIDownloadList() *UIDownloadList {
	return &UIDownloadList{
		downloads: make([]*UIDownload, 0),
		mutex:     &sync.Mutex{},
		refMutex:  &sync.Mutex{},
	}
}

func (ui *UIDownloadList) MakeUI() fyne.CanvasObject {
	ui.pauseResumeTool = widget.NewToolbarAction(theme.MediaPauseIcon(), ui.PauseAll)
	replyTool := widget.NewToolbarAction(theme.MailReplyAllIcon(), ui.ReplyAll)
	removeTool := widget.NewToolbarAction(theme.ContentRemoveIcon(), ui.RemoveSelect)
	selectAll := widget.NewToolbarAction(theme.GridIcon(), ui.SelectAll)
	toolBar := widget.NewToolbar(ui.pauseResumeTool, replyTool, selectAll, removeTool)
	ui.downloadList = widget.NewList(ui.GetLen, ui.CreateDownload, ui.UpDateDownload)
	ui.co = container.New(layout.NewBorderLayout(toolBar, nil, nil, nil), toolBar, ui.downloadList)
	return ui.co
}

func (ui *UIDownloadList) SetLen(length int) {
	var cLength int
	ui.mutex.Lock()
	cLength = len(ui.downloads)
	if cLength < length {
		ui.downloads = append(ui.downloads, make([]*UIDownload, length-cLength)...)
	} else if cLength > length {
		i := 0
		for i < len(ui.downloads) {
			if ui.downloads[i].isRemoved {
				ui.downloads = append(ui.downloads[:i], ui.downloads[i+1:]...)
			} else {
				i++
			}
		}
	}
	ui.mutex.Unlock()
	ui.downloadList.Refresh()
}

func (ui *UIDownloadList) GetLen() int {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()
	return len(ui.downloads)
}

func (ui *UIDownloadList) SelectAll() {
	for _, download := range ui.downloads {
		download.isChecked = true
	}
	ui.co.Refresh()
}

func (ui *UIDownloadList) RemoveSelect() {
	removeds := make([]*MangaData, 0, len(ui.downloads))
	for _, d := range ui.downloads {
		if d.isChecked {
			removeds = append(removeds, d.data)
		}
	}
	controller.downloads.Remove(removeds)
}

func (ui *UIDownloadList) ReplyAll() {
	//controller.downloads.Reply()
}

func (ui *UIDownloadList) PauseAll() {
	controller.downloads.Pause()
	ui.pauseResumeTool.Icon = theme.MediaPlayIcon()
	ui.co.Refresh()
}

func (ui *UIDownloadList) ResumeAll() {
	controller.downloads.Resume()
	ui.pauseResumeTool.Icon = theme.MediaPauseIcon()
	ui.co.Refresh()
}

func (ui *UIDownloadList) CreateDownload() (co fyne.CanvasObject) {
	portrait := canvas.NewRectangle(color.NRGBA{128, 128, 128, 128})
	portrait.SetMinSize(fyne.NewSize(120, 160))

	imageCo := container.New(layout.NewHBoxLayout(), portrait)
	pauseResumeButton := widget.NewButtonWithIcon("", theme.MediaPauseIcon(), func() {})
	replyButton := widget.NewButtonWithIcon("", theme.MailReplyIcon(), func() {})
	removeButton := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {})
	selectCheck := widget.NewCheck("", func(selected bool) {})
	informationLabel := widget.NewLabel("Imágenes: %s/%d Progreso: %s %% Tamaño: %s.%s")
	nameLabel := widget.NewLabel("Leadale Daichi Nite")
	downloadProgress := widget.NewProgressBar()
	downloadProgress.Min, downloadProgress.Max = 0, 100
	readButton := widget.NewButtonWithIcon("Leer", theme.MediaPhotoIcon(), func() {})
	controls := container.New(layout.NewHBoxLayout(), selectCheck, pauseResumeButton, replyButton, removeButton, readButton)
	infoCo := container.New(layout.NewVBoxLayout(), nameLabel, controls, informationLabel, downloadProgress)
	fullContainer := container.New(layout.NewBorderLayout(nil, nil, imageCo, nil), imageCo, infoCo)
	return fullContainer
}

func (ui *UIDownloadList) UpDateDownload(index int, co fyne.CanvasObject) {
	if ui.downloads[index] == nil {
		ui.downloads[index] = NewUIDownload(index)
	}
	ui.downloads[index].UpDate(co)
}

func (ui *UIDownloadList) CanClose() bool {
	return false
}

func (ui *UIDownloadList) OnClose() {}
func (ui *UIDownloadList) OnShow() {}

func (ui *UIDownloadList) Refresh() {
	go func() {
		ui.refMutex.Lock()
		defer ui.refMutex.Unlock()
		ui.co.Refresh()
	}()
}
