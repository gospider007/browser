package cdp

import (
	"context"
)

func (obj *WebSock) RuntimeEvaluate(ctx context.Context, expression string) (RecvData, error) {
	return obj.send(ctx, commend{
		Method: "Runtime.evaluate",
		Params: map[string]any{
			"awaitPromise":  true,
			"expression":    expression,
			"returnByValue": true,
		},
	})
}
