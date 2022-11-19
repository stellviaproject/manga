package main

import "sync"

type CESearchList struct {
	list  []*Manga
	index map[string]int
	mutex *sync.Mutex
}

func NewCESearchList() *CESearchList {
	return &CESearchList{
		list:  make([]*Manga, 0, 1000),
		index: make(map[string]int),
		mutex: &sync.Mutex{},
	}
}

func (ls *CESearchList) Len() int {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	return len(ls.list)
}

func (ls *CESearchList) Register(manga *Manga) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	ls.index[manga.GetURL()] = len(ls.list)
	ls.list = append(ls.list, manga)
}

func (ls *CESearchList) Index(mangaURL string) (*Manga, bool) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	index, ok := ls.index[mangaURL]
	if !ok {
		return nil, ok
	}
	item := ls.list[index]
	return item, ok
}

func (ls *CESearchList) Get(index int) *Manga {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	item := ls.list[index]
	return item
}

func (ls *CESearchList) Clear() {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	ls.list = make([]*Manga, 0)
	ls.index = make(map[string]int)
}
