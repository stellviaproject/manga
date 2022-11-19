package main

import (
	"encoding/json"
	"log"

	fyne "fyne.io/fyne/v2"
)

type CCConfig struct {
	config *CEConfig
}

func NewCCConfig() *CCConfig {
	return &CCConfig{
		config: NewCEConfig(),
	}
}

func (cc *CCConfig) Load() {
	app := fyne.CurrentApp()
	reader, err := app.Storage().Open("config.json")
	if err != nil {
		log.Println(err)
		return
	}
	defer reader.Close()
	dec := json.NewDecoder(reader)
	if err := dec.Decode(cc.config); err != nil {
		log.Println(err)
	}
}

func (cc *CCConfig) Save() {
	app := fyne.CurrentApp()
	writer, err := app.Storage().Create("config.json")
	if err != nil {
		log.Println(err)
		return
	}
	defer writer.Close()
	enc := json.NewEncoder(writer)
	if err := enc.Encode(cc.config); err != nil {
		log.Println(err)
	}
}

func (c *CCConfig) GetProxy() Proxy {
	proxy := c.config.GetProxy()
	return *proxy
}

func (c *CCConfig) SetProxy(user, password, url string, isSystemProxy bool) error {
	if isSystemProxy {
		proxy := NewSystemProxy()
		c.config.SetProxy(proxy)
		controller.browser.SetProxy(proxy)
	} else {
		proxy := NewProxy(user, password, url)
		c.config.SetProxy(proxy)
		controller.browser.SetProxy(proxy)
	}
	return nil
}
