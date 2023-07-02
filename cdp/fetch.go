package cdp

import (
	"context"
	"net/http"

	"gitee.com/baixudong/gospider/tools"
)

func (obj *WebSock) FetchRequestEnable(preCtx context.Context) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Fetch.enable",

		Params: map[string]any{
			"patterns": []map[string]any{
				{
					"requestStage": "Request",
				},
			},
		},
	})
}
func (obj *WebSock) FetchResponseEnable(preCtx context.Context) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Fetch.enable",
		Params: map[string]any{
			"patterns": []map[string]any{
				{
					"requestStage": "Response",
				},
			},
		},
	})
}
func (obj *WebSock) FetchDisable(preCtx context.Context) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Fetch.disable",
	})
}
func (obj *WebSock) FetchContinueRequest(preCtx context.Context, requestId string, options ...RequestOption) (RecvData, error) {
	var option RequestOption
	if len(options) > 0 {
		option = options[0]
	}
	params := map[string]any{
		"requestId": requestId,
	}
	if option.Headers != nil {
		headers := []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}{}
		for hk, hvs := range option.Headers {
			for _, hv := range hvs {
				headers = append(headers, struct {
					Name  string "json:\"name\""
					Value string "json:\"value\""
				}{
					Name:  hk,
					Value: hv,
				})
			}
		}
		params["headers"] = headers
	}
	if option.Url != "" {
		params["url"] = option.Url
	}
	if option.Method != "" {
		params["method"] = option.Method
	}
	if option.PostData != "" {
		params["postData"] = tools.Base64Encode(option.PostData)
	}
	return obj.send(preCtx, commend{
		Method: "Fetch.continueRequest",
		Params: params,
	})
}
func (obj *WebSock) FetchFailRequest(preCtx context.Context, requestId, errorReason string) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Fetch.failRequest",
		Params: map[string]any{
			"requestId":   requestId,
			"errorReason": errorReason,
		},
	})
}
func (obj *WebSock) FetchGetResponseBody(preCtx context.Context, requestId string) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Fetch.getResponseBody",
		Params: map[string]any{
			"requestId": requestId,
		},
	})
}
func (obj *WebSock) FetchFulfillRequest(preCtx context.Context, requestId string, fulData FulData) (RecvData, error) {
	if fulData.Headers == nil {
		fulData.Headers = http.Header{
			"Content-Type": []string{tools.GetContentTypeWithBytes(tools.StringToBytes(fulData.Body))},
		}
	}
	headers := []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}{}
	for hk, hvs := range fulData.Headers {
		for _, hv := range hvs {
			headers = append(headers, struct {
				Name  string "json:\"name\""
				Value string "json:\"value\""
			}{
				Name:  hk,
				Value: hv,
			})
		}
	}
	if fulData.StatusCode == 0 {
		fulData.StatusCode = 200
	}
	if fulData.ResponsePhrase == "" {
		fulData.ResponsePhrase = "200 OK"
	}

	return obj.send(preCtx, commend{
		Method: "Fetch.fulfillRequest",
		Params: map[string]any{
			"requestId":       requestId,
			"responseCode":    fulData.StatusCode,
			"responseHeaders": headers,
			"body":            tools.Base64Encode(fulData.Body),
			"responsePhrase":  fulData.ResponsePhrase,
		},
	})
}
