package main

import "sync"

type CEReadList struct {
	list []*MangaRead
	rw   *sync.RWMutex
}

func NewCEMangaList() *CEReadList {
	return &CEReadList{
		list: make([]*MangaRead, 0, 10),
		rw:   &sync.RWMutex{},
	}
}

func (ce *CEReadList) Get(index int) *MangaRead {
	ce.rw.RLock()
	defer ce.rw.RUnlock()
	return ce.list[index]
}

func (ce *CEReadList) GetMangas() []*MangaRead {
	ce.rw.RLock()
	defer ce.rw.RUnlock()
	mangas := make([]*MangaRead, len(ce.list))
	copy(mangas, ce.list)
	return mangas
}

func (ce *CEReadList) Length() int {
	ce.rw.RLock()
	defer ce.rw.RUnlock()
	return len(ce.list)
}

func (ce *CEReadList) AddMangas(mangas ...*MangaRead) {
	ce.rw.Lock()
	defer ce.rw.Unlock()
	ce.list = append(ce.list, mangas...)
}
