<p align="center">
  <a href="https://github.com/gospider007/browser"><img src="https://go.dev/images/favicon-gopher.png"></a>
</p>
<p align="center"><strong>Browser</strong> <em>- A next-generation Browser client for Golang.</em></p>
<p align="center">
<a href="https://github.com/gospider007/browser">
    <img src="https://img.shields.io/github/last-commit/gospider007/requests">
</a>
<a href="https://github.com/gospider007/browser">
    <img src="https://img.shields.io/badge/build-passing-brightgreen">
</a>
<a href="https://github.com/gospider007/browser">
    <img src="https://img.shields.io/badge/language-golang-brightgreen">
</a>
</p>

Browser is a fully featured Browser client library for Golang. Manipulate browser can be completed with just a few lines of code
---
## Features
  * 一次编译到处运行,支持win,mac,linux
  * 无头环境与有头环境没有任何差别,只要有头测试成功,无头不会有任何问题
  * 基于强大的tls 指纹请求库 Requests，弥补浏览器模拟其他跨平台环境时在tls 指纹方面的特征,从而被反爬
  * 没有任何人机操控的痕迹,目前没有发现被检测到人机操控的网站
  * ifrmae 跨域反爬，无限嵌套接管所有iframe，注入指纹
  * 支持指纹模拟,伪造任意的指纹,用以对抗反指纹网站
  * 标签页并发安全,单浏览器，单浏览器并发上百标签页互不干扰，用以大规模高速渲染
  * 数据缓存，对静态资源全部保存本地，提升10倍大规模渲染速度
  * 支持单标签页代理，每一个标签页都是独立的代理
  * 请求的绝对掌控，控制每一个请求,完全使用golang 代替浏览器发送请求
  * 环境固化，页面的所有环境全部保存至本地保存,无网络情况下，加载rpc所需的所有环境
## Supported Go Versions
Recommended to use `go1.21.3` and above.
Initially Browser started supporting `go modules`

## CDP 调试
  * 浏览器打开 chrome://inspect/#devices
  * 点击 Open dedicated DevTools for Node
  * 添加连接，例如: 127.0.0.1:9200 ， 完成后返回： chrome://inspect/#devices 这个页面，就可以调试了
## Installation

```bash
go get github.com/gospider007/browser
```
## Usage
```go
import "github.com/gospider007/browser"
```
### Quickly send requests
```go
package main

import (
	"log"

	"github.com/gospider007/browser"
)

func main() {
	browCli, err := browser.NewClient(nil)
	if err != nil {
		log.Panic(err)
	}
	defer browCli.Close()
	page, err := browCli.NewPage(nil)
	if err != nil {
		log.Panic(err)
	}
	defer page.Close()
	log.Print("开始加载页面")
	err = page.GoTo(nil, "https://www.baidu.com")
	if err != nil {
		log.Panic(err)
	}
	log.Print("等待没有网络请求")
	err = page.WaitNetwork(nil)
	if err != nil {
		log.Panic(err)
	}
	log.Print("等待页面关闭")
	<-page.Done()
	page.Close()
}
```

# Contributing
If you have a bug report or feature request, you can [open an issue](../../issues/new)
# Contact
If you have questions, feel free to reach out to us in the following ways:
* QQ Group (Chinese): 939111384 - <a href="http://qm.qq.com/cgi-bin/qm/qr?_wv=1027&k=yI72QqgPExDqX6u_uEbzAE_XfMW6h_d3&jump_from=webapi"><img src="https://pub.idqqimg.com/wpa/images/group.png"></a>
* WeChat (Chinese): gospider007

## Sponsors
If you like and it really helps you, feel free to reward me with a cup of coffee, and don't forget to mention your github id.
<table>
    <tr>
        <td align="center">
            <img src="https://github.com/gospider007/tools/blob/master/play/wx.jpg?raw=true" height="200px" width="200px"   alt=""/>
            <br />
            <sub><b>Wechat</b></sub>
        </td>
        <td align="center">
            <img src="https://github.com/gospider007/tools/blob/master/play/qq.jpg?raw=true" height="200px" width="200px"   alt=""/>
            <br />
            <sub><b>Alipay</b></sub>
        </td>
    </tr>
</table>