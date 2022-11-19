package main

import (
	"log"

	"fyne.io/fyne/v2/storage"
)

type CCReader struct {
	list *CEReadList
}

func NewCCReader() *CCReader {
	return &CCReader{
		list: NewCEMangaList(),
	}
}

func (cc *CCReader) LoadMangas() {
	mangaURIs, err := storage.List(controller.downloadURI)
	if err != nil {
		log.Panic(err)
	}
	for _, mangaURI := range mangaURIs {
		if canList, err := storage.CanList(mangaURI); err == nil && canList {
			portraitURI, err := storage.ParseURI("file://" + mangaURI.Path() + "/view.jpg")
			if err != nil {
				log.Panic(err)
			}
			if exists, err := storage.Exists(portraitURI); err != nil || !exists {
				portraitURI = nil
			}
			manga := NewMangaRead(mangaURI, portraitURI)
			chapterURIs, err := storage.List(mangaURI)
			if err != nil {
				log.Panic(err)
			}
			chapters := make([]*ChapterRead, 0, len(chapterURIs))
			for _, chapterURI := range chapterURIs {
				manga.AddChapters(NewChapterRead(chapterURI))
			}
			manga.AddChapters(chapters...)
			cc.list.AddMangas(manga)
		}
	}
}
