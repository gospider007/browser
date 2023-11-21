package main

import (
	"log"
	"testing"

	"github.com/gospider007/browser"
)

func TestProxy(t *testing.T) {
	client, err := browser.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	page, err := client.NewPage(nil, browser.PageOption{Proxy: "http://127.0.0.1:7007"})
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()

	err = page.GoTo(nil, "https://myip.top")
	if err != nil {
		log.Fatal(err)
	}
	err = page.WaitNetwork(nil)
	if err != nil {
		log.Fatal(err)
	}
	html, err := page.Html(nil)
	if err != nil {
		log.Panic(err)
	}
	log.Print(html)
}
