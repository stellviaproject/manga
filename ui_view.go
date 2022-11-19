package main

import (
	"fmt"
	"image/color"
	"sync"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type UIView struct {
	sinopsisCo  *fyne.Container
	mangaImage  *canvas.Image
	fullLabel   *widget.Label
	chapterList []*UIChapter
	listMutex   *sync.Mutex
	co          *fyne.Container
	manga       *Manga
}

func NewUIView(manga *Manga) *UIView {
	return &UIView{
		chapterList: make([]*UIChapter, 0),
		listMutex:   new(sync.Mutex),
		manga:       manga,
	}
}

func (ui *UIView) Show() {
	controller.uiAppTabs.Show(ui)
}

func (ui *UIView) CanClose() bool {
	return true
}

func (ui *UIView) MakeUI() fyne.CanvasObject {
	portrait := canvas.NewRectangle(color.NRGBA{128, 128, 128, 128})
	portrait.SetMinSize(fyne.NewSize(120, 160))
	mangaName := ui.manga.GetName()
	ui.fullLabel = widget.NewLabel(MangaFullInfo(mangaName, "", ui.manga.GetLastChapter(), "", "", nil, nil))
	ui.fullLabel.Wrapping = fyne.TextWrapWord
	descriptionLabel := widget.NewLabel("Sinopsis:\n" + ui.manga.GetDescription())
	descriptionLabel.Wrapping = fyne.TextWrapWord
	chapterList := widget.NewList(ui.ChapterLen, ui.CreateChapter, ui.UpDateChapter)
	download := widget.NewToolbarAction(theme.DownloadIcon(), ui.Download)
	selectAll := widget.NewToolbarAction(theme.GridIcon(), ui.SelectAll)
	selectToBeginning := widget.NewToolbarAction(theme.ContentUndoIcon(), ui.SelectToBegining)
	selectToEnd := widget.NewToolbarAction(theme.ContentRedoIcon(), ui.SelectToEnd)
	portraitCo := container.New(layout.NewHBoxLayout(), portrait)
	toolBar := widget.NewToolbar(download, selectAll, selectToBeginning, selectToEnd)

	infoCo := container.New(layout.NewBorderLayout(ui.fullLabel, nil, nil, nil), ui.fullLabel, descriptionLabel)
	vScroll := container.NewVScroll(infoCo)
	ui.sinopsisCo = container.New(layout.NewBorderLayout(nil, nil, portraitCo, nil), portraitCo, vScroll)
	ui.co = container.New(layout.NewBorderLayout(ui.sinopsisCo, nil, nil, nil), ui.sinopsisCo, chapterList)
	return container.New(layout.NewBorderLayout(nil, toolBar, nil, nil), toolBar, ui.co)
}

func MangaFullInfo(name, alternatives, lastChapter, anio, state string, genres, authors []string) string {
	genresFull := ""
	if genres != nil {
		for i, c := 0, len(genres)-1; i < c; i++ {
			genresFull += genres[i] + ", "
		}
		if len(genres) > 0 {
			genresFull += genres[len(genres)-1]
		}
	}
	authorsFull := ""
	if authors != nil {
		for i, c := 0, len(authors)-1; i < c; i++ {
			authorsFull += authors[i] + ", "
		}
		if len(authors) > 0 {
			authorsFull += authors[len(authors)-1]
		}
	}
	return fmt.Sprintf("Nombre: %s\nAlternativa(s): %s\nGénero(s): %s\nAutor(es): %s\nAño: %s\nEstado: %s\nUltimo Capítulo: %s",
		name,
		alternatives,
		genresFull,
		authorsFull,
		anio,
		state,
		lastChapter,
	)
}

func (ui *UIView) UpDate(manga *Manga) {
	go func() {
		portrait := controller.GetPortrait(manga)
		if portrait != nil {
			ui.mangaImage = canvas.NewImageFromImage(portrait)
			ui.mangaImage.SetMinSize(fyne.NewSize(120, 160))
			ui.sinopsisCo.Objects[0].(*fyne.Container).Objects[0] = ui.mangaImage
			ui.sinopsisCo.Refresh()
		}
	}()
	ui.fullLabel.SetText(MangaFullInfo(manga.GetName(), manga.GetAlternatives(), manga.GetLastChapter(), manga.GetAnio(), manga.GetState(), manga.GetGenres(), manga.GetAuthors()))
	ui.listMutex.Lock()
	ui.chapterList = make([]*UIChapter, manga.Length())
	//checked = false para UIChapter por el momento hasta que se implemente la ui de descarga y su controller
	for i, ch := range manga.GetChapters() {
		ui.chapterList[i] = NewUIChapter(ch, false)
	}
	ui.listMutex.Unlock()
}

func (ui *UIView) ChapterLen() int {
	var length int
	ui.listMutex.Lock()
	length = len(ui.chapterList)
	ui.listMutex.Unlock()
	return length
}

func (ui *UIView) GetSelection() []*Chapter {
	selection := make([]*Chapter, 0, len(ui.chapterList))
	for _, c := range ui.chapterList {
		if c.checked {
			selection = append(selection, c.chapter)
		}
	}
	return selection
}

func (ui *UIView) CreateChapter() fyne.CanvasObject {
	downloadCheck := widget.NewCheck("", func(ch bool) {})
	viewButton := widget.NewButtonWithIcon("", theme.ConfirmIcon(), func() {})
	chapterInfo := widget.NewLabel("")
	leftCo := container.New(layout.NewHBoxLayout(), downloadCheck, viewButton)
	return container.New(layout.NewBorderLayout(nil, nil, leftCo, nil), leftCo, chapterInfo)
}

func (ui *UIView) UpDateChapter(index int, co fyne.CanvasObject) {
	ui.listMutex.Lock()
	border := co.(*fyne.Container)
	leftCo := border.Objects[0].(*fyne.Container)
	chapterInfo := border.Objects[1].(*widget.Label)
	ch := ui.chapterList[index]
	chapterInfo.SetText(ch.chapter.GetName() + "    " + ch.chapter.GetDate())
	downloadCheck := leftCo.Objects[0].(*widget.Check)
	downloadCheck.SetChecked(ch.checked)
	downloadCheck.OnChanged = ui.chapterList[index].OnChanged
	ui.listMutex.Unlock()
}

func (ui *UIView) Download() {
	controller.downloads.DoDownload(controller.view.GetManga(ui))
}

func (ui *UIView) SelectAll() {
	for _, ch := range ui.chapterList {
		ch.checked = true
	}
	ui.co.Refresh()
}

func (ui *UIView) SelectToBegining() {
	check := false
	for i := len(ui.chapterList) - 1; i >= 0; i-- {
		//Encontrar el que tiene checked=true para poner los demas true hasta el inicio
		if ui.chapterList[i].checked {
			check = true
		}
		//Chequear el resto si ya se encontro
		if check {
			ui.chapterList[i].checked = check
		}
	}
	if !check {
		ui.SelectAll()
	} else {
		ui.co.Refresh()
	}
}

func (ui *UIView) SelectToEnd() {
	check := false
	for _, ch := range ui.chapterList {
		if ch.checked {
			check = true
		}
		if check {
			ch.checked = check
		}
	}
	if !check {
		ui.SelectAll()
	} else {
		ui.co.Refresh()
	}
}

func (ui *UIView) OnClose() {
	if controller.view.HasView(ui) {
		controller.view.Remove(ui)
	}
}

func (ui *UIView) OnShow() {}
