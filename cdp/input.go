package cdp

import (
	"context"
)

type DispatchKeyEventOption struct {
	Type           string `json:"type"`
	Key            string `json:"key"`
	Text           string `json:"text"`
	UnmodifiedText string `json:"unmodifiedText"`
}

func (obj *WebSock) InputDispatchKeyEvent(ctx context.Context, option DispatchKeyEventOption) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "Input.dispatchKeyEvent",
		Params: map[string]any{
			"type":           option.Type,
			"key":            option.Key,
			"text":           option.Text,
			"unmodifiedText": option.UnmodifiedText,
		},
	})
}

type DispatchMouseEventOption struct {
	Type       string  `json:"type"`
	Button     string  `json:"button"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	ClickCount int64   `json:"clickCount"`
	DeltaX     float64 `json:"deltaX"`
	DeltaY     float64 `json:"deltaY"`
}

func (obj *WebSock) InputDispatchMouseEvent(ctx context.Context, option DispatchMouseEventOption) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "Input.dispatchMouseEvent",
		Params: map[string]any{
			"type":       option.Type,
			"button":     option.Button,
			"clickCount": option.ClickCount,
			"x":          option.X,
			"y":          option.Y,
			"deltaX":     option.DeltaX,
			"deltaY":     option.DeltaY,
		},
	})
}

func (obj *WebSock) InputDispatchTouchEvent(ctx context.Context, Type string, TouchPoints []Point) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "Input.dispatchTouchEvent",
		Params: map[string]any{
			"type":        Type,
			"touchPoints": TouchPoints,
		},
	})
}
