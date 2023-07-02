package cdp

import (
	"context"
)

func (obj *WebSock) DOMEnable(ctx context.Context) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "DOM.enable",
	})
}
func (obj *WebSock) DOMDescribeNode(ctx context.Context, nodeId, backendNodeId int64) (RecvData, error) {
	params := map[string]any{
		"depth": 0,
	}
	if backendNodeId != 0 {
		params["backendNodeId"] = backendNodeId
	} else {
		params["nodeId"] = nodeId
	}
	return obj.send(ctx, commend{
		Method: "DOM.describeNode",
		Params: params,
	})
}
func (obj *WebSock) DOMResolveNode(ctx context.Context, backendNodeId int64) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "DOM.resolveNode",
		Params: map[string]any{
			"backendNodeId": backendNodeId,
		},
	})
}
func (obj *WebSock) DOMGetFrameOwner(ctx context.Context, frameId string) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "DOM.getFrameOwner",
		Params: map[string]any{
			"frameId": frameId,
		},
	})
}

func (obj *WebSock) DOMRequestNode(ctx context.Context, objectId string) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "DOM.requestNode",
		Params: map[string]any{
			"objectId": objectId,
		},
	})
}

func (obj *WebSock) DOMSetOuterHTML(ctx context.Context, nodeId int64, outerHTML string) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "DOM.setOuterHTML",
		Params: map[string]any{
			"nodeId":    nodeId,
			"outerHTML": outerHTML,
		},
	})
}
func (obj *WebSock) DOMGetOuterHTML(ctx context.Context, nodeId int64, backendNodeId int64) (RecvData, error) {
	params := map[string]any{}
	if backendNodeId != 0 {
		params["backendNodeId"] = backendNodeId
	} else {
		params["nodeId"] = nodeId
	}
	return obj.send(ctx, commend{
		Method: "DOM.getOuterHTML",
		Params: params,
	})
}
func (obj *WebSock) DOMFocus(ctx context.Context, nodeId int64) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "DOM.focus",
		Params: map[string]any{
			"nodeId": nodeId,
		},
	})
}
func (obj *WebSock) DOMQuerySelector(ctx context.Context, nodeId int64, selector string) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "DOM.querySelector",
		Params: map[string]any{
			"nodeId":   nodeId,
			"selector": selector,
		},
	})
}
func (obj *WebSock) DOMQuerySelectorAll(ctx context.Context, nodeId int64, selector string) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "DOM.querySelectorAll",
		Params: map[string]any{
			"nodeId":   nodeId,
			"selector": selector,
		},
	})
}
func (obj *WebSock) DOMGetBoxModel(ctx context.Context, nodeId int64) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "DOM.getBoxModel",
		Params: map[string]any{
			"nodeId": nodeId,
		},
	})
}
func (obj *WebSock) DOMGetDocument(ctx context.Context) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "DOM.getDocument",
		Params: map[string]any{
			"depth": 0,
		},
	})
}
func (obj *WebSock) DOMScrollIntoViewIfNeeded(ctx context.Context, nodeId int64) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "DOM.scrollIntoViewIfNeeded",
		Params: map[string]any{
			"nodeId": nodeId,
		},
	})
}
