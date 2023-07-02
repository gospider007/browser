package cdp

import "context"

func (obj *WebSock) StorageClear(preCtx context.Context, href string) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Storage.clearDataForOrigin",
		Params: map[string]any{
			"origin":       href,
			"storageTypes": "all",
		},
	})
}
