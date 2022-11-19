package main

import (
	"encoding/json"
	"fyne.io/fyne/v2/storage"
	"sync"

	fyne "fyne.io/fyne/v2"
)

type ChapterData struct {
	name                                          string       //readonly
	chapterURL                                    string       //readonly
	images                                        []*ImageData //readonly
	isFull                                        bool         //read and write
	count                                         int
	uri                                           fyne.URI
	fails                                         int
	isFullMut, uriMutex, countMutex, failedsMutex *sync.Mutex
}

func NewChapterData(chapter *Chapter) *ChapterData {
	return &ChapterData{
		name:         chapter.name,
		chapterURL:   chapter.chapterURL,
		images:       make([]*ImageData, 0),
		isFullMut:    &sync.Mutex{},
		uriMutex:     &sync.Mutex{},
		countMutex:   &sync.Mutex{},
		failedsMutex: &sync.Mutex{},
	}
}

func (c *ChapterData) MarshalJSON() ([]byte, error) {
	cd := struct {
		Name string `json:"name"`
		ChapterURL string `json:"chapter_url"`
		IsFull bool `json:"is_full"`
		Count int `json:"count"`
		Uri string `json:"uri"`
		Fails int `json:"fails"`
		Images []*ImageData `json:"images"`
	}{
		Name: c.name,
		ChapterURL: c.chapterURL,
		IsFull: c.isFull,
		Count: c.count,
		Uri: c.uri.String(),
		Fails: c.fails,
		Images: c.images,
	}
	return json.Marshal(cd)
}

func (c *ChapterData) UnmarshalJSON(data []byte) error {
	cd := struct {
		Name string `json:"name"`
		ChapterURL string `json:"chapter_url"`
		IsFull bool `json:"is_full"`
		Count int `json:"count"`
		Uri string `json:"uri"`
		Fails int `json:"fails"`
		Images []*ImageData `json:"images"`
	}{}
	if err := json.Unmarshal(data, &cd); err != nil {
		return err
	}
	c.name = cd.Name
	c.chapterURL = cd.ChapterURL
	c.isFull = cd.IsFull
	c.count = cd.Count
	uri, err := storage.ParseURI(cd.Uri)
	if err != nil {
		return err
	}
	c.uri = uri
	c.fails = cd.Fails
	c.images = cd.Images
	return nil
}

func (c *ChapterData) AddImage(image *ImageData) {
	c.images = append(c.images, image)
}

func (c *ChapterData) GetImages() []*ImageData {
	images := make([]*ImageData, len(c.images))
	copy(images, c.images)
	return images
}

func (c *ChapterData) IsFull() bool {
	c.isFullMut.Lock()
	defer c.isFullMut.Unlock()
	return c.isFull
}

func (c *ChapterData) SetIsFull(isFull bool) {
	c.isFullMut.Lock()
	defer c.isFullMut.Unlock()
	c.isFull = isFull
}

func (c *ChapterData) GetURI() fyne.URI {
	c.uriMutex.Lock()
	defer c.uriMutex.Unlock()
	return c.uri
}

func (c *ChapterData) SetURI(uri fyne.URI) {
	c.uriMutex.Lock()
	defer c.uriMutex.Unlock()
	c.uri = uri
}

func (c *ChapterData) GetCount() int {
	c.countMutex.Lock()
	defer c.countMutex.Unlock()
	return c.count
}

func (c *ChapterData) CountOne() {
	c.countMutex.Lock()
	defer c.countMutex.Unlock()
	c.count++
}

func (c *ChapterData) Reply() {
	c.countMutex.Lock()
	defer c.countMutex.Unlock()
	c.fails = 0
	c.count -= c.fails
}

func (c *ChapterData) SetCount(count int) {
	c.countMutex.Lock()
	defer c.countMutex.Unlock()
	c.count = count
}
