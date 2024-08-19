package browser

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/gospider007/bs4"
	"github.com/gospider007/cdp"
	"github.com/gospider007/gson"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type NodeType int64

var (
	NodeTypeElement               NodeType = 1
	NodeTypeAttribute             NodeType = 2
	NodeTypeText                  NodeType = 3
	NodeTypeCDATA                 NodeType = 4
	NodeTypeEntityReference       NodeType = 5
	NodeTypeEntity                NodeType = 6
	NodeTypeProcessingInstruction NodeType = 7
	NodeTypeComment               NodeType = 8
	NodeTypeDocument              NodeType = 9
	NodeTypeDocumentType          NodeType = 10
	NodeTypeDocumentFragment      NodeType = 11
	NodeTypeNotation              NodeType = 12
)

func (obj NodeType) HtmlNodeType() html.NodeType {
	switch obj {
	case NodeTypeElement:
		return html.ElementNode
	case NodeTypeText:
		return html.TextNode
	case NodeTypeComment:
		return html.CommentNode
	case NodeTypeDocument:
		return html.DocumentNode
	case NodeTypeDocumentType:
		return html.DoctypeNode
	default:
		return html.RawNode
	}
}
func (obj *Page) parseJsonDom(ctx context.Context, data *gson.Client) (*html.Node, error) {
	attrs := []html.Attribute{}
	attributes := data.Get("attributes").Array()
	for i := 0; i < len(attributes)/2; i++ {
		attrs = append(attrs, html.Attribute{
			Key: attributes[i*2].String(),
			Val: attributes[i*2+1].String(),
		})
	}
	backendNodeId := data.Get("backendNodeId").Int()
	nodeId := data.Get("nodeId").Int()
	parentId := data.Get("parentId").Int()
	frameId := data.Get("frameId").String()
	if nodeId != 0 {
		attrs = append(attrs, html.Attribute{Key: "gospiderNodeId", Val: strconv.Itoa(int(nodeId))})
	}
	if backendNodeId != 0 {
		attrs = append(attrs, html.Attribute{Key: "gospiderBackendNodeId", Val: strconv.Itoa(int(backendNodeId))})
	}
	if parentId != 0 {
		attrs = append(attrs, html.Attribute{Key: "gospiderParentId", Val: strconv.Itoa(int(parentId))})
	}
	if frameId != "" {
		attrs = append(attrs, html.Attribute{Key: "gospiderFrameId", Val: frameId})
	}
	nodeType := NodeType(data.Get("nodeType").Int())
	curNode := &html.Node{Type: nodeType.HtmlNodeType(), Attr: attrs}
	curNode.DataAtom = atom.Lookup(data.Get("localName").Bytes())
	switch nodeType {
	case NodeTypeText:
		curNode.Data = data.Get("nodeValue").String()
		if len(curNode.Data) == 10003 && strings.HasSuffix(curNode.Data, "…") && nodeId != 0 {
			rc, err := obj.webSock.DOMGetOuterHTML(ctx, nodeId, backendNodeId)
			if err != nil {
				return nil, err
			}
			jsonData, err := gson.Decode(rc.Result)
			if err != nil {
				return nil, err
			}
			outerHTML := jsonData.Get("outerHTML").String()
			if outerHTML != "" {
				curNode.Data = outerHTML
			}
		}
	case NodeTypeElement:
		curNode.Data = data.Get("localName").String()
	default:
		if curNode.Data = data.Get("nodeValue").String(); curNode.Data == "" {
			curNode.Data = data.Get("localName").String()
		}
	}
	for _, children := range data.Get("children").Array() {
		node, err := obj.parseJsonDom(ctx, children)
		if err != nil {
			return nil, err
		}
		if node != nil {
			curNode.AppendChild(node)
		}
	}
	for _, children := range data.Get("contentDocument.children").Array() {
		node, err := obj.parseJsonDom(ctx, children)
		if err != nil {
			return nil, err
		}
		if node != nil {
			curNode.AppendChild(node)
		}
	}
	return curNode, nil
}

type Dom struct {
	baseUrl string
	webSock *cdp.WebSock
	nodeId  int64
	frameId string
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
func (obj *Dom) NodeId() int64 {
	return obj.nodeId
}
func (obj *Dom) FrameId() string {
	return obj.frameId
}
func (obj *Dom) String() string {
	return obj.ele.String()
}
func (obj *Dom) SetHtml(ctx context.Context, contents string) error {
	_, err := obj.webSock.DOMSetOuterHTML(ctx, obj.nodeId, contents)
	return err
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
