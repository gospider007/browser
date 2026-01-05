package browser

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"time"

	"github.com/gospider007/cdp"
	"github.com/gospider007/cmd"
	"github.com/gospider007/re"
	"github.com/gospider007/requests"
	"github.com/gospider007/tools"
)

type Client struct {
	cmdCli           *cmd.Client
	globalReqCli     *requests.Client
	addr             string
	ctx              context.Context
	cnl              context.CancelCauseFunc
	webSock          *cdp.WebSock
	option           *ClientOption
	browserContext   *BrowserContext
	browserContextId string
}
type ClientOption struct {
	Host           string
	Port           int
	ChromePath     string   //chrome path
	UserDir        string   //user dir
	Args           []string //start args
	Headless       bool     //is headless
	UserAgent      string
	Proxy          any                                       //support http,https,socks5,ex: http://127.0.0.1:7005
	GetProxy       func(ctx *requests.Response) (any, error) //pr
	Width          int64                                     //browser width,1200
	Height         int64                                     //browser height,605
	Stealth        bool                                      //is stealth
	Ja3Spec        any                                       //ja3
	MaxRetries     int
	ResultCallBack func(ctx *requests.Response) error
}

// 新建浏览器
func NewClient(preCtx context.Context, options ...ClientOption) (*Client, error) {
	var option ClientOption
	if len(options) > 0 {
		option = options[0]
	}
	if preCtx == nil {
		preCtx = context.TODO()
	}
	if option.MaxRetries == 0 {
		option.MaxRetries = 2
	}
	ctx, cnl := context.WithCancelCause(preCtx)

	globalReqCli, err := requests.NewClient(ctx, requests.ClientOption{
		ResultCallBack: option.ResultCallBack,
		MaxRetries:     option.MaxRetries,
		Proxy:          option.Proxy,
		GetProxy:       option.GetProxy,
		DisJar:         true,
		MaxRedirect:    -1,
	})
	if err != nil {
		cnl(nil)
		return nil, err
	}
	if runtime.GOOS == "linux" {
		option.Headless = true
	}
	if option.Width == 0 {
		option.Width = 1300
	}
	if option.Height == 0 {
		option.Height = 800
	}
	if option.UserAgent == "" {
		option.UserAgent = tools.UserAgent
	}
	client := &Client{
		option:       &option,
		globalReqCli: globalReqCli,
		ctx:          ctx,
		cnl:          cnl,
	}
	if option.Host == "" || option.Port == 0 {
		if err = client.runChrome(); err != nil {
			client.Close()
			return nil, err
		}
		client.addr = net.JoinHostPort(option.Host, strconv.Itoa(option.Port))
	} else {
		client.addr = net.JoinHostPort(option.Host, strconv.Itoa(option.Port))
	}
	go tools.Signal(ctx, client.Close)
	if err = client.init(); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

func (obj *Client) RequestClient() *requests.Client {
	if obj.browserContext != nil {
		return obj.browserContext.RequestClient()
	}
	return obj.globalReqCli
}

// 浏览器初始化
func (obj *Client) init() (err error) {
	defer func() {
		if err != nil {
			obj.Close()
		}
	}()
	var resp *requests.Response
	resp, err = obj.globalReqCli.Request(obj.ctx, "get",
		fmt.Sprintf("http://%s/json/version", obj.addr),
		requests.RequestOption{
			Timeout: time.Second * 3,
			ErrCallBack: func(ctx *requests.Response) error {
				select {
				case <-obj.cmdCli.Ctx().Done():
					return context.Cause(obj.cmdCli.Ctx())
				case <-time.After(time.Second):
				}
				return nil
			},
			ResultCallBack: func(ctx *requests.Response) error {
				if ctx.StatusCode() == 200 {
					return nil
				}
				time.Sleep(time.Second)
				return errors.New("code error")
			},
			MaxRetries: 10,
			DisProxy:   true,
		})
	if err != nil {
		if obj.cmdCli.Err() != nil {
			return obj.cmdCli.Err()
		}
		return err
	}
	jsonData, err := resp.Json()
	if err != nil {
		return err
	}
	wsUrl := jsonData.Get("webSocketDebuggerUrl").String()
	if wsUrl == "" {
		return errors.New("not fouond browser wsUrl")
	}
	browWsRs := re.Search(`devtools/browser/(.*)`, wsUrl)
	if browWsRs == nil {
		return errors.New("not fouond browser id")
	}
	obj.browserContextId = browWsRs.Group(1)
	obj.webSock, err = cdp.NewWebSock(
		obj.ctx,
		obj.globalReqCli,
		fmt.Sprintf("ws://%s/devtools/browser/%s", obj.addr, obj.browserContextId),
	)
	return err
}

// 浏览器初始化

// 浏览器是否结束的 chan

func (obj *Client) Context() context.Context {
	return obj.webSock.Context()
}
func (obj *Client) Error() (err error) {
	return obj.webSock.Error()
}

// 返回浏览器远程控制的地址
func (obj *Client) Addr() string {
	return obj.addr
}

// 关闭浏览器
func (obj *Client) Close() {
	if obj.globalReqCli != nil {
		obj.globalReqCli.Close()
	}
	if obj.browserContext != nil {
		obj.browserContext.Close()
	}
	if obj.webSock != nil {
		obj.webSock.BrowserClose()
		obj.webSock.CloseWithError(errors.New("webSock browser closed"))
	}
	if obj.cmdCli != nil {
		obj.cmdCli.Close()
	}
	obj.cnl(errors.New("client browser closed"))
}

type PageOption struct {
	Option      requests.ClientOption
	Stealth     bool //是否开启随机指纹
	requestFunc func(context.Context, *cdp.Route)
}

// 新建标签页
func (obj *Client) NewPage(preCtx context.Context, options ...PageOption) (*Page, error) {
	var err error
	if obj.browserContext == nil {
		obj.browserContext, err = obj.NewBrowserContext(preCtx)
		if err != nil {
			return nil, err
		}
	}
	return obj.browserContext.NewPage(preCtx, options...)
}

func (obj *Client) BrowserSetPermission(ctx context.Context, permission string, setting string, origins ...string) error {
	return obj.webSock.BrowserSetPermission(ctx, permission, setting, origins...)
}
func (obj *Client) BrowserGrantPermissions(ctx context.Context, permissions []string, origins ...string) error {
	return obj.webSock.BrowserGrantPermissions(ctx, permissions, origins...)
}
