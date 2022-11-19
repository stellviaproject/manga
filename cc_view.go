package main

import "sync"

const VIEW_MAX_COUNT int = 6

type CCView struct {
	queue   []*UIView
	views   map[*Manga]*UIView
	reverse map[*UIView]*Manga
	mutex   *sync.Mutex
}

func NewCCView() *CCView {
	return &CCView{
		queue:   make([]*UIView, 0),
		views:   make(map[*Manga]*UIView),
		reverse: make(map[*UIView]*Manga),
		mutex:   new(sync.Mutex),
	}
}

func (cc *CCView) Show(manga *Manga) {
	if ui, ok := cc.views[manga]; ok {
		ui.Show()
	} else {
		go func() {
			cc.mutex.Lock()
			defer cc.mutex.Unlock()
			ui = NewUIView(manga)
			cc.queue = append(cc.queue, ui)
			cc.reverse[ui] = manga
			cc.views[manga] = ui
			if len(cc.queue) == VIEW_MAX_COUNT {
				first := cc.queue[0]
				controller.uiAppTabs.Close(first)
				manga := cc.reverse[first]
				delete(cc.views, manga)
				delete(cc.reverse, first)
				cc.queue = cc.queue[1:]
			}
			controller.uiAppTabs.Add(manga.ShortTitle(), ui, true)
			if !manga.IsFullInfo() {
				controller.LoadFullInfo(manga)
			}
			ui.UpDate(manga)
		}()
	}
}

func (cc *CCView) GetSelection(manga *Manga) []*Chapter {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	return cc.views[manga].GetSelection()
}

func (cc *CCView) GetView(manga *Manga) (*UIView, bool) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	view, ok := cc.views[manga]
	return view, ok
}

func (cc *CCView) GetManga(view *UIView) *Manga {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	return cc.reverse[view]
}

func (cc *CCView) HasView(view *UIView) bool {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	_, ok := cc.reverse[view]
	return ok
}

func (cc *CCView) Remove(view *UIView) {
	manga := cc.reverse[view]
	delete(cc.views, manga)
	delete(cc.reverse, view)
	for i, v := range cc.queue {
		if v == view {
			cc.queue = append(cc.queue[0:i], cc.queue[i+1:]...)
			break
		}
	}
}
