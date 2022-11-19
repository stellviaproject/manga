package main

import (
	"strings"
	"sync"
	"time"

	fyne "fyne.io/fyne/v2"
)

type ChapterRead struct {
	chapterURI   fyne.URI
	isReaded     bool
	lastReaded   time.Time
	lastImage    fyne.URI
	images       []fyne.URI
	size         int
	name         string
	rwSize       *sync.RWMutex
	rwLastReaded *sync.RWMutex
	rwIsReaded   *sync.RWMutex
	rwName       *sync.RWMutex
}

func NewChapterRead(chapterURI fyne.URI) *ChapterRead {
	path := chapterURI.Path()
	lastIndex := strings.LastIndex(path, "/")
	name := path[lastIndex+1:]
	return &ChapterRead{
		chapterURI:   chapterURI,
		name:         name,
		rwSize:       &sync.RWMutex{},
		rwLastReaded: &sync.RWMutex{},
		rwIsReaded:   &sync.RWMutex{},
		rwName:       &sync.RWMutex{},
	}
}

func (ce *ChapterRead) Size() int {
	ce.rwSize.RLock()
	defer ce.rwSize.RUnlock()
	return ce.size
}

func (ce *ChapterRead) LastReaded() time.Time {
	ce.rwLastReaded.RLock()
	defer ce.rwLastReaded.RUnlock()
	return ce.lastReaded
}

func (ce *ChapterRead) IsReaded() bool {
	ce.rwIsReaded.RLock()
	defer ce.rwIsReaded.RUnlock()
	return ce.isReaded
}

func (ce *ChapterRead) Name() string {
	ce.rwName.RLock()
	defer ce.rwName.RUnlock()
	return ce.name
}
