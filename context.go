package browser

import (
	"context"
	"errors"
	"fmt"

	"github.com/gospider007/cdp"
	"github.com/gospider007/requests"
)

type BrowserContext struct {
	isReplaceRequest bool
	browserContextId string
	globalReqCli     *requests.Client
	webSock          *cdp.WebSock
	option           *BrowserContextOption
	ctx              context.Context
	cnl              context.CancelCauseFunc
	addr             string
}

func (obj *BrowserContext) init(contextId string) (err error) {
	defer func() {
		if err != nil {
			obj.Close()
		}
	}()
	obj.webSock, err = cdp.NewWebSock(
		obj.ctx,
		obj.globalReqCli,
		fmt.Sprintf("ws://%s/devtools/browser/%s", obj.addr, contextId),
		requests.RequestOption{},
	)
	return err
}
func (obj *BrowserContext) NewPage(preCtx context.Context, options ...PageOption) (*Page, error) {
	if preCtx == nil {
		preCtx = obj.ctx
	}
	rs, err := obj.webSock.TargetCreateTarget(preCtx, obj.browserContextId, "")
	if err != nil {
		return nil, err
	}
	targetId, ok := rs.Result["targetId"].(string)
	if !ok {
		return nil, errors.New("not found targetId")
	}
	return obj.NewPageWithTargetId(preCtx, targetId, options...)
}
func (obj *BrowserContext) NewPageWithTargetId(preCtx context.Context, targetId string, options ...PageOption) (*Page, error) {
	var option PageOption
	if len(options) > 0 {
		option = options[0]
	}
	isReplaceRequest := obj.isReplaceRequest
	if !isReplaceRequest {
		if option.Option.Proxy != nil && option.Option.Proxy != obj.option.Proxy {
			if p, ok := option.Option.Proxy.(string); ok && p != obj.option.Proxy {
				isReplaceRequest = true
			}
		}
	}
	if option.Option.Proxy == "" {
		option.Option.Proxy = obj.option.Proxy
	}
	if !option.Stealth {
		option.Stealth = obj.option.Stealth
	}
	ctx, cnl := context.WithCancel(obj.ctx)
	globalReqCli, err := obj.globalReqCli.Clone(ctx)
	if err != nil {
		cnl()
		return nil, err
	}
	page := &Page{
		option:           option,
		targetId:         targetId,
		targetType:       "page",
		addr:             obj.addr,
		ctx:              ctx,
		cnl:              cnl,
		globalReqCli:     globalReqCli,
		isReplaceRequest: isReplaceRequest,
		loadNotices:      make(chan struct{}, 1),
		stopNotices:      make(chan struct{}, 1),
		networkNotices:   make(chan struct{}, 1),
	}
	if err := page.init(); err != nil {
		return nil, err
	}
	if isReplaceRequest {
		// log.Print("enabel replace request...")
		if err := page.Request(preCtx, defaultRequestFunc); err != nil {
			return nil, err
		}
	}
	return page, nil
}
func (obj *BrowserContext) Addr() string {
	return obj.addr
}
func (obj *BrowserContext) Error() (err error) {
	return obj.webSock.Error()
}
func (obj *BrowserContext) Context() context.Context {
	return obj.webSock.Context()
}

func (obj *BrowserContext) Close() error {
	if obj.webSock != nil {
		obj.webSock.TargetDisposeBrowserContext(obj.browserContextId)
	}
	if obj.globalReqCli != nil {
		obj.globalReqCli.Close()
	}
	obj.cnl(errors.New("context closed"))
	return nil
}
func (obj *BrowserContext) RequestClient() *requests.Client {
	return obj.globalReqCli
}
