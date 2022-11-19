package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/PuerkitoBio/goquery"

	//"github.com/chai2010/webp"
	"golang.org/x/image/webp"
)

const CACHE_MAX int = 128

type Cache struct {
	cacheMap map[string]*CacheItem
	queue    []string
	cacheURI fyne.URI
	rw       sync.RWMutex
}

func (cache *Cache) getURI(id int) fyne.URI {
	uri, err := storage.ParseURI("file://" + cache.cacheURI.Path() + fmt.Sprintf("/%d.cache", id))
	if err != nil {
		panic(err)
	}
	return uri
}

func (cache *Cache) saveCache() error {
	cache.rw.RLock()
	defer cache.rw.RUnlock()
	js, err := json.Marshal(cache.cacheMap)
	if err != nil {
		return err
	}
	fileURI, err := storage.ParseURI("file://" + cache.cacheURI.Path() + "/data.json")
	if err != nil {
		log.Panic(err)
	}
	writer, err := storage.Writer(fileURI)
	if err != nil {
		return err
	}
	defer writer.Close()
	writer.Write(js)
	for _, c := range cache.cacheMap {
		c.unLoad(cache)
	}
	return nil
}

func (cache *Cache) loadCache() error {
	cache.rw.Lock()
	defer cache.rw.Unlock()
	dataURI, err := storage.ParseURI("file://" + cache.cacheURI.Path() + "/data.json")
	if err != nil {
		log.Panic(err)
	}
	reader, err := storage.Reader(dataURI)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(reader)
	var cacheItems []*CacheItem
	err = decoder.Decode(&cacheItems)
	if err != nil {
		return err
	}
	cache.cacheMap = make(map[string]*CacheItem)
	for _, item := range cacheItems {
		cache.cacheMap[item.url.String()] = item
		//item.resourceMutex.Lock()
		//if err := item.load(cache); err != nil {
		//	log.Println(err)
		//} //else if item.resource != nil {
		//cache.cacheItems = append(cache.cacheItems, item)
		//}
		/*if err := item.load(cache); err != nil {
			log.Panic(err)
		}*/
		//item.resourceMutex.Unlock()
	}
	return nil
}

func newCache() *Cache {
	cache := new(Cache)
	cache.cacheMap = make(map[string]*CacheItem)
	//cache.cacheItems = make([]*CacheItem, 0, 1000)
	//cache.rw = sync.NewCond(&sync.Mutex{})
	var err error
	app := fyne.CurrentApp()
	root := app.Storage().RootURI()
	cache.cacheURI, err = storage.ParseURI("file://" + root.Path() + "/cache")
	if err != nil {
		panic(err)
	}
	if exists, err := storage.Exists(cache.cacheURI); err != nil {
		panic(err)
	} else if exists {
		cache.loadCache()
	} else {
		if err := storage.CreateListable(cache.cacheURI); err != nil {
			panic(err)
		}
	}
	return cache
}

func (cache *Cache) index(u string) (*CacheItem, bool) {
	internal, ok := cache.cacheMap[u]
	if !ok {
		return nil, false
	}
	if internal.isLoaded {
		//Actualizar la cola
		for i, c := 0, len(cache.queue); i < c; i++ {
			if cache.queue[i] == internal.url.String() {
				nQueue := make([]string, len(cache.queue))
				copy(nQueue, cache.queue[0:i])
				copy(nQueue[i:], cache.queue[i+1:])
				cache.queue = nQueue[:len(nQueue)-1]
				i = c
			}
		}
		cache.queue = append(cache.queue, u)
		return internal, ok
	}
	internal.resourceMutex.Lock()
	defer internal.resourceMutex.Unlock()
	if err := internal.load(cache); err != nil {
		delete(cache.cacheMap, internal.url.String())
		return nil, false
	}
	return internal, ok
}

func (cache *Cache) register(u string, item *CacheItem) {
	cache.cacheMap[u] = item
	if len(cache.queue) >= CACHE_MAX {
		cacheItem := cache.cacheMap[cache.queue[0]]
		cacheItem.unLoad(cache)
		cacheItem.resourceMutex.Lock()
		cacheItem.resource = nil
		cacheItem.resourceMutex.Unlock()
		cache.queue = cache.queue[1:]
	}
	cache.queue = append(cache.queue, u)
	//item.id = len(cache.cacheItems)
	//cache.cacheItems = append(cache.cacheItems, item)
	//cache.rw.Signal()
}

type Browser struct {
	basic    string
	client   *http.Client
	internal *Cache
	rw       *sync.RWMutex
}

func NewBrowser() *Browser {
	br := new(Browser)
	br.rw = &sync.RWMutex{}
	br.internal = newCache()
	cookies, err := cookiejar.New(nil)
	if err != nil {
		log.Panic(err)
	}

	br.client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			lastRequest := via[len(via)-1]
			if req.URL.Host != lastRequest.URL.Host {
				req.Header.Del("Authorization")
			}
			return nil
		},
		Jar:     cookies,
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          20,
			ResponseHeaderTimeout: 30 * time.Second,
			Proxy:                 http.ProxyFromEnvironment,
		},
	}
	return br
}

func (br *Browser) SetProxy(proxy *Proxy) {
	br.rw.Lock()
	defer br.rw.Unlock()
	t := br.client.Transport.(*http.Transport)
	if proxy.GetUseSystemProxy() {
		br.basic = ""
		t.Proxy = http.ProxyFromEnvironment
	} else {
		auth := fmt.Sprintf("%s:%s", proxy.GetUser(), proxy.GetPassword())
		proxyURL, err := url.Parse(proxy.GetURL())
		if err != nil {
			log.Println(err)
			return
		}
		br.basic = "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		t.Proxy = http.ProxyURL(proxyURL)
		hdr := http.Header{}
		hdr.Add("Proxy-Authorization", br.basic)
		t.ProxyConnectHeader = hdr
		t.GetProxyConnectHeader = func(ctx context.Context, proxyURL *url.URL, target string) (http.Header, error) {
			header := http.Header{}
			header.Add("Proxy-Authorization", br.basic)
			return header, nil
		}
	}
}

func (br *Browser) UnLoadCache() {
	if err := br.internal.saveCache(); err != nil {
		log.Panic(err)
	}
}

func (br *Browser) LoadCache() {
	if err := br.internal.loadCache(); err != nil {
		log.Println(err)
	}
}

func (br *Browser) cache(u string) (*CacheItem, error) {
	//internal.StartBackTracking()
	item := &CacheItem{
		mutex:         &sync.Mutex{},
		resourceMutex: &sync.RWMutex{},
		usageMutex:    &sync.RWMutex{},
		usage:         time.Now(),
	}
	item.resourceMutex.Lock()
	br.internal.rw.Lock()
	if ready, ok := br.internal.index(u); ok {
		_, isFormat := ready.err.(FormatError)
		ready.SetUsage(time.Now())
		if ready.resource != nil && (ready.err == nil || isFormat) {
			defer item.resourceMutex.Unlock()
			defer br.internal.rw.Unlock()
			return ready, nil
		}
		ready.resourceMutex.Lock()
		item.resourceMutex.Unlock()
		item = ready
		item.retryCount++
	} else {
		br.internal.register(u, item) //seccion critica para cache
	}
	br.internal.rw.Unlock()
	defer item.resourceMutex.Unlock()
	if _, err := br.fetch(u, item, false); err != nil {
		return nil, err
	} else {
		//seccion critica para item
		if item.url.Path == "/" {
			item.format = "html"
		} else {
			index := strings.LastIndex(item.url.Path, ".")
			if index == -1 {
				item.format = "html"
			} else {
				item.format = item.url.Path[index+1:]
				switch item.format {
				case "jpg":
					item.format = "jpeg"
				}
				//fmt.Printf("%s %s\n\r", item.format, string((item.resource).([]byte)[0:30]))
			}
		}
		err := item.load(br.internal)
		return item, err
	}
}

func (br *Browser) fetch(u string, item *CacheItem, isRes bool) (interface{}, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	header := map[string][]string{
		"Host":                      {parsedURL.Host},
		"Filename":                  {parsedURL.Path},
		"Connection":                {"keep-alive"},
		"Upgrade-Insecure-Requests": {"1"},
		"Sec-Fetch-Dest":            {"document"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-Site":            {"same-origin"},
		"Sec-Fetch-User":            {"?1"},
		"User-Agent":                {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:103.0) Gecko/20100101 Firefox/103.0"},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
		"Accept-Language":           {"es-ES,es;q=0.8,en-US;q=0.5,en;q=0.3"},
		"Referrer-Policy":           {"strict-origin-when-cross-origin"},
		"Pragma":                    {"no-cache"},
		"Cache-Control":             {"no-cache"},
	}
	req := &http.Request{
		Method:     "GET",
		URL:        parsedURL,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
		Body:       nil,
		Host:       parsedURL.Host,
	}
	if item != nil {
		item.url = parsedURL
		trace := &httptrace.ClientTrace{
			ConnectStart: func(network, addr string) {
				item.connect = time.Now()
			},
			ConnectDone: func(network, addr string, err error) {
				item.connectDuration = time.Since(item.connect)
			},
			GetConn: func(hostPort string) {
				item.start = time.Now()
			},
			GotFirstResponseByte: func() {
				item.firstByteDuration = time.Since(item.start)
			},
		}
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	}
	br.rw.RLock()
	if br.basic != "" {
		req.Header.Add("Proxy-Authorization", br.basic)
		httputil.DumpRequest(req, false)
	}
	response, err := br.client.Do(req)
	br.rw.RUnlock()
	c := 0
	for response == nil && c < 4 {
		response, err = br.client.Do(req)
		c++
	}
	if response != nil && response.StatusCode != 200 {
		err = fmt.Errorf("error for status code %d", response.StatusCode)
	}
	if err != nil {
		if item != nil {
			item.err = err
		}
		return nil, err
	}
	if item != nil {
		item.err = nil
		if isRes {
			item.resource = response
		} else {
			item.resource = response.Body
		}
		item.isLoaded = false
	}
	if isRes {
		return response, nil
	}
	return response.Body, nil
}

func (br *Browser) Get(u string) (interface{}, error) {
	item, err := br.cache(u)
	if err != nil {
		return nil, err
	}
	return item.Resource()
}

func (br *Browser) Download(u string) (io.ReadCloser, error) {
	reader, err := br.fetch(u, nil, false)
	return reader.(io.ReadCloser), err
}

func (br *Browser) DownloadWithResponse(u string) (*http.Response, error) {
	res, err := br.fetch(u, nil, true)
	if res != nil {
		return res.(*http.Response), err
	}
	return nil, fmt.Errorf("response is nil")
}

type FormatError struct {
	message string
}

func (e FormatError) Error() string {
	return e.message
}

type CacheItem struct {
	start, connect                     time.Time
	connectDuration, firstByteDuration time.Duration
	retryCount                         int
	err                                error
	id                                 int
	usage                              time.Time
	resource                           interface{}
	format                             string
	isLoaded                           bool
	url                                *url.URL
	mutex                              *sync.Mutex
	resourceMutex                      *sync.RWMutex
	usageMutex                         *sync.RWMutex
}

func (ci *CacheItem) SetUsage(usage time.Time) {
	ci.usageMutex.Lock()
	defer ci.usageMutex.Unlock()
	ci.usage = usage
}

func (ci *CacheItem) Usage() time.Time {
	ci.usageMutex.RLock()
	defer ci.usageMutex.RUnlock()
	return ci.usage
}

func (ci *CacheItem) Resource() (interface{}, error) {
	ci.resourceMutex.RLock()
	defer ci.resourceMutex.RUnlock()
	return ci.resource, ci.err
}

func (ci *CacheItem) IsLoaded() bool {
	ci.resourceMutex.RLock()
	defer ci.resourceMutex.RUnlock()
	return ci.resource != nil
}

func (ci *CacheItem) MarshalJSON() ([]byte, error) {
	ci.mutex.Lock()
	defer ci.mutex.Unlock()
	errStr := ""
	if ci.err != nil {
		errStr = ci.err.Error()
	}
	cacheItem := struct {
		Start             string `json:"Start"`
		Connect           string `json:"Connect"`
		FirstByteDuration int64  `json:"FirstByteDuration"`
		ConnectDuration   int64  `json:"ConnectDuration"`
		Format            string `json:"Format"`
		RetryCount        int    `json:"RetryCount"`
		ID                int    `json:"ID"`
		URL               string `json:"URL"`
		Error             string `json:"Error"`
	}{
		Start:             ci.start.Format(time.RFC3339Nano),
		Connect:           ci.connect.Format(time.RFC3339Nano),
		FirstByteDuration: int64(ci.firstByteDuration),
		ConnectDuration:   int64(ci.connectDuration),
		Format:            ci.format,
		RetryCount:        ci.retryCount,
		ID:                ci.id,
		URL:               ci.url.String(),
		Error:             errStr,
	}
	return json.Marshal(cacheItem)
}

func (ci *CacheItem) UnmarshalJSON(data []byte) error {
	item := struct {
		Start             string `json:"Start"`
		Connect           string `json:"Connect"`
		FirstByteDuration int64  `json:"FirstByteDuration"`
		ConnectDuration   int64  `json:"ConnectDuration"`
		Format            string `json:"Format"`
		RetryCount        int    `json:"RetryCount"`
		ID                int    `json:"ID"`
		URL               string `json:"URL"`
		Error             string `json:"Error"`
	}{}
	err := json.Unmarshal(data, &item)
	if err != nil {
		return err
	}
	ci.start, err = time.Parse(time.RFC3339Nano, item.Start)
	if err != nil {
		return err
	}
	ci.connect, err = time.Parse(time.RFC3339Nano, item.Connect)
	if err != nil {
		return err
	}
	ci.firstByteDuration = time.Duration(item.FirstByteDuration)
	ci.connectDuration = time.Duration(item.ConnectDuration)
	ci.format = item.Format
	ci.retryCount = item.RetryCount
	ci.id = item.ID
	parsedURL, err := url.Parse(item.URL)
	if err != nil {
		return err
	}
	ci.url = parsedURL
	if item.Error != "" {
		ci.err = fmt.Errorf(item.Error)
	}
	ci.mutex = &sync.Mutex{}
	ci.resourceMutex = &sync.RWMutex{}
	ci.usageMutex = &sync.RWMutex{}
	/*
		Start             string `json:Start`
		Connect           string `json:Connect`
		FirstByteDuration int64 `json:FirstByteDuration`
		ConnectDuration   int64 `json:ConnectDuration`
		Format            string `json:"Format"`
		RetryCount        int    `json:"RetryCount"`
		ID                int    `json:"ID"`
		URL               string `json:"URL"`
		Error             string `json:"Error"`
	*/
	/*for k, v := range raw {
		switch strings.ToLower(k) {
		case "start":
			ci.start, err = time.Parse(time.RFC3339Nano, v)
			if err != nil {
				return err
			}
		case "connect":
			ci.connect, err = time.Parse(time.RFC3339Nano, v)
			if err != nil {
				return err
			}
		case "firstbyteduration":
			var fbd int64
			fbd, err = strconv.ParseInt(v, 10, 64)
			ci.firstByteDuration = time.Duration(fbd)
		case "connectduration":
			var cd int64
			cd, err = strconv.ParseInt(v, 10, 64)
			ci.connectDuration = time.Duration(cd)
		case "format":
			switch v {
			case "jpeg", "png", "webp", "html":
				ci.format = v
			default:
				return fmt.Errorf("unsupported format '%s'", v)
			}
		case "retrycount":
			ci.retryCount, err = strconv.Atoi(v)
			if err != nil {
				return err
			}
		case "id":
			ci.id, err = strconv.Atoi(v)
			if err != nil {
				return err
			}
		case "url":
			ci.url, err = url.Parse(v)
			if err != nil {
				return err
			}
		case "error":
			ci.err = fmt.Errorf(v)
		}
	}*/
	return nil
}

func (ci *CacheItem) unLoad(cache *Cache) {
	ci.resourceMutex.Lock()
	defer ci.resourceMutex.Unlock()
	if ci.isLoaded {
		writer, err := storage.Writer(cache.getURI(ci.id))
		if err != nil {
			panic(err)
		}
		defer writer.Close()
		switch res := ci.resource.(type) {
		case *goquery.Document:
			page, err := res.Html()
			if err != nil {
				log.Panic(err)
			}
			_, err = writer.Write([]byte(page))
		case image.Image:
			switch ci.format {
			case "jpeg":
				err = jpeg.Encode(writer, res, &jpeg.Options{Quality: 100})
			case "png":
				err = png.Encode(writer, res)
			}
		}
		if err != nil {
			log.Panic(err)
		}
		ci.resource = nil
		ci.isLoaded = false
	}
}

func (ci *CacheItem) load(cache *Cache) error {
	if !ci.isLoaded {
		var reader io.ReadCloser
		var err error
		if ci.resource == nil && ci.err == nil {
			reader, err = storage.Reader(cache.getURI(ci.id))
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if ci.err != nil {
			return ci.err
		} else {
			reader = ci.resource.(io.ReadCloser)
		}
		defer reader.Close()
		switch ci.format {
		case "html":
			data, err := ioutil.ReadAll(reader)
			if err != nil {
				log.Println(err)
				ci.err = err
				return ci.err
			}
			byteReader := bytes.NewReader(data)
			ci.resource, ci.err = goquery.NewDocumentFromReader(byteReader)
		case "webp":
			ci.resource, ci.err = webp.Decode(reader)
			ci.format = "jpeg"
		case "jpeg", "png":
			ci.resource, _, ci.err = image.Decode(reader)
		default:
			ci.err = FormatError{"request format error"}
			return ci.err
		}
		ci.isLoaded = true
		/*if ci.err != nil {
			log.Println("")
		}
		if _, ok := ci.resource.(io.ReadCloser); ok {
			log.Println("ReadCloser")
		}*/
		return ci.err
	}
	return nil
}
