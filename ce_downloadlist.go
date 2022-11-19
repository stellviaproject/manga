package main

import "sync"

type CEDownloadList struct {
	list     []*MangaData
	indexURL map[string]int
	mutex    *sync.Mutex
}

func NewCEDownloadList() *CEDownloadList {
	return &CEDownloadList{
		list:     make([]*MangaData, 0),
		indexURL: make(map[string]int),
		mutex:    &sync.Mutex{},
	}
}

func (ls *CEDownloadList) Len() int {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	return len(ls.list)
}

func (ls *CEDownloadList) Add(data *MangaData) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	ls.indexURL[data.mangaURL] = len(ls.list)
	ls.list = append(ls.list, data)
}

func (ls *CEDownloadList) Get(index int) *MangaData {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	return ls.list[index]
}

func (ls *CEDownloadList) GetMangas() []*MangaData {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	mangas := make([]*MangaData, len(ls.list))
	copy(mangas, ls.list)
	return mangas
}

func (ls *CEDownloadList) RemoveOne(data *MangaData) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	index := ls.indexURL[data.mangaURL]
	delete(ls.indexURL, data.mangaURL)
	ls.list = append(ls.list[0:index], ls.list[index+1:]...)
	//Reindexar
	for i, m := range ls.list {
		ls.indexURL[m.mangaURL] = i
	}
}

func (ls *CEDownloadList) Remove(mangas []*MangaData) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	for _, m := range mangas {
		index := ls.indexURL[m.mangaURL]
		delete(ls.indexURL, m.mangaURL)
		ls.list = append(ls.list[0:index], ls.list[index+1:]...)
	}
	for i, m := range ls.list {
		ls.indexURL[m.mangaURL] = i
	}
}

func (ls *CEDownloadList) Index(mangaURL string) (*MangaData, bool) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	dx, ok := ls.indexURL[mangaURL]
	if ok {
		return ls.list[dx], ok
	}
	return nil, ok
}
