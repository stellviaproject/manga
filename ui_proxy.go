package main

import (
	"log"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type UIProxy struct {
	co                  *fyne.Container
	user, password, url string
	useSystemProxy      bool
}

func NewUIProxy() *UIProxy {
	proxy := controller.config.GetProxy()
	return &UIProxy{
		user:     proxy.user,
		password: proxy.password,
		url:      proxy.url,
	}
}

func (ui *UIProxy) MakeUI() fyne.CanvasObject {
	proxy := controller.config.GetProxy()
	ui.user = proxy.GetUser()
	ui.password = proxy.GetPassword()
	ui.url = proxy.GetURL()
	userLabel, userEntry := widget.NewLabel("User: "), widget.NewEntry()
	passwordLabel, passwordEntry := widget.NewLabel("Password: "), widget.NewEntry()
	passwordEntry.Password = true
	urlLabel, urlEntry := widget.NewLabel("Proxy-URL: "), widget.NewEntry()
	labels := container.New(layout.NewVBoxLayout(), userLabel, passwordLabel, urlLabel)
	entries := []fyne.CanvasObject{userEntry, passwordEntry, urlEntry}
	ents := container.New(layout.NewVBoxLayout(), entries...)
	//Set data
	urlEntry.Text = ui.url
	userEntry.Text = ui.user
	passwordEntry.Text = ui.password
	urlEntry.Refresh()
	userEntry.Refresh()
	passwordEntry.Refresh()
	//user := container.New(layout.NewBorderLayout(nil, nil, userLabel, nil), userLabel, userText)
	//password := container.New(layout.NewBorderLayout(nil, nil, passwordLabel, nil), passwordLabel, passwordText)
	//host := container.New(layout.NewBorderLayout(nil, nil, hostLabel, nil), hostLabel, hostEntry)
	//port := container.New(layout.NewBorderLayout(nil, nil, portLabel, nil), portLabel, portEntry)
	//useSystemProxyCheck := widget.NewCheck("Use System Proxy", func(value bool) { ui.useSystemProxy = value })
	ui.co = container.NewBorder(widget.NewLabel("Proxy"), nil, labels, nil, ents)
	userEntry.OnChanged = func(s string) {
		ui.user = s
	}
	passwordEntry.OnChanged = func(s string) {
		ui.password = s
	}
	urlEntry.OnChanged = func(s string) {
		if _, err := url.Parse(s); err != nil {
			log.Println(err)
		} else {
			ui.url = urlEntry.Text
		}
	}
	return ui.co
}
