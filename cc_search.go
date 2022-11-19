package main

import (
	"log"
	"strings"
	"sync"
)

type CCSearch struct {
	searchList                            *CESearchList
	nextURL, prevURL                      string //locked by pageMutex when loading
	isFirstURL, isLastURL                 bool   //locked by pageMutex when loading
	pageMutex, imageMutex, isLoadingMutex *sync.Mutex
	isLoading                             bool
}

func NewCCSearch() *CCSearch {
	return &CCSearch{
		searchList:     NewCESearchList(),
		pageMutex:      &sync.Mutex{},
		imageMutex:     &sync.Mutex{},
		isLoadingMutex: &sync.Mutex{},
	}
}

func (cc *CCSearch) Next() {
	if !cc.IsLast() {
		cc.isLoadingMutex.Lock()
		defer cc.isLoadingMutex.Unlock()
		if !cc.isLoading {
			go cc.ProccessPage(cc.nextURL)
			cc.isLoading = true
		}
	}
}

func (cc *CCSearch) Prev() {

}

func (cc *CCSearch) Search(text string) {
	text = strings.TrimSpace(text)
	if len(text) > 0 {
		params := strings.Split(text, " ")
		base := params[0]
		for i, c := 1, len(params); i < c; i++ {
			base += "+" + params[i]
		}
		cc.searchList.Clear()
		controller.uiSearchList.SetLen(0)
		go cc.ProccessPage(SEARCH_BASE + base)
	}
}

func (cc *CCSearch) SetList(page string) {
	cc.searchList.Clear()
	controller.uiSearchList.SetLen(0)
	go cc.ProccessPage(page)
}

func (cc *CCSearch) ProccessPage(u string) {
	log.Println("[ProccessPage] start loading page...")
	cc.pageMutex.Lock()
	page := GetPage(u)
	cc.pageMutex.Unlock()
	log.Println("[ProccessPage] page loaded successfully...")
	log.Println("[ProccessPage] start page proccessing...")
	list := page.Find("ul.direlist > li")
	for i, c := 0, list.Length(); i < c; i++ {
		item := list.Eq(i)
		url, _ := item.Find("dt > a[href]").Attr("href")
		src, _ := item.Find("a > img").Attr("src")
		bookname, _ := item.Find("a.bookname").Html()
		sinopsis, _ := item.Find("dd > p").Html()
		chaptername, _ := item.Find("a.chaptername").Html()
		current := NewManga(src, url, bookname, sinopsis, chaptername)
		cc.searchList.Register(current)
	}
	list = page.Find("ul.pagelist > li")
	if list.Length() > 0 {
		curr, _ := list.Find("a.selected").Attr("href")
		controls := list.Find("a.l")
		next, _ := controls.First().Attr("href")
		prev, _ := controls.Last().Attr("href")
		cc.pageMutex.Lock()
		cc.isFirstURL = curr == prev
		cc.isLastURL = curr == next
		cc.nextURL = next
		cc.prevURL = prev
		cc.pageMutex.Unlock()
	} else {
		cc.pageMutex.Lock()
		cc.isFirstURL = true
		cc.isLastURL = true
		cc.nextURL = u
		cc.prevURL = u
		cc.pageMutex.Unlock()
	}
	controller.uiSearchList.SetLen(cc.searchList.Len())
	log.Println("[ProccessPage] page proccessing successfully...")
	cc.isLoadingMutex.Lock()
	cc.isLoading = false
	cc.isLoadingMutex.Unlock()
}

func (cc *CCSearch) IsFirst() bool {
	cc.pageMutex.Lock()
	fi := cc.isFirstURL
	cc.pageMutex.Unlock()
	return fi
}

func (cc *CCSearch) IsLast() bool {
	cc.pageMutex.Lock()
	la := cc.isLastURL
	cc.pageMutex.Unlock()
	return la
}

func (cc *CCSearch) Get(index int) *Manga {
	return cc.searchList.Get(index)
}
