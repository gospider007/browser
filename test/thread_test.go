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
	err = page.GoTo(ctx, "https://www.baidu.com")
	if err != nil {
		log.Panic(err)
	}
	err = page.WaitNetwork(ctx)
	if err != nil {
		log.Panic(err)
	}
}
func TestThread(t *testing.T) {
	client, err := browser.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	thCli := thread.NewClient(nil, 100)
	for i := 0; i < 100; i++ {
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
