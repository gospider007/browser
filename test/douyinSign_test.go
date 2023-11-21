package main

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/gospider007/browser"
	"github.com/gospider007/cdp"
	"github.com/gospider007/tools"
)

var signFunc = `(params) => {
	return new Promise((resolve, reject) => {
		let xhr = new XMLHttpRequest();
		xhr.open(params.method, params.url);
		xhr.onload = () => {resolve(xhr.responseURL)};
		xhr.send(null)
	})
}
`

func douyinSign() {
	client, err := browser.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	page, err := client.NewPage(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()
	pageDone := make(chan struct{})
	var pageOk bool
	log.Print("进行请求拦截设置")
	page.Request(nil, func(ctx context.Context, r *cdp.Route) {
		if strings.Contains(r.Url(), "&msToken=") { //判断 sign 加密函数是否加密成功
			select {
			case pageDone <- struct{}{}: //通知加密函数加载完成
			default:
			}
			pageOk = true
		}
		if pageOk { //加密函数加载成功，现在开始拦截所有请求，对所有发送的请求内容返回空
			r.FulFill(ctx)
		}
	})
	log.Print("访问抖音")
	err = page.GoTo(nil, "https://www.douyin.com/")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("访问抖音")
	select {
	case <-page.Done():
		log.Panic("页面关闭了")
	case <-pageDone:
	case <-time.After(time.Second * 30):
		log.Panic("页面超时了")
	}
	log.Print("生成鼠标轨迹")
	for i := 0; i < 5; i++ {
		if err = page.Move(nil, tools.RanFloat64(100, 200), tools.RanFloat64(200, 300), 20); err != nil {
			return
		}
		if err = page.Move(nil, tools.RanFloat64(400, 500), tools.RanFloat64(800, 900), 20); err != nil {
			return
		}
		time.Sleep(time.Second)
	}
	log.Print("开始测试 sign 生成")
	t := time.Now()
	for i := 0; i < 100; i++ {
		rsTest, err := page.Eval(nil, signFunc, map[string]interface{}{
			"method": "GET",
			"url":    "/aweme/v1/web/search/item/?device_platform=webapp&aid=6383&channel=channel_pc_web&search_channel=aweme_video_web&sort_type=0&publish_time=0&keyword=孙悟空&search_source=normal_search&query_correct_type=1&is_filter_search=0&from_group_id=&offset=0&count=10&pc_client_type=1&version_code=170400&version_name=17.4.0&cookie_enabled=true&screen_width=1920&screen_height=1080&browser_language=zh-CN&browser_platform=Win32&browser_name=Edge&browser_version=113.0.1774.57&browser_online=true&engine_name=Blink&engine_version=113.0.0.0&os_name=Windows&os_version=10&cpu_core_num=6&device_memory=8&platform=PC&downlink=1.6&effective_type=4g&round_trip_time=100",
		})
		if err != nil {
			log.Panic(err)
		}
		log.Print(rsTest)
	}
	log.Print("100次耗时：", time.Since(t))
}
func TestDouyinSign(t *testing.T) {
	douyinSign()
}
