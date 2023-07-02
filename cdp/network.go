package cdp

import (
	"context"
	"fmt"
	"strings"
)

type Cookie struct {
	Name         string  `json:"name,omitempty"`  //必填
	Value        string  `json:"value,omitempty"` //必填
	Url          string  `json:"url,omitempty"`
	Domain       string  `json:"domain,omitempty"` //必填
	Path         string  `json:"path,omitempty"`
	Secure       bool    `json:"secure,omitempty"`
	HttpOnly     bool    `json:"httpOnly,omitempty"`
	SameSite     string  `json:"sameSite,omitempty"`
	Expires      float64 `json:"expires,omitempty"`
	Priority     string  `json:"priority,omitempty"`
	SameParty    bool    `json:"sameParty,omitempty"`
	SourceScheme string  `json:"sourceScheme,omitempty"`
	SourcePort   int     `json:"sourcePort,omitempty"`
	PartitionKey int     `json:"partitionKey,omitempty"`
	Session      bool    `json:"session,omitempty"`
	Size         int64   `json:"size,omitempty"`
}
type Cookies []Cookie

func (obj Cookies) String() string {
	cooks := []string{}
	for _, cook := range obj {
		cooks = append(cooks, fmt.Sprintf("%s=%s", cook.Name, cook.Value))
	}
	return strings.Join(cooks, "; ")
}
func (obj Cookies) Gets(key string) []string {
	vals := []string{}
	for _, cook := range obj {
		if cook.Name == key {
			vals = append(vals, cook.Value)
		}
	}
	return vals
}
func (obj Cookies) Get(key string) (string, bool) {
	vals := obj.Gets(key)
	if l := len(vals); l == 0 {
		return "", false
	} else {
		return vals[l-1], true
	}
}
func (obj Cookies) Map() map[string][]string {
	data := map[string][]string{}
	for _, cook := range obj {
		dds, ok := data[cook.Name]
		if ok {
			dds = append(dds, cook.Value)
		} else {
			data[cook.Name] = []string{cook.Value}
		}
	}
	return data
}

func (obj *WebSock) NetworkSetCookies(preCtx context.Context, cookies []Cookie) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Network.setCookies",
		Params: map[string]any{
			"cookies": cookies,
		},
	})
}
func (obj *WebSock) NetworkEnable(preCtx context.Context) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Network.enable",
	})
}
func (obj *WebSock) NetworkClearBrowserCookies(preCtx context.Context) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Network.clearBrowserCookies",
	})
}
func (obj *WebSock) NetworkClearBrowserCache(preCtx context.Context) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Network.clearBrowserCache",
	})
}
func (obj *WebSock) NetworkGetCookies(preCtx context.Context, urls ...string) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Network.getCookies",
		Params: map[string]any{
			"urls": urls,
		},
	})
}
func (obj *WebSock) NetworkSetCacheDisabled(preCtx context.Context, cacheDisabled bool) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Network.setCacheDisabled",
		Params: map[string]any{
			"cacheDisabled": cacheDisabled,
		},
	})
}
