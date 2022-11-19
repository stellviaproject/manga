package main

import (
	"html"
	"sync"
)

type Chapter struct {
	name       string
	chapterURL string
	date       string
	mutex      *sync.Mutex
}

func NewChapter(name string, chapterURL string, date string) *Chapter {
	name = html.UnescapeString(name)
	date = html.UnescapeString(date)
	return &Chapter{
		name:       name,
		chapterURL: chapterURL,
		date:       date,
		mutex:      &sync.Mutex{},
	}
}

func (c *Chapter) GetName() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.name
}

func (c *Chapter) GetDate() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.date
}

func (c *Chapter) GetURL() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.chapterURL
}
