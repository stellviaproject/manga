package main

import (
	"image"
	"strings"
	"sync"
	"time"
	"unicode"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

type MangaRead struct {
	mangaURI      fyne.URI
	chapters      []*ChapterRead
	size          int
	portrait      image.Image
	portraitURI   fyne.URI
	chaptersMutex *sync.RWMutex
	name          string
	nameRW        *sync.RWMutex
	portraitRW    *sync.RWMutex
}

func NewMangaRead(mangaURI fyne.URI, portraitURI fyne.URI) *MangaRead {
	path := mangaURI.Path()
	return &MangaRead{
		name:          Title(path[strings.LastIndex(path, "/")+1:]),
		mangaURI:      mangaURI,
		portraitURI:   portraitURI,
		chaptersMutex: &sync.RWMutex{},
		nameRW:        &sync.RWMutex{},
		portraitRW:    &sync.RWMutex{},
	}
}

func Title(text string) string {
	r := ""
	text = strings.Trim(text, " ")
	text = strings.Replace(text, "_", " ", -1)
	ws := strings.Split(text, " ")
	for _, w := range ws {
		if l := len(w); l > 2 {
			r += string(unicode.ToUpper(rune(text[0]))) + strings.ToLower(text[1:]) + " "
		} else if l > 0 {
			r += strings.ToLower(w) + " "
		}
	}
	r = strings.Trim(r, " ")
	return r
}

func (ce *MangaRead) LastChapter() *ChapterRead {
	ce.chaptersMutex.RLock()
	defer ce.chaptersMutex.RUnlock()
	if len(ce.chapters) == 0 {
		return nil
	}
	last := ce.chapters[0]
	for _, c := range ce.chapters {
		if c.IsReaded() && time.Since(c.LastReaded()) < time.Since(last.LastReaded()) {
			last = c
		}
	}
	return last
}

func (ce *MangaRead) AddChapters(chapters ...*ChapterRead) {
	ce.chaptersMutex.Lock()
	defer ce.chaptersMutex.Unlock()
	ce.chapters = append(ce.chapters, chapters...)
	for _, c := range chapters {
		ce.size += c.Size()
	}
}

func (ce *MangaRead) GetName() string {
	ce.nameRW.Lock()
	defer ce.nameRW.Unlock()
	return ce.name
}

func (ce *MangaRead) Count() int {
	ce.chaptersMutex.RLock()
	defer ce.chaptersMutex.RUnlock()
	return len(ce.chapters)
}

func (ce *MangaRead) GetImage() image.Image {
	ce.portraitRW.Lock()
	defer ce.portraitRW.Unlock()
	if ce.portrait == nil && ce.portraitURI != nil {
		reader, err := storage.Reader(ce.portraitURI)
		if err != nil {
			return nil
		}
		defer reader.Close()
		ce.portrait, _, err = image.Decode(reader)
		if err != nil {
			return nil
		}
	}
	return ce.portrait
}
