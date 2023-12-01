package main

import (
	"context"
	"log"
	"testing"

	"github.com/gospider007/browser"
	"github.com/gospider007/thread"
)

func test(ctx context.Context, browCli *browser.Client, num int) {
	page, err := browCli.NewPage(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()
	log.Print(num, " goto")
	// err = page.GoTo(ctx, "https://www.baidu.com")
	err = page.GoTo(ctx, "http://www.baidu.com/")
	if err != nil {
		log.Panic(err)
	}
	log.Print(num, " 开始等待")
	err = page.WaitPageStop(ctx)
	if err != nil {
		log.Panic(err)
	}
	log.Print(num, " 开始等待 ok")
}
func TestThread(t *testing.T) {
	client, err := browser.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	total := 100
	thCli := thread.NewClient(nil, total)
	for i := 0; i < int(total); i++ {
		_, err = thCli.Write(&thread.Task{
			Func: test,
			Args: []any{client, i},
		})
		if err != nil {
			log.Panic(err)
		}
	}
	err = thCli.JoinClose()
	if err != nil {
		log.Panic(err)
	}
}
