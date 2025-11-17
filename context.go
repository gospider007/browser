package browser

import (
	"context"
	"errors"

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
	page := &Page{
		option:           option,
		targetId:         targetId,
		targetType:       "page",
		addr:             obj.addr,
		ctx:              ctx,
		cnl:              cnl,
		globalReqCli:     obj.globalReqCli,
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
func (obj *BrowserContext) Done() <-chan struct{} {
	return obj.webSock.Done()
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
