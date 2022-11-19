package main

import (
	"encoding/json"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"sync"

	"fyne.io/fyne/v2/storage"

	fyne "fyne.io/fyne/v2"
)

type MangaData struct {
	mangaURL                            string         //readonly
	name                                string         //readonly
	description                         string         //readonly
	chapter                             string         //readonly
	alternatives                        string         //readonly
	genres                              []string       //readonly
	authors                             []string       //readonly
	year                                string         //readonly
	state                               string         //readonly
	chapters                            []*ChapterData //read and write
	count, fullCount, fails             int            //read and write
	image                               image.Image    //readonly
	indexURL                            map[string]int //read and write
	chaptersMutex, countMutex, uriMutex *sync.Mutex
	uri                                 fyne.URI
}

func NewMangaData(manga *Manga) *MangaData {
	return &MangaData{
		mangaURL:      manga.mangaURL,
		name:          manga.name,
		description:   manga.description,
		chapter:       manga.lastChapter,
		alternatives:  manga.alternatives,
		genres:        manga.genres,
		authors:       manga.authors,
		year:          manga.anio,
		state:         manga.state,
		image:         manga.GetImage(),
		chapters:      make([]*ChapterData, 0),
		indexURL:      make(map[string]int),
		countMutex:    &sync.Mutex{},
		chaptersMutex: &sync.Mutex{},
		uriMutex:      &sync.Mutex{},
	}
}

func (manga *MangaData) UnmarshalJSON(data []byte) error {
	m := struct {
		MangaURL     string         `json:"manga_url"`
		Name         string         `json:"name"`
		Description  string         `json:"description"`
		Chapter      string         `json:"chapter"`
		Alternatives string         `json:"alternatives"`
		Genres       []string       `json:"genres"`
		Authors      []string       `json:"authors"`
		Year         string         `json:"year"`
		State        string         `json:"state"`
		Chapters     []*ChapterData `json:"chapters"`
		Count        int            `json:"count"`
		FullCount    int            `json:"full_count"`
		Fails        int            `json:"fails"`
		Image        string         `json:"image"`
		IndexURL     map[string]int `json:"index"`
		Uri          string         `json:"uri"`
	}{}
	if err := json.Unmarshal(data, &m); err != nil {
		log.Println(err)
		return err
	}
	manga.countMutex = new(sync.Mutex)
	manga.chaptersMutex = new(sync.Mutex)
	manga.uriMutex = new(sync.Mutex)
	manga.mangaURL = m.MangaURL
	manga.name = m.Name
	manga.description = m.Description
	manga.chapter = m.Chapter
	manga.alternatives = m.Alternatives
	manga.genres = m.Genres
	manga.authors = m.Authors
	manga.year = m.Year
	manga.state = m.State
	manga.chapters = m.Chapters
	manga.count = m.Count
	manga.fails = m.Fails
	manga.fullCount = m.FullCount
	manga.indexURL = m.IndexURL
	uri, err := storage.ParseURI(m.Uri)
	if err != nil {
		log.Println(err)
		return err
	}
	manga.uri = uri
	if m.Image != "" {
		imageURI, err := storage.ParseURI(m.Image)
		if err != nil {
			log.Println(err)
			return err
		}
		reader, err := storage.Reader(imageURI)
		if err != nil {
			log.Println(err)
			return err
		}
		image, err := jpeg.Decode(reader)
		if err != nil {
			log.Println(err)
			return err
		}
		manga.image = image
	}
	return nil
}

func (manga *MangaData) MarshalJSON() ([]byte, error) {
	manga.chaptersMutex.Lock()
	manga.countMutex.Lock()
	manga.uriMutex.Lock()
	defer manga.chaptersMutex.Unlock()
	defer manga.countMutex.Unlock()
	defer manga.uriMutex.Unlock()
	imagePath := ""
	if manga.image != nil {
		imagePath = manga.uri.Path() + "/image.jpg"
	}
	m := struct {
		MangaURL     string         `json:"manga-url"`
		Name         string         `json:"name"`
		Description  string         `json:"description"`
		Chapter      string         `json:"chapter"`
		Alternatives string         `json:"alternatives"`
		Genres       []string       `json:"genres"`
		Authors      []string       `json:"authors"`
		Year         string         `json:"year"`
		State        string         `json:"state"`
		Chapters     []*ChapterData `json:"chapters"`
		Count        int            `json:"count"`
		FullCount    int            `json:"full-count"`
		Fails        int            `json:"fails"`
		Image        string         `json:"image"`
		IndexURL     map[string]int `json:"index"`
		Uri          string         `json:"uri"`
	}{
		MangaURL:     manga.mangaURL,
		Name:         manga.name,
		Description:  manga.description,
		Chapter:      manga.chapter,
		Alternatives: manga.alternatives,
		Genres:       manga.genres,
		Authors:      manga.authors,
		Year:         manga.year,
		State:        manga.state,
		Chapters:     manga.chapters,
		Count:        manga.count,
		FullCount:    manga.fullCount,
		Fails:        manga.fails,
		Image:        imagePath,
		IndexURL:     manga.indexURL,
		Uri:          manga.uri.String(),
	}
	return json.Marshal(m)
}

func (manga *MangaData) Load(folderPath string) error {
	infoURI, err := storage.ParseURI("file://" + folderPath + "/info.json")
	if err != nil {
		return err
	}
	reader, err := storage.Reader(infoURI)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, manga); err != nil {
		return err
	}
	return nil
}

func (manga *MangaData) Save() error {
	if manga.image != nil {
		imageURI, err := storage.ParseURI(manga.uri.Path() + "image.jpg")
		if err != nil {
			return err
		}
		imageWriter, err := storage.Writer(imageURI)
		defer func() {
			if err := imageWriter.Close(); err != nil {
				log.Println(err)
			}
		}()
		if err != nil {
			return err
		}
		if err := jpeg.Encode(imageWriter, manga.image, &jpeg.Options{Quality: 90}); err != nil {
			return err
		}
	}
	infoURI, err := storage.ParseURI(manga.uri.Path() + "/info.json")
	if err != nil {
		return err
	}
	infoWriter, err := storage.Writer(infoURI)
	defer func() {
		if err := infoWriter.Close(); err != nil {
			log.Println(err)
		}
	}()
	if err != nil {
		return err
	}
	data, err := json.Marshal(manga)
	if err != nil {
		return err
	}
	_, err = infoWriter.Write(data)
	if err != nil {
		return err
	}
	return err
}

func (manga *MangaData) SetChapters(chapters []*ChapterData) {
	manga.chaptersMutex.Lock()
	manga.countMutex.Lock()
	defer manga.countMutex.Unlock()
	defer manga.chaptersMutex.Unlock()
	manga.indexURL = make(map[string]int)
	manga.chapters = chapters
	for i, c := range chapters {
		manga.indexURL[c.chapterURL] = i
		manga.fullCount += len(c.images)
	}
}

func (manga *MangaData) GetChapters() []*ChapterData {
	manga.chaptersMutex.Lock()
	defer manga.chaptersMutex.Unlock()
	chapters := make([]*ChapterData, len(manga.chapters))
	copy(chapters, manga.chapters)
	return chapters
}

func (manga *MangaData) AddChapter(chapter *ChapterData) {
	manga.chaptersMutex.Lock()
	manga.countMutex.Lock()
	defer manga.countMutex.Unlock()
	defer manga.chaptersMutex.Unlock()
	index := len(manga.chapters)
	manga.chapters = append(manga.chapters, chapter)
	manga.indexURL[chapter.chapterURL] = index
	manga.fullCount += len(chapter.images)
}

func (manga *MangaData) Index(chapterURL string) (*ChapterData, bool) {
	manga.chaptersMutex.Lock()
	defer manga.chaptersMutex.Unlock()
	dx, ok := manga.indexURL[chapterURL]
	if ok {
		return manga.chapters[dx], ok
	}
	return nil, ok
}

func (manga *MangaData) GetCount() int {
	manga.countMutex.Lock()
	defer manga.countMutex.Unlock()
	return manga.count
}

func (manga *MangaData) SetCount(count int) {
	manga.countMutex.Lock()
	defer manga.countMutex.Unlock()
	manga.count = count
}

func (manga *MangaData) CountOne() {
	manga.countMutex.Lock()
	defer manga.countMutex.Unlock()
	manga.count++
}

func (manga *MangaData) GetFullCount() int {
	manga.countMutex.Lock()
	defer manga.countMutex.Unlock()
	return manga.fullCount
}

func (manga *MangaData) GetURI() fyne.URI {
	manga.uriMutex.Lock()
	defer manga.uriMutex.Unlock()
	return manga.uri
}

func (manga *MangaData) SetURI(uri fyne.URI) {
	manga.uriMutex.Lock()
	defer manga.uriMutex.Unlock()
	manga.uri = uri
}
