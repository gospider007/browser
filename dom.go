package browser

import (
	"context"
	"errors"

	"gitee.com/baixudong/cdp"
	"gitee.com/baixudong/tools"
)

type Dom struct {
	baseUrl  string
	webSock  *cdp.WebSock
	nodeId   int64
	isIframe bool
}

func (obj *Page) getFrameHtml(ctx context.Context, frameId string) (string, error) {
	rs, err := obj.webSock.DOMGetFrameOwner(ctx, frameId)
	if err != nil {
		return "", err
	}
	jsonData, err := tools.Any2json(rs.Result)
	if err != nil {
		return "", err
	}
	if !jsonData.Get("backendNodeId").Exists() {
		return "", errors.New("not fuond backendNodeId")
	}
	backendNodeId := jsonData.Get("backendNodeId").Int()
	rs, err = obj.webSock.DOMDescribeNode(ctx, 0, backendNodeId)
	if err != nil {
		return "", err
	}
	if jsonData, err = tools.Any2json(rs.Result); err != nil {
		return "", err
	}
	if !jsonData.Get("node.contentDocument.backendNodeId").Exists() {
		return "", errors.New("not fuond backendNodeId")
	}
	backendNodeId = jsonData.Get("node.contentDocument.backendNodeId").Int()
	rs, err = obj.webSock.DOMGetOuterHTML(ctx, 0, backendNodeId)
	return rs.Result["outerHTML"].(string), err
}

func (obj *Dom) frame2Dom(ctx context.Context) error {
	rs, err := obj.webSock.DOMDescribeNode(ctx, obj.nodeId, 0)
	if err != nil {
		return err
	}
	jsonData, err := tools.Any2json(rs.Result)
	if err != nil {
		return err
	}
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
	if jsonData, err = tools.Any2json(rs.Result); err != nil {
		return err
	}
	objectId := jsonData.Get("object.objectId").String()
	rs, err = obj.webSock.DOMRequestNode(ctx, objectId)
	if err != nil {
		return err
	}
	if jsonData, err = tools.Any2json(rs.Result); err != nil {
		return err
	}
	obj.nodeId = jsonData.Get("nodeId").Int()
	return err
}
func (obj *Dom) Rect(ctx context.Context) (cdp.Rect, error) {
	rs, err := obj.webSock.DOMGetBoxModel(ctx, obj.nodeId)
	if err != nil {
		return cdp.Rect{}, err
	}
	jsonData, err := tools.Any2json(rs.Result["model"])
	if err != nil {
		return cdp.Rect{}, err
	}
	content := jsonData.Get("content").Array()
	if len(content) == 0 {
		return cdp.Rect{}, errors.New("rect没有content")
	}
	boxData := cdp.Rect{
		X:      content[0].Float(),
		Y:      content[1].Float(),
		Width:  jsonData.Get("width").Float(),
		Height: jsonData.Get("height").Float(),
	}
	return boxData, nil
}
func (obj *Dom) Show(ctx context.Context) error {
	_, err := obj.webSock.DOMScrollIntoViewIfNeeded(ctx, obj.nodeId)
	return err
}
