package browser

import (
	"context"
	"errors"
	"fmt"

	"github.com/gospider007/cdp"
	"github.com/gospider007/gson"
	"github.com/gospider007/requests"
	"github.com/gospider007/tools"
)

type BrowserContext struct {
	browserContextId string
	globalReqCli     *requests.Client
	stealth          bool
	webSock          *cdp.WebSock
	option           ClientOption
	ctx              context.Context
	cnl              context.CancelCauseFunc
	addr             string
}

type BrowserContextOption struct {
	Proxy          any
	Stealth        bool //是否开启随机指纹
	MaxRetries     int
	GetProxy       func(ctx *requests.Response) (any, error)
	ResultCallBack func(ctx *requests.Response) error
}

func (obj *Client) newBrowserContext(ctx context.Context) (string, error) {
	contextData, err := obj.webSock.TargetCreateBrowserContext(ctx)
	if err != nil {
		return "", err
	}
	contextResult, err := gson.Decode(contextData.Result)
	if err != nil {
		return "", err
	}
	browserContextId := contextResult.Get("browserContextId").String()
	if browserContextId == "" {
		return "", errors.New("not found browserContextId")
	}
	return browserContextId, nil
}
func (obj *Client) NewBrowserContext(preCtx context.Context, options ...requests.ClientOption) (*BrowserContext, error) {
	var option requests.ClientOption
	if len(options) > 0 {
		option = options[0]
	}
	if preCtx == nil {
		preCtx = obj.ctx
	}
	browserContext := &BrowserContext{
		addr:   obj.addr,
		option: obj.option,
	}
	var err error
	browserContext.ctx, browserContext.cnl = context.WithCancelCause(obj.ctx)
	browserContext.globalReqCli, err = obj.globalReqCli.Clone(browserContext.ctx)
	if err != nil {
		return nil, err
	}
	if err = tools.Merge(browserContext.globalReqCli.ClientOption, option); err != nil {
		return nil, err
	}
	browserContext.browserContextId, err = obj.newBrowserContext(preCtx)
	if err != nil {
		return nil, err
	}
	return browserContext, browserContext.init(obj.browserContextId)
}
func (obj *BrowserContext) init(browserContextId string) (err error) {
	defer func() {
		if err != nil {
			obj.Close()
		}
	}()
	obj.webSock, err = cdp.NewWebSock(
		obj.ctx,
		obj.globalReqCli,
		fmt.Sprintf("ws://%s/devtools/browser/%s", obj.addr, browserContextId),
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
	var option PageOption
	if len(options) > 0 {
		option = options[0]
	}
	if !option.Stealth {
		option.Stealth = obj.option.Stealth
	}
	return newPageWithTargetId(obj, nil, preCtx, targetId, option)
}
func newPageWithTargetId(browserContext *BrowserContext, pageCtx context.Context, preCtx context.Context, targetId string, option PageOption) (*Page, error) {
	if pageCtx == nil {
		pageCtx = browserContext.ctx
	}
	ctx, cnl := context.WithCancel(pageCtx)
	globalReqCli, err := browserContext.globalReqCli.Clone(ctx)
	if err != nil {
		cnl()
		return nil, err
	}
	err = tools.Merge(globalReqCli.ClientOption, option.Option)
	if err != nil {
		cnl()
		return nil, err
	}
	page := &Page{
		option:         option,
		targetId:       targetId,
		targetType:     "page",
		browserContext: browserContext,
		ctx:            ctx,
		cnl:            cnl,
		globalReqCli:   globalReqCli,
		loadNotices:    make(chan struct{}, 1),
		stopNotices:    make(chan struct{}, 1),
		networkNotices: make(chan struct{}, 1),
	}
	if err := page.init(preCtx); err != nil {
		return nil, err
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
