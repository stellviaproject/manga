package main

import (
	"image/color"
	"sync"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type UIManga struct {
	portrait                             fyne.CanvasObject
	portraitChannel                      chan fyne.CanvasObject
	index                                int
	name, chapter, description, mangaURL string
	checked                              bool
	mutex, portraitMutex                 *sync.Mutex
}

func NewUIManga(index int) *UIManga {
	um := new(UIManga)
	um.index = index
	manga := controller.search.Get(index)
	um.name = manga.GetName()
	um.mangaURL = manga.GetURL()
	um.chapter = manga.GetLastChapter()
	um.description = manga.GetDescription()
	um.mutex = &sync.Mutex{}
	um.portraitMutex = &sync.Mutex{}
	um.portraitChannel = make(chan fyne.CanvasObject)
	go func() {
		/*um.mutex.Lock()
		if !um.loaded {

			um.loaded = true
		}
		um.mutex.Unlock()*/
		img := controller.GetPortrait(controller.search.Get(index))
		if img != nil {
			portrait := canvas.NewImageFromImage(img)
			portrait.SetMinSize(fyne.NewSize(120, 160))
			portrait.FillMode = canvas.ImageFillContain
			um.portraitChannel <- portrait
		} else {
			portrait := canvas.NewRectangle(color.NRGBA{128, 128, 128, 128})
			portrait.SetMinSize(fyne.NewSize(120, 160))
			um.portraitChannel <- portrait
		}
	}()
	return um
}

func (ui *UIManga) UpDate(co fyne.CanvasObject) {
	c := co.(*fyne.Container)
	imageCo := c.Objects[0].(*fyne.Container)
	ui.portraitMutex.Lock()
	if ui.portrait == nil {
		ui.portraitMutex.Unlock()
		rect := canvas.NewRectangle(color.NRGBA{128, 128, 128, 128})
		rect.SetMinSize(fyne.NewSize(120, 160))
		imageCo.Objects[0] = rect
		go func() {
			if tempPortrait, ok := <-ui.portraitChannel; ok {
				close(ui.portraitChannel)

				ui.portraitMutex.Lock()
				ui.portrait = tempPortrait
				ui.portraitMutex.Unlock()

				imageCo.Objects[0] = ui.portrait
				controller.uiSearchList.listWidget.Refresh()
			}
		}()
	} else {
		imageCo.Objects[0] = ui.portrait
		ui.portraitMutex.Unlock()
	}

	optCo := c.Objects[1].(*fyne.Container)
	infoCo := c.Objects[2].(*fyne.Container)

	nameLabel := infoCo.Objects[0].(*widget.Label)
	chapterLabel := infoCo.Objects[1].(*widget.Label)
	descriptionLabel := infoCo.Objects[2].(*container.Scroll).Content.(*widget.Label)

	nameLabel.SetText(ui.name)
	chapterLabel.SetText(ui.chapter)
	descriptionLabel.SetText(ui.description)

	viewButton := optCo.Objects[0].(*widget.Button)
	downloadCheck := optCo.Objects[1].(*widget.Check)
	downloadButton := optCo.Objects[2].(*widget.Button)

	viewButton.OnTapped = ui.View
	downloadCheck.OnChanged = ui.Check
	downloadButton.OnTapped = ui.Download
}

func (ui *UIManga) Download() {
	controller.downloads.DoDownload(controller.search.Get(ui.index))
}

func (ui *UIManga) Check(checked bool) {
	ui.checked = checked
	controller.uiSearchList.listWidget.Refresh()
}

func (ui *UIManga) View() {
	controller.view.Show(controller.search.Get(ui.index))
}
