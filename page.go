package browser

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	uurl "net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gospider007/bs4"
	"github.com/gospider007/cdp"
	"github.com/gospider007/gson"
	"github.com/gospider007/re"
	"github.com/gospider007/requests"
	"github.com/gospider007/tools"
)

type Page struct {
	addr         string
	id           string
	mouseX       float64
	mouseY       float64
	ctx          context.Context
	cnl          context.CancelFunc
	preWebSock   *cdp.WebSock
	globalReqCli *requests.Client

	baseUrl          string
	webSock          *cdp.WebSock
	stealth          bool
	isReplaceRequest bool

	pageLoadId         int64
	pageStop           bool
	domLoad            bool
	loadNotices        chan struct{}
	stopNotices        chan struct{}
	networkNotices     chan struct{}
	networkNoticesSize atomic.Int64

	requestFunc  func(context.Context, *cdp.Route)
	responseFunc func(context.Context, *cdp.Route)

	iframes map[string]string
}

func defaultRequestFunc(ctx context.Context, r *cdp.Route) { r.RequestContinue(ctx) }
func (obj *Page) delStopNotice() {
	select {
	case <-obj.stopNotices:
	default:
	}
	select {
	case <-obj.loadNotices:
	default:
	}
}
func (obj *Page) delLoadNotice() {
	select {
	case <-obj.loadNotices:
	default:
	}
}
func (obj *Page) addStopNotice() {
	select {
	case obj.stopNotices <- struct{}{}:
	default:
	}
	obj.addLoadNotice()
}
func (obj *Page) addLoadNotice() {
	select {
	case obj.loadNotices <- struct{}{}:
	default:
	}
	obj.addNetworkNotice()
}
func (obj *Page) addNetworkNotice() {
	select {
	case obj.networkNotices <- struct{}{}:
	default:
	}
}
func (obj *Page) pageStartLoadMain(ctx context.Context, rd cdp.RecvData) {
	if obj.id == rd.Params["frameId"].(string) {
		if rd.Id > obj.pageLoadId {
			obj.pageLoadId = rd.Id
			obj.domLoad = false
			obj.pageStop = false
			clear(obj.iframes)
			obj.delStopNotice()
		}
	}
}
func (obj *Page) pageEndLoadMain(ctx context.Context, rd cdp.RecvData) {
	if obj.id == rd.Params["frameId"].(string) {
		if rd.Id > obj.pageLoadId {
			obj.pageLoadId = rd.Id
			obj.domLoad = true
			obj.pageStop = true
			obj.addStopNotice()
		}
	}
}
func (obj *Page) domLoadMain(ctx context.Context, rd cdp.RecvData) {
	obj.domLoad = true
	obj.addLoadNotice()
}
func (obj *Page) frameLoadMain(ctx context.Context, rd cdp.RecvData) {
	jsonData, err := gson.Decode(rd.Params)
	if err != nil {
		return
	}
	href := jsonData.Get("frame.url").String()
	frameId := jsonData.Get("frame.id").String()
	if href != "" && frameId != "" {
		obj.iframes[href] = frameId
	}
}
func (obj *Page) routeMain(ctx context.Context, rd cdp.RecvData) {
	obj.networkNoticesSize.Add(1)
	defer func() {
		obj.networkNoticesSize.Add(-1)
		obj.addNetworkNotice()
	}()
	routeData := cdp.RouteData{}
	if _, err := gson.Decode(rd.Params, &routeData); err == nil {
		route := cdp.NewRoute(obj.webSock, routeData)
		if route.IsResponse() {
			if obj.responseFunc != nil {
				obj.responseFunc(ctx, route)
				if !route.Used() {
					route.Continue(ctx)
				}
			} else {
				route.Continue(ctx)
			}
		} else {
			if strings.HasSuffix(route.Url(), "/favicon.ico") {
				route.FulFill(ctx, cdp.FulData{
					StatusCode: 404,
				})
			} else if obj.requestFunc != nil {
				obj.requestFunc(ctx, route)
				if !route.Used() {
					if obj.isReplaceRequest {
						route.RequestContinue(ctx)
					} else {
						route.Continue(ctx)
					}
				}
			} else {
				route.Continue(ctx)
			}
		}
	}
}

func (obj *Page) addEvent(method string, fun func(ctx context.Context, rd cdp.RecvData)) {
	obj.webSock.AddEvent(method, fun)
}

//go:embed stealthRaw.js
var stealthRaw string

func (obj *Page) init(globalReqCli *requests.Client, option PageOption) error {
	var err error
	if obj.webSock, err = cdp.NewWebSock(
		obj.ctx,
		globalReqCli,
		fmt.Sprintf("ws://%s/devtools/page/%s", obj.addr, obj.id),
		cdp.WebSockOption{
			Proxy: option.Proxy,
		},
	); err != nil {
		return err
	}
	obj.addEvent("Page.frameStartedLoading", obj.pageStartLoadMain)
	obj.addEvent("Page.frameStoppedLoading", obj.pageEndLoadMain)
	obj.addEvent("Page.domContentEventFired", obj.domLoadMain)
	obj.addEvent("Page.frameNavigated", obj.frameLoadMain)
	obj.addEvent("Fetch.requestPaused", obj.routeMain)
	if _, err = obj.webSock.PageEnable(obj.ctx); err != nil {
		return err
	}
	if option.Stealth || obj.stealth {
		if err = obj.AddScript(obj.ctx, createFp()); err != nil {
			return err
		}
	}
	return obj.AddScript(obj.ctx, `Object.defineProperty(window, "RTCPeerConnection",{"get":undefined});Object.defineProperty(window, "mozRTCPeerConnection",{"get":undefined});Object.defineProperty(window, "webkitRTCPeerConnection",{"get":undefined});`)
}

type FpOption struct {
	Browser         string //("chrome" | "firefox" | "safari" | "edge")
	Device          string //"mobile" | "desktop"
	OperatingSystem string //"windows" | "macos" | "linux" | "android" | "ios"
	UserAgent       string
	Locales         []string
	Locale          string
}

func createFp(options ...FpOption) string {
	screen := map[string]any{
		"availHeight":      672,
		"availWidth":       1280,
		"pixelDepth":       24,
		"height":           720,
		"width":            1280,
		"availTop":         0,
		"availLeft":        0,
		"colorDepth":       24,
		"innerHeight":      0,
		"outerHeight":      672,
		"outerWidth":       1280,
		"innerWidth":       0,
		"screenX":          0,
		"pageXOffset":      0,
		"pageYOffset":      0,
		"devicePixelRatio": 1.5,
		"clientWidth":      0,
		"clientHeight":     18,
		"hasHDR":           false,
	}
	audioCodecs := map[string]any{
		"ogg": "probably",
		"mp3": "probably",
		"wav": "probably",
		"m4a": "maybe",
		"aac": "probably",
	}
	videoCodecs := map[string]any{
		"ogg":  "probably",
		"h264": "probably",
		"webm": "probably",
	}
	battery := map[string]any{
		"charging":        true,
		"chargingTime":    0,
		"dischargingTime": nil,
		"level":           1,
	}
	videoCard := map[string]any{
		"vendor":   "Google Inc. (Intel)",
		"renderer": "ANGLE (Intel, Intel(R) UHD Graphics Direct3D11 vs_5_0 ps_5_0, D3D11)",
	}
	multimediaDevices := map[string]any{
		"speakers": []map[string]any{
			{
				"deviceId": "",
				"kind":     "audiooutput",
				"label":    "",
				"groupId":  "",
			},
		},
		"micros": []map[string]any{
			{
				"deviceId": "",
				"kind":     "audioinput",
				"label":    "",
				"groupId":  "",
			},
		},
		"webcams": []map[string]any{
			{
				"deviceId": "",
				"kind":     "videoinput",
				"label":    "",
				"groupId":  "",
			},
		},
	}
	appVersion := re.Sub("Mozilla/", "", requests.UserAgent)
	version := re.Search(`Chrome/(\d+)?\.`, requests.UserAgent).Group(1)
	navigator := map[string]any{
		"userAgent": requests.UserAgent,
		"userAgentData": map[string]any{
			"brands": []map[string]any{
				{
					"brand":   "Microsoft Edge",
					"version": version,
				},
				{
					"brand":   "Not;A=Brand",
					"version": "8",
				},
				{
					"brand":   "Chromium",
					"version": version,
				},
			},
			"mobile":   false,
			"platform": "Windows",
		},
		"language": "zh-CN",
		"languages": []string{
			"zh-CN",
			"en",
			"en-GB",
			"en-US",
		},
		"platform":            "Win32",
		"deviceMemory":        8,
		"hardwareConcurrency": 8,
		"maxTouchPoints":      10,
		"product":             "Gecko",
		"productSub":          "20030107",
		"vendor":              "Google Inc.",
		"vendorSub":           "",
		"doNotTrack":          nil,
		"appCodeName":         "Mozilla",
		"appName":             "Netscape",
		"appVersion":          appVersion,
		"webdriver":           false,
	}
	fp := map[string]any{
		"screen":            screen,
		"audioCodecs":       audioCodecs,
		"videoCodecs":       videoCodecs,
		"battery":           battery,
		"videoCard":         videoCard,
		"multimediaDevices": multimediaDevices,
		"navigator":         navigator,
		"userAgent":         requests.UserAgent,
		"historyLength":     5,
	}
	val, _ := gson.Encode(fp)
	return strings.ReplaceAll(stealthRaw, `"@@__gospiderFpData__@@"`, tools.BytesToString(val))
}
func (obj *Page) AddScript(ctx context.Context, script string) error {
	_, err := obj.webSock.PageAddScriptToEvaluateOnNewDocument(ctx, script)
	return err
}
func (obj *Page) Screenshot(ctx context.Context, rect cdp.Rect, options ...cdp.ScreenshotOption) ([]byte, error) {
	jsonData, err := obj.Eval(ctx, `()=>{return document.documentElement.scrollTop}`, nil)
	if err != nil {
		return nil, err
	}
	rect.Y += jsonData.Get("result.value").Float()
	rs, err := obj.webSock.PageCaptureScreenshot(ctx, rect, options...)
	if err != nil {
		return nil, err
	}
	imgData, ok := rs.Result["data"].(string)
	if !ok {
		return nil, errors.New("not img data")
	}
	return tools.Base64Decode(imgData)
}

func (obj *Page) Rect(ctx context.Context) (cdp.Rect, error) {
	rs, err := obj.webSock.PageGetLayoutMetrics(ctx)
	var result cdp.Rect
	if err != nil {
		return result, err
	}
	_, err = gson.Decode(rs.Result["cssContentSize"], &result)
	return result, err
}

func (obj *Page) Reload(ctx context.Context) error {
	_, err := obj.webSock.PageReload(ctx)
	return err
}
func (obj *Page) WaitPageStop(preCtx context.Context, timeout ...time.Duration) error {
	var ctx context.Context
	var cnl context.CancelFunc
	if preCtx == nil {
		preCtx = obj.ctx
	}
	if len(timeout) > 0 {
		ctx, cnl = context.WithTimeout(preCtx, timeout[0])
	} else {
		ctx, cnl = context.WithCancel(preCtx)
	}
	defer cnl()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-obj.ctx.Done():
			return obj.ctx.Err()
		case <-obj.stopNotices:
			if obj.pageStop {
				obj.addStopNotice()
				return nil
			}
		}
	}
}
func (obj *Page) WaitDomLoad(preCtx context.Context, timeout ...time.Duration) error {
	var ctx context.Context
	var cnl context.CancelFunc
	if preCtx == nil {
		preCtx = obj.ctx
	}
	if len(timeout) > 0 {
		ctx, cnl = context.WithTimeout(preCtx, timeout[0])
	} else {
		ctx, cnl = context.WithCancel(preCtx)
	}
	defer cnl()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-obj.ctx.Done():
			return obj.ctx.Err()
		case <-obj.loadNotices:
			if obj.domLoad {
				obj.addLoadNotice()
				return nil
			}
		}
	}
}
func (obj *Page) WaitNetwork(preCtx context.Context, timeout ...time.Duration) error {
	var ctx context.Context
	var cnl context.CancelFunc
	if preCtx == nil {
		preCtx = obj.ctx
	}
	if len(timeout) > 0 {
		ctx, cnl = context.WithTimeout(preCtx, timeout[0])
	} else {
		ctx, cnl = context.WithCancel(preCtx)
	}
	defer cnl()
	if obj.requestFunc == nil && obj.responseFunc == nil {
		obj.isReplaceRequest = true
		if err := obj.Request(ctx, nil); err != nil {
			return err
		}
	}
	msTime := time.Millisecond * 1500
	basTime := time.Millisecond * 300
	msN := int(msTime/basTime) + 1
	msT := time.NewTimer(basTime)
	defer msT.Stop()
	var zeroNum int
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-obj.ctx.Done():
			return obj.ctx.Err()
		case <-obj.networkNotices:
			if obj.pageStop && obj.networkNoticesSize.Load() <= 0 {
				zeroNum++
			} else {
				zeroNum = 0
			}
		case <-msT.C:
			if obj.pageStop && obj.networkNoticesSize.Load() <= 0 {
				zeroNum++
			} else {
				zeroNum = 0
			}
		}
		if zeroNum > msN {
			return nil
		}
		msT.Reset(basTime)
	}
}
func (obj *Page) GoTo(preCtx context.Context, url string) error {
	obj.baseUrl = url
	_, err := obj.webSock.PageNavigate(preCtx, url)
	return err
}

// ex:   ()=>{}  或者  (params)=>{}
func (obj *Page) Eval(ctx context.Context, expression string, params map[string]any) (*gson.Client, error) {
	var value string
	if params != nil {
		con, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		value = tools.BytesToString(con)
	}
	// log.Print(fmt.Sprintf(`(async %s)(%s)`, expression, value))
	rs, err := obj.webSock.RuntimeEvaluate(ctx, fmt.Sprintf(`(async %s)(%s)`, expression, value))
	if err != nil {
		return nil, err
	}
	return gson.Decode(rs.Result)
}
func (obj *Page) Close() (err error) {
	err = obj.close()
	obj.preWebSock.TargetCloseTarget(obj.id)
	obj.webSock.Close()
	obj.cnl()
	return
}
func (obj *Page) close() error {
	_, err := obj.globalReqCli.Request(context.TODO(), "get", fmt.Sprintf("http://%s/json/close/%s", obj.addr, obj.id), requests.RequestOption{
		TryNum:   10,
		DisProxy: true,
		ResultCallBack: func(ctx context.Context, c *requests.Client, r *requests.Response) error {
			switch r.Text() {
			case "Target is closing", fmt.Sprintf("No such target id: %s", obj.id):
			}
			if text := r.Text(); text == "Target is closing" || text == fmt.Sprintf("No such target id: %s", obj.id) {
				return nil
			}
			return errors.New("req close target error")
		},
	})
	return err
}

func (obj *Page) Done() <-chan struct{} {
	return obj.webSock.Done()
}
func (obj *Page) Request(ctx context.Context, RequestFunc func(context.Context, *cdp.Route)) error {
	if RequestFunc != nil {
		obj.requestFunc = RequestFunc
	} else if obj.isReplaceRequest {
		obj.requestFunc = defaultRequestFunc
	} else {
		obj.requestFunc = nil
	}
	var err error
	if obj.requestFunc != nil {
		_, err = obj.webSock.FetchRequestEnable(ctx)
	} else if obj.responseFunc == nil {
		_, err = obj.webSock.FetchDisable(ctx)
	}
	return err
}
func (obj *Page) Response(ctx context.Context, ResponseFunc func(context.Context, *cdp.Route)) error {
	obj.responseFunc = ResponseFunc
	var err error
	if obj.responseFunc != nil {
		_, err = obj.webSock.FetchResponseEnable(ctx)
	} else if obj.requestFunc == nil {
		_, err = obj.webSock.FetchDisable(ctx)
	}
	return err
}
func (obj *Page) Html(ctx context.Context, contents ...string) (*bs4.Client, error) {
	r, err := obj.webSock.DOMGetDocuments(ctx)
	if err != nil {
		return nil, err
	}
	data, err := gson.Decode(r.Result)
	if err != nil {
		return nil, err
	}
	return bs4.NewClientWithNode(cdp.ParseJsonDom(data.Get("root"))), nil
}
func (obj *Page) setHtml(ctx context.Context, nodeId int64, content string) error {
	_, err := obj.webSock.DOMSetOuterHTML(ctx, nodeId, content)
	return err
}
func (obj *Page) WaitSelector(ctx context.Context, selector string, timeouts ...time.Duration) (*Dom, error) {
	if ctx == nil {
		ctx = obj.ctx
	}
	var timeout time.Duration
	if len(timeouts) > 0 {
		timeout = timeouts[0]
	} else {
		timeout = time.Second * 30
	}
	startTime := time.Now()
	var t *time.Timer
	defer func() {
		if t != nil {
			t.Stop()
		}
	}()
	for time.Since(startTime) <= timeout {
		dom, err := obj.QuerySelector(ctx, selector)
		if err != nil {
			return nil, err
		}
		if dom != nil {
			return dom, nil
		}
		if t == nil {
			t = time.NewTimer(time.Millisecond * 500)
		} else {
			t.Reset(time.Millisecond * 500)
		}
		select {
		case <-t.C:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, errors.New("超时")
}

func (obj *Page) QuerySelector(ctx context.Context, selector string) (*Dom, error) {
	html, err := obj.Html(ctx)
	if err != nil {
		return nil, err
	}
	ele := html.Find(selector)
	if ele == nil {
		return nil, err
	}
	gospiderNodeId := ele.Get("gospiderNodeId")
	nodeId, err := strconv.Atoi(gospiderNodeId)
	if err != nil {
		return nil, err
	}

	dom := &Dom{
		baseUrl: obj.baseUrl,
		webSock: obj.webSock,
		nodeId:  int64(nodeId),
		ele:     ele,
	}
	return dom, err
}
func (obj *Page) QuerySelectorAll(ctx context.Context, selector string) ([]*Dom, error) {
	html, err := obj.Html(ctx)
	if err != nil {
		return nil, err
	}
	doms := []*Dom{}
	for _, ele := range html.Finds(selector) {
		gospiderNodeId := ele.Get("gospiderNodeId")
		nodeId, err := strconv.Atoi(gospiderNodeId)
		if err != nil {
			return nil, err
		}
		dom := &Dom{
			baseUrl: obj.baseUrl,
			webSock: obj.webSock,
			nodeId:  int64(nodeId),
			ele:     ele,
		}
		doms = append(doms, dom)
	}
	return doms, err
}

// 移动操作
func (obj *Page) baseMove(ctx context.Context, x, y float64, kind int, steps ...int) error {
	var step int
	if len(steps) > 0 {
		step = steps[0]
	}
	if step < 1 {
		step = 1
	}
	for _, poi := range tools.GetTrack(
		[2]float64{obj.mouseX, obj.mouseY},
		[2]float64{obj.mouseX + x, obj.mouseY + y},
		float64(step),
	) {
		switch kind {
		case 0:
			if err := obj.move(ctx, cdp.Point{
				X: poi[0],
				Y: poi[1],
			}); err != nil {
				return err
			}
		case 1:
			if err := obj.touchMove(ctx, cdp.Point{
				X: poi[0],
				Y: poi[1],
			}); err != nil {
				return err
			}
		default:
			return errors.New("not found kind")
		}
	}
	obj.mouseX = obj.mouseX + x
	obj.mouseY = obj.mouseY + y
	return nil
}

func (obj *Page) Move(ctx context.Context, x, y float64, steps ...int) error {
	return obj.baseMove(ctx, x, y, 0, steps...)
}

func (obj *Page) move(ctx context.Context, point cdp.Point) error {
	_, err := obj.webSock.InputDispatchMouseEvent(ctx,
		cdp.DispatchMouseEventOption{
			Type: "mouseMoved",
			X:    point.X,
			Y:    point.Y,
		})
	if err != nil {
		return err
	}
	obj.mouseX = point.X
	obj.mouseY = point.Y
	return nil
}
func (obj *Page) TouchMove(ctx context.Context, x, y float64, steps ...int) error {
	return obj.baseMove(ctx, x, y, 1, steps...)
}
func (obj *Page) touchMove(ctx context.Context, point cdp.Point) error { //不需要delta
	_, err := obj.webSock.InputDispatchTouchEvent(ctx, "touchMove", []cdp.Point{
		{
			X: point.X,
			Y: point.Y,
		},
	})
	if err != nil {
		return err
	}
	obj.mouseX = point.X
	obj.mouseY = point.Y
	return nil
}
func (obj *Page) Wheel(ctx context.Context, x, y float64) error {
	_, err := obj.webSock.InputDispatchMouseEvent(ctx,
		cdp.DispatchMouseEventOption{
			Type:   "mouseWheel",
			DeltaX: x,
			DeltaY: y,
		})
	return err
}
func (obj *Page) Down(ctx context.Context, point cdp.Point) error {
	_, err := obj.webSock.InputDispatchMouseEvent(ctx,
		cdp.DispatchMouseEventOption{
			Type:       "mousePressed",
			Button:     "left",
			X:          point.X,
			Y:          point.Y,
			ClickCount: 1,
		})
	if err != nil {
		return err
	}
	obj.mouseX = point.X
	obj.mouseY = point.Y
	return err
}
func (obj *Page) TouchDown(ctx context.Context, point cdp.Point) error {
	_, err := obj.webSock.InputDispatchTouchEvent(ctx, "touchStart",
		[]cdp.Point{
			{
				X: point.X,
				Y: point.Y,
			},
		})
	if err != nil {
		return err
	}
	obj.mouseX = point.X
	obj.mouseY = point.Y
	return nil
}
func (obj *Page) Up(ctx context.Context) error {
	_, err := obj.webSock.InputDispatchMouseEvent(ctx, cdp.DispatchMouseEventOption{
		Type:       "mouseReleased",
		Button:     "left",
		X:          obj.mouseX,
		Y:          obj.mouseY,
		ClickCount: 1,
	})
	return err
}
func (obj *Page) TouchUp(ctx context.Context) error {
	_, err := obj.webSock.InputDispatchTouchEvent(ctx,
		"touchEnd",
		[]cdp.Point{})
	return err
}
func (obj *Page) Click(ctx context.Context, point cdp.Point) error {
	if err := obj.Down(ctx, point); err != nil {
		return err
	}
	return obj.Up(ctx)
}
func (obj *Page) TouchClick(ctx context.Context, point cdp.Point) error {
	if err := obj.TouchDown(ctx, point); err != nil {
		return err
	}
	return obj.TouchUp(ctx)
}

// 设置移动设备的属性
func (obj *Page) SetDevice(ctx context.Context, device cdp.Device) error {
	if err := obj.SetUserAgent(ctx, device.UserAgent); err != nil {
		return err
	}
	if err := obj.SetTouch(ctx, device.HasTouch); err != nil {
		return err
	}
	return obj.SetDeviceMetrics(ctx, device)
}

func (obj *Page) SetUserAgent(ctx context.Context, userAgent string) error {
	_, err := obj.webSock.EmulationSetUserAgentOverride(ctx, userAgent)
	return err
}

// 设置设备指标
func (obj *Page) SetDeviceMetrics(ctx context.Context, device cdp.Device) error {
	_, err := obj.webSock.EmulationSetDeviceMetricsOverride(ctx, device)
	return err
}

// 设置设备是否支持触摸
func (obj *Page) SetTouch(ctx context.Context, hasTouch bool) error {
	_, err := obj.webSock.EmulationSetTouchEmulationEnabled(ctx, hasTouch)
	return err
}

func (obj *Page) SetCookies(ctx context.Context, href string, cookies ...cdp.Cookie) error {
	if len(cookies) == 0 {
		return nil
	}
	if href == "" {
		href = obj.baseUrl
	}
	var err error
	for i := 0; i < len(cookies); i++ {
		if cookies[i].Domain == "" {
			if cookies[i].Url == "" {
				cookies[i].Url = href
			}
			if cookies[i].Url != "" {
				us, err := uurl.Parse(cookies[i].Url)
				if err != nil {
					return err
				}
				cookies[i].Domain = us.Hostname()
			}
		}
	}
	_, err = obj.webSock.NetworkSetCookies(ctx, cookies)
	return err
}
func (obj *Page) GetCookies(ctx context.Context, urls ...string) (cdp.Cookies, error) {
	if len(urls) == 0 {
		urls = append(urls, obj.baseUrl)
	}
	rs, err := obj.webSock.NetworkGetCookies(ctx, urls...)
	if err != nil {
		return nil, err
	}
	jsonData, err := gson.Decode(rs.Result)
	if err != nil {
		return nil, err
	}
	result := []cdp.Cookie{}
	for _, cookie := range jsonData.Get("cookies").Array() {
		var cook cdp.Cookie
		if _, err = gson.Decode(cookie.Raw(), &cook); err != nil {
			return result, err
		}
		result = append(result, cook)
	}
	return result, nil
}

func (obj *Page) ClearCookies(ctx context.Context) (err error) {
	_, err = obj.webSock.NetworkClearBrowserCookies(ctx)
	return
}
func (obj *Page) ClearCache(ctx context.Context) (err error) {
	_, err = obj.webSock.NetworkClearBrowserCache(ctx)
	return
}
func (obj *Page) ClearStorage(ctx context.Context) (err error) {
	_, err = obj.webSock.StorageClear(ctx, obj.baseUrl)
	return
}
