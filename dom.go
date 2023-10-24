package browser

import (
	"context"
	"errors"

	"github.com/gospider007/bs4"
	"github.com/gospider007/cdp"
	"github.com/gospider007/gson"
)

type Dom struct {
	baseUrl string
	webSock *cdp.WebSock
	nodeId  int64
	ele     *bs4.Client
}

func (obj *Dom) Rect(ctx context.Context) (cdp.Rect, error) {
	rs, err := obj.webSock.DOMGetBoxModel(ctx, obj.nodeId)
	if err != nil {
		return cdp.Rect{}, err
	}
	model, ok := rs.Result["model"]
	if !ok {
		return cdp.Rect{}, errors.New("not rect")
	}
	jsonData, err := gson.Decode(model)
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

func (obj *Dom) String() string {
	return obj.ele.String()
}

func (obj *Dom) Focus(ctx context.Context) error {
	_, err := obj.webSock.DOMFocus(ctx, obj.nodeId)
	return err
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
