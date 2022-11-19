package main

import (
	"log"
	"strings"
	"sync"
)

//Se va a encargar de gestionar las descargas
type CCDownload struct {
	workers      []*DownloadWorker
	workerMap    map[*MangaData]*DownloadWorker
	downloadList *CEDownloadList
	mutex        *sync.Mutex
	condition    *sync.Cond
	worker       *DownloadWorker
	workerMutex  *sync.Mutex
}

func NewCCDownload() *CCDownload {
	return &CCDownload{
		workers:      make([]*DownloadWorker, 0, 10),
		downloadList: NewCEDownloadList(),
		workerMap:    make(map[*MangaData]*DownloadWorker),
		mutex:        &sync.Mutex{},
		condition:    sync.NewCond(&sync.Mutex{}),
		workerMutex:  &sync.Mutex{},
	}
}

func (cc *CCDownload) Proccess() {
	go func() {
		for {
			cc.condition.L.Lock()
			if len(cc.workers) == 0 {
				log.Println("[Proccess] Download thread is going to sleep...")
				cc.condition.Wait()
			}
			cc.condition.L.Unlock()

			cc.condition.L.Lock()
			{ //Seccion critica para cc.worker y cc.workers
				//cc.worker esta en una seccion critica porque esta siendo modificado
				//y otro hilo podria estar usandolo como lectura
				cc.workerMutex.Lock()
				cc.worker = cc.workers[0]
				cc.workerMutex.Unlock()
			}
			cc.condition.L.Unlock()
			//No hay seccion critica porque solo este hilo puede modificarlo
			cc.worker.Resume()
			cc.worker.Proccess()

			cc.condition.L.Lock()
			delete(cc.workerMap, cc.worker.manga)
			cc.workers = cc.workers[1:] //Seccion critica para workers
			cc.condition.L.Unlock()
		}
	}()
}

func (cc *CCDownload) PauseManga(manga *MangaData) {
	cc.condition.L.Lock()
	defer cc.condition.L.Unlock()

	if w, ok := cc.workerMap[manga]; ok {
		w.Pause()
	}
}

func (cc *CCDownload) RemoveManga(manga *MangaData) {
	cc.condition.L.Lock()
	defer cc.condition.L.Unlock()

	if w, ok := cc.workerMap[manga]; ok {
		w.Pause()
		for i, c := 0, len(cc.workers); i < c; i++ {
			if cc.workers[i].manga == manga {
				delete(cc.workerMap, manga)
				cc.workers = append(cc.workers[:i], cc.workers[i+1:]...)
				break
			}
		}
	}

	cc.downloadList.RemoveOne(manga)
	controller.uiDownloadList.SetLen(cc.downloadList.Len())
}

/*func (cc *CCDownload) ReplyManga(manga *MangaData) {
	cc.workerMutex.Lock()
	defer cc.workerMutex.Unlock()
	if cc.worker.manga == manga {
		cc.worker.Reply()
	} else if _, ok := cc.workerMap[manga]; !ok {
		if manga.HasFaileds() {
			cc.rw.L.Lock()
			defer cc.rw.L.Unlock()
			w := NewDownloadWorker(manga)
			w.Reply()
			cc.workers = append(cc.workers, w)
			cc.workerMap[manga] = w
			cc.rw.Signal()
		}
	}
}*/

func (cc *CCDownload) ResumeManga(manga *MangaData) {
	cc.condition.L.Lock()
	defer cc.condition.L.Unlock()
	if w, ok := cc.workerMap[manga]; ok {
		w.Resume()
	} else {
		w = NewDownloadWorker(manga)
		cc.workers = append(cc.workers, w)
		cc.workerMap[manga] = w
	}
}

func (cc *CCDownload) Remove(mangas []*MangaData) {
	cc.condition.L.Lock()
	defer cc.condition.L.Unlock()

	for i, ci := 0, len(mangas); i < ci; i++ {
		for j := 0; j < len(cc.workers); j++ {
			if cc.workers[j].manga == mangas[i] {
				cc.workers[j].Pause()
				delete(cc.workerMap, mangas[i])
				cc.workers = append(cc.workers[:i], cc.workers[i+1:]...)
			}
		}
	}
	cc.downloadList.Remove(mangas)
	controller.uiDownloadList.SetLen(cc.downloadList.Len())
}

/*func (cc *CCDownload) Reply() {
	cc.rw.L.Lock()
	defer cc.rw.L.Unlock()
	mangas := cc.downloadList.GetMangas()
	for _, manga := range mangas {
		if _, ok := cc.workerMap[manga]; !ok {
			w := NewDownloadWorker(manga)
			cc.workers = append(cc.workers, w)
			cc.workerMap[manga] = w
		}
	}
	cc.worker.Reply()
	cc.rw.Signal()
}*/

func (cc *CCDownload) Pause() {
	cc.condition.L.Lock()
	defer cc.condition.L.Unlock()
	for _, w := range cc.workers {
		w.Pause()
	}
	cc.worker.Pause()
}

func (cc *CCDownload) Resume() {
	cc.condition.L.Lock()
	defer cc.condition.L.Unlock()
	for _, w := range cc.workers {
		w.Resume()
	}
	for _, manga := range cc.downloadList.GetMangas() {
		if _, ok := cc.workerMap[manga]; !ok {
			w := NewDownloadWorker(manga)
			cc.workers = append(cc.workers, w)
			cc.workerMap[manga] = w
		}
	}
}

func (cc *CCDownload) DoDownload(manga *Manga) {
	go func(manga *Manga) {
		log.Println("[DoDownload] manga goten from search controller...")
		log.Println("[DoDownload] looking for user chapter selection...")
		var selection []*Chapter = nil
		if view, ok := controller.view.GetView(manga); ok {
			selection = view.GetSelection()
			log.Println("[DoDownload] user chapter selection gotten...")
		}
		if selection == nil {
			selection = manga.GetChapters()
			log.Println("[DoDownload] there is not user selection...")
		}
		if len(selection) == 0 {
			log.Println("[DoDownload] manga data is not fully, loading started...")
			controller.LoadFullInfo(manga)
			log.Println("[DoDownload] manga data is loaded successfully...")
			selection = manga.GetChapters()
			log.Println("[DoDownload] all chapters are selected to dowload...")
		}
		cc.mutex.Lock()
		log.Println("[DoDownload] testing is download exists allready...")
		if data, ok := cc.downloadList.Index(manga.mangaURL); ok {
			log.Println("[DoDownload] download exists and will be updated...")
			cc.UpDateSelection(data, selection)
			log.Println("[DoDownload] download update is successfully...")
			//scontroller.uiDownloadList.downloadList.Refresh()
		} else {
			log.Println("[DoDownload] download does not exist and will be created...")
			data = cc.NewDownload(manga, selection)
			cc.downloadList.Add(data)
			log.Println("[DoDownload] download created is succesfully...")
			log.Println("[DoDownload] wake up download thread...")
			cc.condition.L.Lock()
			cc.workers = append(cc.workers, NewDownloadWorker(data))
			cc.condition.Signal()
			cc.condition.L.Unlock()
			controller.uiDownloadList.SetLen(cc.downloadList.Len())
		}
		cc.mutex.Unlock()
	}(manga)
}

func (cc *CCDownload) NewDownload(manga *Manga, selection []*Chapter) *MangaData {
	log.Println("[DoDownload.NewDownload] getting manga portrait...")
	controller.GetPortrait(manga)
	log.Println("[DoDownload.NewDownload] manga portrait gotten is successfully...")
	data := NewMangaData(manga)
	//data.Save()
	//Build ChapterData for all chapters
	log.Println("[DoDownload.NewDownload] loading manga download index...")
	chapters := make([]*ChapterData, len(selection))
	chapMutex := &sync.Mutex{}
	wait := new(sync.WaitGroup)
	wait.Add(len(selection))
	for i, ch := range selection {
		go func(i int, ch *Chapter) {
			defer wait.Done()
			chapMutex.Lock()
			chapters[i] = NewChapterData(ch)
			chapMutex.Unlock()
			log.Printf("[DoDownload.NewDownload] getting page for chapter '%s' with url[%s]...\n\r", chapters[i].name, chapters[i].chapterURL)
			chapterPage := GetPage(chapters[i].chapterURL)
			log.Println("[DoDownload.NewDownload] page gotten is successfully...")
			log.Println("[DoDownload.NewDownload] looking for manga chapter images...")
			imageList := chapterPage.Find("#page")
			//rw := sync.NewCond(&sync.Mutex{})
			//count := 0
			if imageList.Length() > 0 {
				images := imageList.Eq(0).Find("option[value]")
				for j, c := 0, images.Length(); j < c; j++ {
					imageURL, _ := images.Eq(j).Attr("value")
					chapters[i].AddImage(NewImageData(j, SITEURL+imageURL))
				}
			}
		}(i, ch)
	}
	wait.Wait()
	log.Println("[DoDownload.NewDownload] setting chapters in download list...")
	data.SetChapters(chapters)
	return data
}

func (cc *CCDownload) UpDateSelection(data *MangaData, selection []*Chapter) {
	log.Println("[DoDownload.NewDownload] download workers will be paused...")
	const ScriptBegin = "all_imgs_url: ["
	wait := new(sync.WaitGroup)
	//Build ChapterData for all chapters
	for _, ch := range selection {
		log.Println("[DoDownload.NewDownload] looking for selected chapters...")
		if _, ok := data.Index(ch.GetURL()); !ok {
			wait.Add(1)
			go func(ch *Chapter) {
				defer wait.Done()
				current := NewChapterData(ch)
				log.Printf("[DoDownload.NewDownload] getting page for chapter '%s' with url[%s]...\n\r", current.name, current.chapterURL)
				chapterPage := GetPage(current.chapterURL)
				log.Println("[DoDownload.NewDownload] page gotten is successfully...")
				log.Println("[DoDownload.NewDownload] looking for manga chapter images...")
				imageList := chapterPage.Find("#page")
				if imageList.Length() > 0 {
					lastURL, _ := imageList.Eq(0).Find("option[value]").Last().Html()
					pHTML := strings.Index(lastURL, ".html")
					URL := SITEURL + lastURL[:pHTML] + "-1.html"
					page := GetPage(URL)
					images := page.Find("div.pic_box > img")
					for j, c := 0, images.Length(); j < c; j++ {
						imageURL, _ := images.Eq(j).Attr("src")
						current.AddImage(NewImageData(j, imageURL))
					}
				}
				data.AddChapter(current)
			}(ch)
		}
	}
	wait.Wait()
}

func (cc *CCDownload) Get(index int) *MangaData {
	return cc.downloadList.Get(index)
}
