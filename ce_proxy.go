package main

import (
	"encoding/json"
	"sync"
)

type Proxy struct {
	url            string
	user           string
	password       string
	useSystemProxy bool
	rwProxy        *sync.RWMutex
}

func NewSystemProxy() *Proxy {
	return &Proxy{
		useSystemProxy: true,
		rwProxy:        &sync.RWMutex{},
	}
}

func NewProxy(user, password, url string) *Proxy {
	return &Proxy{
		user:           user,
		password:       password,
		url:            url,
		useSystemProxy: false,
		rwProxy:        &sync.RWMutex{},
	}
}

func (ce *Proxy) SetURL(url string) {
	ce.rwProxy.Lock()
	defer ce.rwProxy.Unlock()
	ce.url = url
}

func (ce *Proxy) GetURL() string {
	ce.rwProxy.RLock()
	defer ce.rwProxy.RUnlock()
	return ce.url
}

func (ce *Proxy) SetUser(user string) {
	ce.rwProxy.Lock()
	defer ce.rwProxy.Unlock()
	ce.user = user
}

func (ce *Proxy) GetUser() string {
	ce.rwProxy.RLock()
	defer ce.rwProxy.RUnlock()
	return ce.user
}

func (ce *Proxy) GetPassword() string {
	ce.rwProxy.RLock()
	defer ce.rwProxy.RUnlock()
	return ce.password
}

func (ce *Proxy) SetPassword(password string) {
	ce.rwProxy.Lock()
	defer ce.rwProxy.Unlock()
	ce.password = password
}

func (ce *Proxy) GetUseSystemProxy() bool {
	ce.rwProxy.RLock()
	defer ce.rwProxy.RUnlock()
	return ce.useSystemProxy
}

func (ce *Proxy) SetUseSystemProxy(useSystemProxy bool) {
	ce.rwProxy.Lock()
	defer ce.rwProxy.Unlock()
	ce.useSystemProxy = useSystemProxy
}

func (ce *Proxy) MarshalJSON() ([]byte, error) {
	ce.rwProxy.Lock()
	defer ce.rwProxy.Unlock()
	config := struct {
		User           string `json:"user"`
		Password       string `json:"password"`
		URL            string `json:"proxyurl"`
		UseSystemProxy bool   `json:"use-system-proxy"`
	}{
		User:           ce.user,
		Password:       ce.password,
		URL:            ce.url,
		UseSystemProxy: ce.useSystemProxy,
	}
	return json.Marshal(config)
}

func (ce *Proxy) UnmarshalJSON(data []byte) error {
	ce.rwProxy.Lock()
	defer ce.rwProxy.Unlock()
	config := struct {
		User           string `json:"user"`
		Password       string `json:"password"`
		URL            string `json:"proxyurl"`
		UseSystemProxy bool   `json:"use-system-proxy"`
	}{}
	err := json.Unmarshal(data, &config)
	if err != nil {
		return err
	}
	ce.user = config.User
	ce.password = config.Password
	ce.url = config.URL
	ce.useSystemProxy = config.UseSystemProxy
	return nil
}
