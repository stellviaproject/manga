package main

import (
	"encoding/json"
	"fmt"
	"sync"

	fyne "fyne.io/fyne/v2"
)

type ImageData struct {
	imageURL          string
	imageId           int
	isDownloaded      bool
	isDownloadedMutex *sync.Mutex
	err               error
	errMutex          *sync.Mutex
	uri               fyne.URI
	uriMutex          *sync.Mutex
}

func NewImageData(imageId int, imageURL string) *ImageData {
	return &ImageData{
		imageId:           imageId,
		imageURL:          imageURL,
		isDownloadedMutex: &sync.Mutex{},
		uriMutex:          &sync.Mutex{},
		errMutex:          &sync.Mutex{},
	}
}

func (image *ImageData) MarshalJSON() ([]byte, error) {
	i := struct {
		ImageURL string `json:"image_url"`
		ImageID int `json:"image_id"`
		IsDownloaded bool `json:"is_downloaded"`
		Err string `json:"err"`
	}{
		ImageURL: image.imageURL,
		ImageID: image.imageId,
		IsDownloaded: image.isDownloaded,
		Err: image.err.Error(),
	}
	return json.Marshal(i)
}

func (image *ImageData) UnmarshalJSON(data []byte) error {
	i := struct {
		ImageURL string `json:"image_url"`
		ImageID int `json:"image_id"`
		IsDownloaded bool `json:"is_downloaded"`
		Err string `json:"err"`
	}{}
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	image.imageURL = i.ImageURL
	image.imageId = i.ImageID
	image.isDownloaded = i.IsDownloaded
	image.err = fmt.Errorf(i.Err)
	return nil
}

func (image *ImageData) SetError(err error) {
	image.errMutex.Lock()
	defer image.errMutex.Unlock()
	image.err = err
}

func (image *ImageData) Error() error {
	image.errMutex.Lock()
	defer image.errMutex.Unlock()
	return image.err
}

func (image *ImageData) IsDownloaded() bool {
	image.isDownloadedMutex.Lock()
	defer image.isDownloadedMutex.Unlock()
	return image.isDownloaded
}

func (image *ImageData) SetIsDownloaded(downloaded bool) {
	image.isDownloadedMutex.Lock()
	defer image.isDownloadedMutex.Unlock()
	image.isDownloaded = downloaded
}

func (image *ImageData) SetURI(uri fyne.URI) {
	image.uriMutex.Lock()
	defer image.uriMutex.Unlock()
	image.uri = uri
}

func (image *ImageData) GetURI() fyne.URI {
	image.uriMutex.Lock()
	defer image.uriMutex.Unlock()
	return image.uri
}
