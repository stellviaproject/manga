package main

import (
	"fmt"
	"strings"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type UIDownload struct {
	portrait                       *canvas.Image
	data                           *MangaData
	index                          int
	isChecked, isPaused, isRemoved bool
}

func NewUIDownload(index int) *UIDownload {
	data := controller.downloads.Get(index)
	var portrait *canvas.Image = nil
	if data.image != nil {
		portrait = canvas.NewImageFromImage(data.image)
		portrait.SetMinSize(fyne.NewSize(120, 160))
	}
	return &UIDownload{index: index, data: data, portrait: portrait}
}

func (ui *UIDownload) UpDate(co fyne.CanvasObject) {
	fullContainer := co.(*fyne.Container)

	imageCo := fullContainer.Objects[0].(*fyne.Container)
	infoCo := fullContainer.Objects[1].(*fyne.Container)

	nameLabel := infoCo.Objects[0].(*widget.Label)
	controls := infoCo.Objects[1].(*fyne.Container)
	informationLabel := infoCo.Objects[2].(*widget.Label)
	downloadProgress := infoCo.Objects[3].(*widget.ProgressBar)

	selectCheck := controls.Objects[0].(*widget.Check)
	pauseResumeButton := controls.Objects[1].(*widget.Button)
	replyButton := controls.Objects[2].(*widget.Button)
	removeButton := controls.Objects[3].(*widget.Button)
	readButton := controls.Objects[4].(*widget.Button)

	if ui.portrait != nil {
		imageCo.Objects[0] = ui.portrait
	}
	nameLabel.SetText(ui.data.name)
	informationLabel.SetText(ui.GetInformation())

	downloadProgress.Min, downloadProgress.Max = 0, float64(ui.data.GetFullCount())
	downloadProgress.Value = float64(ui.data.GetCount())
	downloadProgress.Refresh()

	if ui.isPaused {
		pauseResumeButton.Icon = theme.MediaPlayIcon()
	} else {
		pauseResumeButton.Icon = theme.MediaPauseIcon()
	}

	selectCheck.OnChanged = ui.Select
	pauseResumeButton.OnTapped = ui.Pause
	replyButton.OnTapped = ui.Reply
	removeButton.OnTapped = ui.Remove
	readButton.OnTapped = ui.Read
}

func (ui *UIDownload) GetInformation() string {
	imagesCount, fullCount := ui.data.GetCount(), ui.data.GetFullCount()
	countSTR, fullSTR := fmt.Sprintf("%d", imagesCount), fmt.Sprintf("%d", fullCount)
	if len(fullSTR) > len(countSTR) {
		countSTR = strings.Repeat(" ", len(fullSTR)-len(countSTR)) + countSTR
	}
	percent := int(100.0 * float32(imagesCount) / float32(fullCount))
	var percentSTR string
	if percent < 10 {
		percentSTR = "  " + fmt.Sprintf("%d", percent)
	} else if percent < 100 {
		percentSTR = " " + fmt.Sprintf("%d", percent)
	} else {
		percentSTR = fmt.Sprintf("%d", percent)
	}
	return fmt.Sprintf("Imágenes: %s/%s Progreso: %s %%", countSTR, fullSTR, percentSTR)
}

func (ui *UIDownload) Pause() {
	controller.downloads.PauseManga(ui.data)
	ui.isPaused = true
}

func (ui *UIDownload) Remove() {
	controller.downloads.RemoveManga(ui.data)
	ui.isRemoved = true
}

func (ui *UIDownload) Reply() {
	//controller.downloads.ReplyManga(ui.data)
}

func (ui *UIDownload) Resume() {
	controller.downloads.ResumeManga(ui.data)
	ui.isPaused = false
}

func (ui *UIDownload) Select(checked bool) {
	ui.isChecked = checked
}

func (ui *UIDownload) Read() {

}
