package cdp

import "context"

func (obj *WebSock) TargetCreateTarget(preCtx context.Context, url string) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Target.createTarget",
		Params: map[string]any{
			"url": url,
		},
	})
}
func (obj *WebSock) TargetCloseTarget(targetId string) (RecvData, error) {
	return obj.send(obj.ctx, commend{
		Method: "Target.closeTarget",
		Params: map[string]any{
			"targetId": targetId,
		},
	})
}
