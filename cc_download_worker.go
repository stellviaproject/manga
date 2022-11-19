package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"unicode"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

//Se va a encargar de descargar solamente
type DownloadWorker struct {
	manga        *MangaData
	chapterQueue []*ChapterData
	imageQueue   []*ImageData
	channel      chan *ImageData
	state        WorkerState
	stateMutex   *sync.Mutex
}

func NewDownloadWorker(manga *MangaData) *DownloadWorker {
	return &DownloadWorker{
		manga:      manga,
		state:      None,
		stateMutex: &sync.Mutex{},
	}
}

//Detener todas las descargas
func (w *DownloadWorker) Stop() {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	w.state = Stop
}

func (w *DownloadWorker) Pause() {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	w.state = Pause
}

func (w *DownloadWorker) Resume() {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	w.state = Resume
}

func (w *DownloadWorker) Reply() {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	w.state = Reply
}

func (w *DownloadWorker) GetState() WorkerState {
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	return w.state
}

func (w *DownloadWorker) Proccess() {
	if w.GetState() == Resume {
		w.Prepare()
		w.MakeQueue()
		for len(w.chapterQueue) > 0 {
			current := w.chapterQueue[0]
			w.chapterQueue = w.chapterQueue[1:]
			chapterURI, err := storage.ParseURI("file://" + w.manga.GetURI().Path() + "/" + ChapterNameResolution(current.name))
			if err != nil {
				log.Panic(err)
			}
			if exists, err := storage.Exists(chapterURI); err != nil {
				log.Panic(err)
			} else if !exists {
				if err := storage.CreateListable(chapterURI); err != nil {
					log.Panic(err)
				}
			}
			current.SetURI(chapterURI)
			w.DownloadChapter(current)
			w.UpDateQueue()
		}
		w.Stop()
	}
}

func (w *DownloadWorker) UpDateQueue() {
	for _, ch := range w.manga.GetChapters() {
		if !ch.IsFull() && !w.Contains(ch) {
			w.chapterQueue = append(w.chapterQueue, ch)
		}
	}
}

func (w *DownloadWorker) Contains(chapter *ChapterData) bool {
	for _, c := range w.chapterQueue {
		if c == chapter {
			return true
		}
	}
	return false
}

func (w *DownloadWorker) Prepare() {
	folderURI := GetMangaFolder(w.manga)
	if exists, err := storage.Exists(folderURI); err != nil {
		log.Panic(err)
	} else if !exists {
		if err := storage.CreateListable(folderURI); err != nil {
			log.Panic(err)
		}
	}
	w.manga.SetURI(folderURI)
}

const DOWNLOAD_THREADS int = 1

func (w *DownloadWorker) DownloadChapter(chapter *ChapterData) {
	images := chapter.GetImages()
	if len(images) == 0 {
		chapter.SetIsFull(true)
		return
	}
	resolver := NewImageURIResolver(chapter.GetURI(), images)
	w.imageQueue = BuildImageQueue(chapter.GetImages())

	threadsCount := DOWNLOAD_THREADS
	if threadsCount > len(w.imageQueue) {
		threadsCount = len(w.imageQueue)
	}
	w.channel = make(chan *ImageData, threadsCount)
	wait := new(sync.WaitGroup)
	for i, c := 0, len(w.imageQueue); i < c; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			data := <-w.channel
			w.DownloadImage(chapter, data, resolver)
		}()
		w.channel <- w.imageQueue[i]
	}
	wait.Wait()
	/*producer := new(sync.WaitGroup)
	consummer := new(sync.WaitGroup)
	producer.Add(1)
	consummer.Add(len(w.imageQueue))
	go func() {
		defer producer.Done()
	LOOP:
		for len(w.imageQueue) > 0 {
			switch w.GetState() {
			case Pause:
				break LOOP
			}
			image := w.imageQueue[0]
			w.channel <- image
			w.imageQueue = w.imageQueue[1:]
		}
		close(w.channel)
	}()*/
	/*go func() {

	}()*/
	//
	/*rw := sync.NewCond(&sync.Mutex{})
	count := 0
	for data := range w.channel {
		go func(data *ImageData) {
			//defer consummer.Done()
			w.DownloadImage(chapter, data, resolver)
			rw.L.Lock()
			count--
			rw.Signal()
			rw.L.Unlock()
		}(data)
		count++
		rw.L.Lock()
		for count >= threadsCount {
			rw.Wait()
		}
		rw.L.Unlock()
	}
	//
	consummer.Wait()
	producer.Wait()*/
	if w.GetState() != Pause {
		for _, img := range chapter.GetImages() {
			if !img.IsDownloaded() {
				chapter.SetIsFull(false)
				return
			}
		}
		chapter.SetIsFull(true)
	}
}

func (w *DownloadWorker) DownloadImage(chapter *ChapterData, data *ImageData, resolver *ImageURIResolver) {
	/*data.SetIsDownloaded(true)
	chapter.CountOne()
	w.manga.CountOne()
	controller.uiDownloadList.item.Content.Refresh()
	time.Sleep(333 * time.Millisecond)*/
	fileURI := resolver.Resolve(data.imageId, ".webp")
	if exists, err := storage.Exists(fileURI); err == nil {
		if !data.IsDownloaded() {
			if exists {
				if st, err := os.Stat(fileURI.Path()); err != nil {
					log.Panic(err)
				} else if st.Size() > 0 {
					w.manga.CountOne()
					chapter.CountOne()
					data.SetIsDownloaded(true)
					return
				} else {
					log.Printf("[DownloadImage] uri='%s' size='%d'\n\r", fileURI.Path(), st.Size())
				}
			}
		}
		log.Printf("GET PAGE '%s'", data.imageURL)
		page := ForcePage(data.imageURL)
		imageURL, _ := page.Find("meta[property=\"og:image\"]").Attr("content")
		var reader io.ReadCloser
		var err error
		log.Printf("Download name %d '%s'\n\r", data.imageId, data.imageURL)
		reader, err = controller.browser.Download(imageURL)
		if err == nil {
			defer reader.Close()
			data.SetURI(fileURI)
			var writer io.WriteCloser
			// path := fileURI.Path()
			//webpIndex := strings.LastIndex(path, ".webp")
			// tempURI, err := storage.ParseURI("file://" + path[:webpIndex] + ".tmp")
			// if err != nil {
			// 	log.Panicln(err)
			// }
			writer, err = storage.Writer(fileURI)
			defer func() {
				if err := writer.Close(); err != nil {
					log.Println(err)
				}
			}()
			if err == nil {
				dataBytes, err := ioutil.ReadAll(reader)
				/*buffer := make([]byte, 2*1024*1024)
				var n int
				for {
					n, err = reader.Read(buffer)
					if err != nil && err != io.EOF {
						break
					}
					if n == 0 {
						err = nil
						break
					}
					if _, err = writer.Write(buffer[0:n]); err != nil {
						break
					}
				}*/
				//writer.Close()
				if err == nil {
					log.Printf("Download successfully %d '%s'\n\r", data.imageId, data.imageURL)
					_, err = writer.Write(dataBytes)
					if err == nil {
						data.SetIsDownloaded(true)
						chapter.CountOne()
						w.manga.CountOne()
						controller.uiDownloadList.Refresh()
						//os.Rename(tempURI.Path(), fileURI.Path())
						return
					}
				}
			}
			//storage.Delete(tempURI)
		}
		data.SetError(err)
		log.Println(err)
	} else {
		log.Panic(err)
	}
}

func (w *DownloadWorker) MakeQueue() {
	chapters := w.manga.GetChapters()
	for _, ch := range chapters {
		if !ch.IsFull() {
			w.chapterQueue = append(w.chapterQueue, ch)
		}
	}
}

type WorkerState int

const (
	None   WorkerState = iota
	Resume WorkerState = 1
	Pause  WorkerState = 2
	Reply  WorkerState = 3
	Stop   WorkerState = 4
)

func BuildImageQueue(imageQueue []*ImageData) []*ImageData {
	queue := make([]*ImageData, 0, len(imageQueue))
	for _, image := range imageQueue {
		if !image.IsDownloaded() {
			queue = append(queue, image)
		}
	}
	return queue
}

//var m sync.Mutex

func ChapterNameResolution(name string) string {
	//m.Lock()
	//defer m.Unlock()
	//resolved := ""
	name = strings.Trim(name, " ")
	words := strings.Split(name, " ")
	wordsOk := make([]string, 0, len(words))
	compile := func(word []rune) string {
		res := ""
		points := 0
		for i := 0; i < len(word); i++ {
			if unicode.IsLetter(word[i]) || unicode.IsDigit(word[i]) || word[i] == '.' || word[i] == '_' {
				if word[i] == '.' {
					points++
					if points < 2 {
						res += string(word[i])
					}
				} else {
					res += string(word[i])
				}
			}
		}
		return res
	}
	for i := 0; i < len(words); i++ {
		if w := compile([]rune(words[i])); w != "" {
			wordsOk = append(wordsOk, w)
		}
	}
	resolved := ""
	for i := 0; i < len(wordsOk); i++ {
		resolved += wordsOk[i] + " "
	}
	resolved = strings.Trim(resolved, " ")
	return resolved
}

func NameResolution(name string) string {
	resolved := ""
	for i, c := 0, len(name); i < c; i++ {
		if IsLetter(name[i]) || IsDigit(name[i]) {
			resolved += string(name[i])
		} else {
			resolved += " "
		}
	}
	resolved = strings.Trim(resolved, " ")
	splited := strings.Split(resolved, " ")
	for i := 0; i < len(splited); i++ {
		if len(splited[i]) > 2 {
			splited[i] = string(unicode.ToUpper(rune(splited[i][0]))) + splited[i][1:]
		}
	}
	return strings.Join(splited, " ")
}

func MaxImageID(images []*ImageData) int {
	maximun := images[0].imageId
	for _, image := range images {
		if maximun < image.imageId {
			maximun = image.imageId
		}
	}
	return maximun
}

type ImageURIResolver struct {
	length, max int
	chapterURI  fyne.URI
}

func NewImageURIResolver(chapterURI fyne.URI, images []*ImageData) *ImageURIResolver {
	res := new(ImageURIResolver)
	res.max = MaxImageID(images)
	res.length = len(fmt.Sprintf("%d", res.max))
	res.chapterURI = chapterURI
	return res
}

func (res *ImageURIResolver) Resolve(id int, format string) fyne.URI {
	if id > res.max {
		log.Panic(fmt.Errorf("[ImageURIResolver.Resolve] expected id to less than max id"))
	}
	resolved := fmt.Sprintf("%d", id)
	length := len(resolved)
	if length < res.length {
		resolved = strings.Repeat("0", res.length-length) + resolved
	}
	uri, err := storage.ParseURI("file://" + res.chapterURI.Path() + "/" + resolved + format)
	if err != nil {
		log.Panic(err)
	}
	return uri
}

func IsDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func IsLetter(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}

func GetMangaFolder(manga *MangaData) fyne.URI {
	folderName := NameResolution(manga.name)
	folderURI, err := storage.ParseURI("file://" + controller.downloadURI.Path() + "/" + folderName)
	if err != nil {
		log.Panic(err)
	}
	return folderURI
}
