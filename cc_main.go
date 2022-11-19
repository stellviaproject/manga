package main

import (
	"image"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/PuerkitoBio/goquery"
)

var controller *CCMain

const (
	CATEGORY    = "https://es.ninemanga.com/category"
	POPULAR     = "https://es.ninemanga.com/list/Hot-Book/"
	NUEVOS      = "https://es.ninemanga.com/list/New-Book/"
	RECIENTES   = "https://es.ninemanga.com/list/New-Update/"
	SEARCH_BASE = "https://es.ninemanga.com/search/?wd="
	SITEURL     = "https://es.ninemanga.com"
)

//const CATEGORY = "http://127.0.0.1:8100/"

type CCMain struct {
	uiSearch       *UISearch
	uiSearchList   *UISearchList
	uiSearchButton *UISearchEntry
	uiOptions      *UIOptions
	uiDownloadList *UIDownloadList
	uiAppTabs      *UIAppTabs
	uiErrStack     *UIErrStack
	uiReader       *UIReader
	uiConfig       *UIConfig
	search         *CCSearch
	view           *CCView
	downloads      *CCDownload
	reader         *CCReader
	config         *CCConfig
	running        bool
	browser        *Browser
	downloadURI    fyne.URI
	fullInfoMutex  *sync.Mutex
}

func NewCCMain() {
	br := NewBrowser()
	controller = &CCMain{
		uiSearch:       NewUISearch(),
		uiSearchList:   NewUISearchList(),
		uiSearchButton: NewUISearchButton(),
		uiOptions:      NewUIOptions(),
		uiDownloadList: NewUIDownloadList(),
		uiErrStack:     NewUIErrStack(),
		uiAppTabs:      NewUIAppTabs(),
		uiReader:       NewUIReader(),
		search:         NewCCSearch(),
		view:           NewCCView(),
		downloads:      NewCCDownload(),
		reader:         NewCCReader(),
		config:         NewCCConfig(),
		running:        true,
		browser:        br,
		fullInfoMutex:  &sync.Mutex{},
	}
	controller.SetRoot()
	//cargar la configuracion
	controller.config.Load()
	proxy := controller.config.GetProxy()
	br.SetProxy(&proxy)
	controller.reader.LoadMangas()
	controller.MakeUI()
	//crear la interfaz de configuracion con lo cargado
	controller.uiConfig = NewUIConfig()
	go controller.search.ProccessPage(CATEGORY)
	log.Println("[Main] main controller loaded successfully...")
	controller.downloads.Proccess()
}

func (cc *CCMain) Close() {
	cc.browser.UnLoadCache()
	cc.config.Save()
}

func (s *CCMain) GetPortrait(manga *Manga) image.Image {
	//s.imageMutex.Lock()
	mangaImage := manga.GetImage()
	if mangaImage == nil {
		mangaImage = GetImage(manga.GetImageURL())
		manga.SetImage(mangaImage)
	}
	//s.imageMutex.Unlock()
	return mangaImage
}

func (c *CCMain) SetRoot() {
	root := fyne.CurrentApp().Storage().RootURI()
	log.Println("[CCMain.SetRoot] creating root download folder...")
	var err error
	if runtime.GOOS == "windows" {
		path := root.Path()
		appDataIndex := strings.Index(path, "AppData")
		c.downloadURI, err = storage.ParseURI("file://" + path[:appDataIndex] + "Downloads/Mangas")
		if err != nil {
			log.Panic(err)
		}
	} else {
		c.downloadURI = root
	}
	if exists, err := storage.Exists(c.downloadURI); err != nil {
		log.Panic(err)
	} else if !exists {
		if err = storage.CreateListable(c.downloadURI); err != nil {
			log.Panic(err)
		}
	}
}

func (cc *CCMain) MakeUI() fyne.CanvasObject {
	log.Println("[MakeUI] makeing user interface...")
	cc.uiAppTabs.MakeUI()
	cc.uiAppTabs.Add("Búsqueda", cc.uiSearch, true)
	cc.uiAppTabs.Add("Descargas", cc.uiDownloadList, false)
	//cc.uiAppTabs.Add("Lectura", cc.uiReader, false)
	log.Println("[MakeUI] user interface made successfully...")
	return cc.uiAppTabs.appTabs
}

func (cc *CCMain) MainUI() fyne.CanvasObject {
	return cc.uiAppTabs.appTabs
}

func (cc *CCMain) LoadFullInfo(manga *Manga) {
	cc.fullInfoMutex.Lock()
	view := GetPage(manga.GetURL())
	bookintro := view.Find("div.bookintro")
	if bookintro.Length() > 0 {
		items := bookintro.Find("ul.message > li")
		field := items.Find("ul.message > li > b")
		for i, c := 0, items.Length(); i < c; i++ {
			item := items.Eq(i)
			inner, _ := field.Html()
			switch inner {
			case "Alternativa(s):":
				alternatives, _ := item.Last().Html()
				manga.SetAlternatives(alternatives)
			case "Género(s):":
				genres := item.Find("a[href]")
				for i, c := 0, genres.Length(); i < c; i++ {
					genre, _ := genres.Eq(i).Html()
					manga.AddGenre(genre)
				}
			case "Autor(s):":
				authors := item.Find("a[href]")
				for i, c := 0, authors.Length(); i < c; i++ {
					author, _ := authors.Eq(i).Html()
					manga.AddAuthor(author)
				}
			case "Año":
				anio, _ := item.Find("a[href]").Html()
				manga.SetAnio(anio)
			case "Estado:":
				state, _ := item.Find("a[href]").Html()
				manga.SetState(state)
			}
		}
	}
	warning := view.Find("div.warning > a")
	if warning.Length() > 0 {
		href, _ := warning.Attr("href")
		view = GetPage(href)
	}
	chapters := view.Find("li > a.chapter_list_a")
	chapterSpan := view.Find("div.silde > ul > li > span")
	if chapters.Length() > 0 {
		for i, c := 0, chapters.Length(); i < c; i++ {
			span := chapterSpan.Eq(i)
			chapter := chapters.Eq(i)
			url, _ := chapter.Attr("href")
			title, _ := chapter.Attr("title")
			date, _ := span.Html()
			manga.AddChapter(NewChapter(title, url, date))
		}
	}
	manga.Reverse()
	manga.SetFullInfo()
	cc.fullInfoMutex.Unlock()
}

func GetImage(imageURL string) image.Image {
	for {
		i, err := controller.browser.Get(imageURL)
		if err == nil {
			if img, ok := i.(image.Image); ok {
				return img
			}
			return nil
		} else if _, ok := err.(FormatError); ok {
			return nil
		}
		time.Sleep(time.Second)
	}
}

func GetPage(u string) *goquery.Document {
	for {
		i, err := controller.browser.Get(u)
		if err != nil {
			controller.uiErrStack.NotifyError(InternetConnectionError{"verify internet connection"}, nil, true, false)
		} else if doc, ok := i.(*goquery.Document); ok {
			return doc
		}
	}
}

func ForcePage(u string) *goquery.Document {
	for {
		i, err := controller.browser.Get(u)
		if err != nil {
			time.Sleep(time.Second)
		} else if doc, ok := i.(*goquery.Document); ok {
			return doc
		}
	}
}

type InternetConnectionError struct {
	message string
}

func (e InternetConnectionError) Error() string {
	return e.message
}
