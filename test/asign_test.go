package main

import (
	"context"
	_ "embed"
	"log"
	"testing"

	"github.com/gospider007/browser"
	"github.com/gospider007/cdp"
)

//go:embed acrawler.html
var acrawer string

func douyinAsign() {
	client, err := browser.NewClient(nil, browser.ClientOption{Headless: true})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	page, err := client.NewPage(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()
	page.Request(nil, func(ctx context.Context, r *cdp.Route) {
		r.FulFill(ctx, cdp.FulData{
			Body: acrawer,
		})
	})
	if err = page.GoTo(nil, "https://so.toutiao.com/"); err != nil {
		log.Fatal(err)
	}
	err = page.WaitDomLoad(nil)
	if err != nil {
		log.Fatal(err)
	}
	rs, err := page.Eval(nil, `(params)=>{
		window.byted_acrawler.init({ aid: 99999999, dfp: 0 }); 
		return window.byted_acrawler.sign("", params.ac);
	}`, map[string]any{"ac": "111111111"})
	if err != nil {
		log.Fatal(err)
	}
	log.Print(rs)
}
func TestDouyinAsign(t *testing.T) {
	douyinAsign()
}
