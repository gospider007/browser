package browser

import (
	"context"
	"errors"
	"strings"
	"time"

	"gitee.com/baixudong/gospider/bs4"
	"gitee.com/baixudong/gospider/cdp"
	"gitee.com/baixudong/gospider/re"
	"gitee.com/baixudong/gospider/tools"
)

type Dom struct {
	baseUrl  string
	webSock  *cdp.WebSock
	nodeId   int64
	isIframe bool
}

func (obj *Dom) dom2Iframe(ctx context.Context) error {
	rs, err := obj.webSock.DOMDescribeNode(ctx, obj.nodeId)
	if err != nil {
		return err
	}
	jsonData := tools.Any2json(rs.Result)
	var backendNodeId int64
	if jsonData.Get("node.contentDocument.backendNodeId").Exists() {
		backendNodeId = jsonData.Get("node.contentDocument.backendNodeId").Int()
	} else {
		obj.isIframe = true
		return nil
	}
	rs, err = obj.webSock.DOMResolveNode(ctx, backendNodeId)
	if err != nil {
		return err
	}
	objectId := tools.Any2json(rs.Result).Get("object.objectId").String()
	rs, err = obj.webSock.DOMRequestNode(ctx, objectId)
	if err != nil {
		return err
	}
	obj.nodeId = tools.Any2json(rs.Result).Get("nodeId").Int()
	return err
}
func (obj *Dom) Html(ctx context.Context, contents ...string) (*bs4.Client, error) {
	if len(contents) > 0 {
		return nil, obj.setHtml(ctx, contents[0])
	}
	return obj.html(ctx)
}
func (obj *Dom) setHtml(ctx context.Context, content string) error {
	_, err := obj.webSock.DOMSetOuterHTML(ctx, obj.nodeId, content)
	return err
}
func (obj *Dom) html(ctx context.Context) (*bs4.Client, error) {
	rs, err := obj.webSock.DOMGetOuterHTML(ctx, obj.nodeId)
	if err != nil {
		return nil, err
	}
	html := bs4.NewClient(rs.Result["outerHTML"].(string), obj.baseUrl)
	if obj.isIframe {
		return html, nil
	}
	iframes := []*bs4.Client{}
	for _, iframe := range html.Finds("iframe") {
		if !strings.HasPrefix(iframe.Get("src"), "javascript:") {
			iframes = append(iframes, iframe)
		}
	}
	if len(iframes) > 0 {
		pageFrams, err := obj.QuerySelectorAll(ctx, "iframe")
		if err != nil {
			return nil, err
		}
		if len(iframes) != len(pageFrams) {
			return nil, errors.New("iframe error")
		}
		for i, ifram := range iframes {
			dh, err := pageFrams[i].Html(ctx)
			if err != nil {
				return nil, err
			}
			ifram.Html(dh.Html())
		}
	}
	return html, nil
}

func (obj *Dom) Focus(ctx context.Context) error {
	_, err := obj.webSock.DOMFocus(ctx, obj.nodeId)
	return err
}

func (obj *Dom) SendText(ctx context.Context, text string) error {
	err := obj.Focus(ctx)
	if err != nil {
		return err
	}
	for _, chr := range text {
		err = obj.sendChar(ctx, chr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obj *Dom) sendChar(ctx context.Context, chr rune) error {
	_, err := obj.webSock.InputDispatchKeyEvent(ctx, cdp.DispatchKeyEventOption{
		Type: "keyDown",
		Key:  "Unidentified",
	})
	if err != nil {
		return err
	}
	_, err = obj.webSock.InputDispatchKeyEvent(ctx, cdp.DispatchKeyEventOption{
		Type:           "keyDown",
		Key:            "Unidentified",
		Text:           string(chr),
		UnmodifiedText: string(chr),
	})
	if err != nil {
		return err
	}
	_, err = obj.webSock.InputDispatchKeyEvent(ctx, cdp.DispatchKeyEventOption{
		Type: "keyUp",
		Key:  "Unidentified",
	})
	return err
}
func (obj *Dom) QuerySelector(ctx context.Context, selector string) (*Dom, error) {
	dom, err := obj.querySelector(ctx, selector)
	if err != nil {
		return dom, err
	}
	if dom == nil && selector != "iframe" {
		iframes, err := obj.querySelectorAll(ctx, "iframe")
		if err != nil {
			return nil, err
		}
		for _, iframe := range iframes {
			dom, err = iframe.querySelector(ctx, selector)
			if err != nil || dom != nil {
				return dom, err
			}
		}
	}
	return dom, err
}
func (obj *Dom) querySelector(ctx context.Context, selector string) (*Dom, error) {
	rs, err := obj.webSock.DOMQuerySelector(ctx, obj.nodeId, selector)
	if err != nil {
		return nil, err
	}
	nodeId := int64(rs.Result["nodeId"].(float64))
	if nodeId == 0 {
		return nil, nil
	}
	dom := &Dom{
		baseUrl: obj.baseUrl,
		webSock: obj.webSock,
		nodeId:  nodeId,
	}
	if re.Search(`^iframe\W|\Wiframe\W|\Wiframe$|^iframe$`, selector) != nil {
		if err = dom.dom2Iframe(ctx); err != nil {
			return nil, err
		}
	}
	return dom, nil
}
func (obj *Dom) QuerySelectorAll(ctx context.Context, selector string) ([]*Dom, error) {
	dom, err := obj.querySelectorAll(ctx, selector)
	if err != nil {
		return dom, err
	}
	if dom == nil && selector != "iframe" {
		iframes, err := obj.querySelectorAll(ctx, "iframe")
		if err != nil {
			return nil, err
		}
		doms := []*Dom{}
		for _, iframe := range iframes {
			dom, err = iframe.querySelectorAll(ctx, selector)
			if err != nil {
				return dom, err
			}
			doms = append(doms, dom...)
		}
		return doms, err
	}
	return dom, err
}
func (obj *Dom) querySelectorAll(ctx context.Context, selector string) ([]*Dom, error) {
	rs, err := obj.webSock.DOMQuerySelectorAll(ctx, obj.nodeId, selector)
	if err != nil {
		return nil, err
	}
	doms := []*Dom{}
	for _, nodeId := range tools.Any2json(rs.Result["nodeIds"]).Array() {
		dom := &Dom{
			baseUrl: obj.baseUrl,
			webSock: obj.webSock,
			nodeId:  nodeId.Int(),
		}
		if re.Search(`^iframe\W|\Wiframe\W|\Wiframe$|^iframe$`, selector) != nil {
			if err = dom.dom2Iframe(ctx); err != nil {
				return nil, err
			}
		}
		doms = append(doms, dom)
	}
	return doms, nil
}
func (obj *Dom) WaitSelector(preCtx context.Context, selector string, timeouts ...int64) (*Dom, error) {
	var timeout int64 = 30
	if len(timeouts) > 0 {
		timeout = timeouts[0]
	}
	startTime := time.Now().Unix()
	for time.Now().Unix()-startTime <= timeout {
		dom, err := obj.QuerySelector(preCtx, selector)
		if err != nil {
			return nil, err
		}
		if dom != nil {
			return dom, nil
		}
		time.Sleep(time.Millisecond * 500)
	}
	return nil, errors.New("超时")
}

func (obj *Dom) Rect(ctx context.Context) (cdp.Rect, error) {
	rs, err := obj.webSock.DOMGetBoxModel(ctx, obj.nodeId)
	if err != nil {
		return cdp.Rect{}, err
	}
	jsonData := tools.Any2json(rs.Result["model"])
	content := jsonData.Get("content").Array()
	boxData := cdp.Rect{
		X:      content[0].Float(),
		Y:      content[1].Float(),
		Width:  jsonData.Get("width").Float(),
		Height: jsonData.Get("height").Float(),
	}
	return boxData, nil
}

func (obj *Dom) Img(ctx context.Context, options ...cdp.ScreenshotOption) ([]byte, error) {
	rect, err := obj.Rect(ctx)
	if err != nil {
		return nil, err
	}
	rs, err := obj.webSock.PageCaptureScreenshot(ctx, rect, options...)
	if err != nil {
		return nil, err
	}
	imgData, ok := rs.Result["data"].(string)
	if !ok {
		return nil, errors.New("not img data")
	}
	return tools.Base64Decode(imgData)
}
func (obj *Dom) Show(ctx context.Context) error {
	_, err := obj.webSock.DOMScrollIntoViewIfNeeded(ctx, obj.nodeId)
	return err
}
