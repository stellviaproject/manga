package main

import (
	_ "image/jpeg"
	"sync"
)

type CEConfig struct {
	Proxy *Proxy `json:"proxy"`
	rw    *sync.RWMutex
}

func NewCEConfig() *CEConfig {
	return &CEConfig{
		Proxy: NewSystemProxy(),
		rw:    &sync.RWMutex{},
	}
}

func (ce *CEConfig) GetProxy() *Proxy {
	ce.rw.RLock()
	defer ce.rw.RUnlock()
	return ce.Proxy
}

func (ce *CEConfig) SetProxy(proxy *Proxy) {
	ce.rw.Lock()
	defer ce.rw.Unlock()
	ce.Proxy = proxy
}
