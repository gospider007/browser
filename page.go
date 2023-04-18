package browser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	uurl "net/url"
	"strings"
	"time"

	"gitee.com/baixudong/gospider/bs4"
	"gitee.com/baixudong/gospider/cdp"
	"gitee.com/baixudong/gospider/db"
	"gitee.com/baixudong/gospider/re"
	"gitee.com/baixudong/gospider/requests"
	"gitee.com/baixudong/gospider/tools"

	"github.com/tidwall/gjson"
)

type Page struct {
	host         string
	port         int
	id           string
	mouseX       float64
	mouseY       float64
	ctx          context.Context
	cnl          context.CancelFunc
	preWebSock   *cdp.WebSock
	globalReqCli *requests.Client
	headless     bool
	nodeId       int64
	baseUrl      string
	webSock      *cdp.WebSock
	stealth      bool
	dataCache    bool

	pageStarId int64
	pageEndId  int64
	pageDone   chan struct{}
}

func defaultRequestFunc(ctx context.Context, r *cdp.Route) { r.RequestContinue(ctx) }
func (obj *Page) pageStop() bool {
	return obj.pageEndId >= obj.pageStarId
}
func (obj *Page) pageStartLoadMain(ctx context.Context, rd cdp.RecvData) {
	if obj.id == rd.Params["frameId"].(string) {
		if rd.Id > obj.pageStarId {
			obj.pageStarId = rd.Id
		}
	}
	select {
	case obj.pageDone <- struct{}{}:
	default:
	}
}
func (obj *Page) pageEndLoadMain(ctx context.Context, rd cdp.RecvData) {
	if obj.id == rd.Params["frameId"].(string) {
		if rd.Id > obj.pageEndId {
			obj.pageEndId = rd.Id
		}
	}
	select {
	case obj.pageDone <- struct{}{}:
	default:
	}
}
func (obj *Page) addEvent(method string, fun func(ctx context.Context, rd cdp.RecvData)) {
	obj.webSock.AddEvent(method, fun)
}
func (obj *Page) delEvent(method string) {
	obj.webSock.DelEvent(method)
}
func (obj *Page) init(globalReqCli *requests.Client, option PageOption, db *db.Client[cdp.FulData]) error {
	var err error
	if obj.webSock, err = cdp.NewWebSock(
		obj.ctx,
		globalReqCli,
		fmt.Sprintf("ws://%s:%d/devtools/page/%s", obj.host, obj.port, obj.id),
		cdp.WebSockOption{
			Proxy:     option.Proxy,
			DataCache: option.DataCache,
			Ja3Spec:   option.Ja3Spec,
			Ja3:       option.Ja3,
		},
		db,
	); err != nil {
		return err
	}
	obj.addEvent("Page.frameStartedLoading", obj.pageStartLoadMain)
	obj.addEvent("Page.frameStoppedLoading", obj.pageEndLoadMain)
	if _, err = obj.webSock.PageEnable(obj.ctx); err != nil {
		return err
	}
	// if obj.headless {
	// 	if err = obj.AddScript(obj.ctx, stealth); err != nil {
	// 		return err
	// 	}
	// 	if err = obj.AddScript(obj.ctx, stealth3); err != nil {
	// 		return err
	// 	}
	// }
	if option.Stealth || obj.stealth {
		if err = obj.AddScript(obj.ctx, stealth2); err != nil {
			return err
		}
	}
	return obj.AddScript(obj.ctx, `Object.defineProperty(window, "RTCPeerConnection",{"get":undefined});Object.defineProperty(window, "mozRTCPeerConnection",{"get":undefined});Object.defineProperty(window, "webkitRTCPeerConnection",{"get":undefined});`)
}
func (obj *Page) AddScript(ctx context.Context, script string) error {
	_, err := obj.webSock.PageAddScriptToEvaluateOnNewDocument(ctx, script)
	return err
}
func (obj *Page) Screenshot(ctx context.Context, options ...cdp.ScreenshotOption) ([]byte, error) {
	rs, err := obj.webSock.PageCaptureScreenshot(ctx, cdp.Rect{}, options...)
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
	return result, tools.Any2struct(rs.Result["cssContentSize"], &result)
}
func (obj *Page) Reload(ctx context.Context) error {
	_, err := obj.webSock.PageReload(ctx)
	if err != nil {
		return err
	}
	return obj.WaitStop(ctx)
}
func (obj *Page) PageLoadDone() <-chan struct{} {
	return obj.pageDone
}
func (obj *Page) LoadId() int64 {
	return obj.pageStarId
}
func (obj *Page) WaitStop(preCtx context.Context, waits ...int) error {
	var ctx context.Context
	var cnl context.CancelFunc
	wait := 2
	if len(waits) > 0 {
		wait = waits[0]
	}
	if preCtx == nil {
		ctx, cnl = context.WithTimeout(obj.ctx, time.Second*30)
		defer cnl()
	} else {
		ctx = preCtx
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-obj.ctx.Done():
			return obj.ctx.Err()
		case <-obj.pageDone:
		case <-time.After(time.Second * time.Duration(wait)):
			if obj.pageStop() {
				return nil
			}
		}
	}
}
func (obj *Page) GoTo(preCtx context.Context, url string) error {
	var err error
	obj.baseUrl = url
	var ctx context.Context
	var cnl context.CancelFunc
	if preCtx == nil {
		ctx, cnl = context.WithTimeout(obj.ctx, time.Second*30)
		defer cnl()
	} else {
		ctx = preCtx
	}
	_, err = obj.webSock.PageNavigate(ctx, url)
	if err != nil {
		return err
	}
	return obj.WaitStop(ctx)
}

// ex:   ()=>{}  或者  (params)=>{}
func (obj *Page) Eval(ctx context.Context, expression string, params map[string]any) (gjson.Result, error) {
	var value string
	if params != nil {
		con, err := json.Marshal(params)
		if err != nil {
			return gjson.Result{}, err
		}
		value = tools.BytesToString(con)
	}
	// log.Print(fmt.Sprintf(`(async %s)(%s)`, expression, value))
	rs, err := obj.webSock.RuntimeEvaluate(ctx, fmt.Sprintf(`(async %s)(%s)`, expression, value))
	return tools.Any2json(rs.Result), err
}
func (obj *Page) Close() error {
	defer obj.cnl()
	_, err := obj.preWebSock.TargetCloseTarget(obj.id)
	if err != nil {
		err = obj.close()
	}
	obj.webSock.Close(nil)
	return err
}
func (obj *Page) close() error {
	resp, err := obj.globalReqCli.Request(context.TODO(), "get", fmt.Sprintf("http://%s:%d/json/close/%s", obj.host, obj.port, obj.id), requests.RequestOption{DisProxy: true})
	if err != nil {
		return err
	}
	if resp.Text() == "Target is closing" {
		return nil
	}
	return errors.New(resp.Text())
}

func (obj *Page) Done() <-chan struct{} {
	return obj.webSock.Done()
}
func (obj *Page) Request(ctx context.Context, RequestFunc func(context.Context, *cdp.Route)) error {
	if RequestFunc == nil {
		if obj.dataCache {
			obj.webSock.RequestFunc = defaultRequestFunc
		} else {
			obj.webSock.RequestFunc = nil
		}
	} else {
		obj.webSock.RequestFunc = RequestFunc
	}
	var err error
	if obj.webSock.RequestFunc != nil {
		_, err = obj.webSock.FetchRequestEnable(ctx)
	} else if obj.webSock.ResponseFunc == nil {
		_, err = obj.webSock.FetchDisable(ctx)
	}
	return err
}
func (obj *Page) Response(ctx context.Context, ResponseFunc func(context.Context, *cdp.Route)) error {
	obj.webSock.ResponseFunc = ResponseFunc
	var err error
	if obj.webSock.ResponseFunc != nil {
		_, err = obj.webSock.FetchResponseEnable(ctx)
	} else if obj.webSock.RequestFunc == nil {
		_, err = obj.webSock.FetchDisable(ctx)
	}
	return err
}
func (obj *Page) initNodeId(ctx context.Context) (error, bool) {
	rs, err := obj.webSock.DOMGetDocument(ctx)
	if err != nil {
		return err, false
	}
	jsonData := tools.Any2json(rs.Result["root"])
	href := jsonData.Get("baseURL").String()
	if href != "" {
		obj.baseUrl = href
	}
	nodeId := jsonData.Get("nodeId").Int()
	ok := obj.nodeId == nodeId
	obj.nodeId = nodeId
	return nil, ok
}
func (obj *Page) Html(ctx context.Context, contents ...string) (*bs4.Client, error) {
	err, _ := obj.initNodeId(ctx)
	if err != nil {
		return nil, err
	}
	if len(contents) > 0 {
		return nil, obj.setHtml(ctx, contents[0])
	}
	return obj.html(ctx)
}
func (obj *Page) setHtml(ctx context.Context, content string) error {
	_, err := obj.webSock.DOMSetOuterHTML(ctx, obj.nodeId, content)
	return err
}
func (obj *Page) html(ctx context.Context) (*bs4.Client, error) {
	rs, err := obj.webSock.DOMGetOuterHTML(ctx, obj.nodeId)
	if err != nil {
		return nil, err
	}
	html := bs4.NewClient(rs.Result["outerHTML"].(string), obj.baseUrl)
	iframes := []*bs4.Client{}
	for _, iframe := range html.Finds("iframe") {
		if !strings.HasPrefix(iframe.Get("src"), "javascript:") {
			iframes = append(iframes, iframe)
		}
	}
	if len(iframes) > 0 {
		pageFrams, err := obj.QuerySelectorAll(ctx, "iframe")
		if err != nil {
			return nil, err
		}
		if len(iframes) != len(pageFrams) {
			return nil, errors.New("iframe error")
		}
		for i, ifram := range iframes {
			dh, err := pageFrams[i].Html(ctx)
			if err != nil {
				return nil, err
			}
			ifram.Html(dh.Html())
		}
	}
	return html, nil
}
func (obj *Page) WaitSelector(preCtx context.Context, selector string, timeouts ...int64) (*Dom, error) {
	var timeout int64 = 30
	if len(timeouts) > 0 {
		timeout = timeouts[0]
	}
	startTime := time.Now().Unix()
	for time.Now().Unix()-startTime <= timeout {
		dom, err := obj.QuerySelector(preCtx, selector)
		if err != nil {
			return nil, err
		}
		if dom != nil {
			return dom, nil
		}
		time.Sleep(time.Millisecond * 500)
	}
	return nil, errors.New("超时")
}
func (obj *Page) QuerySelector(ctx context.Context, selector string) (*Dom, error) {
	dom, err := obj.querySelector(ctx, selector)
	if err != nil {
		return dom, err
	}
	if dom == nil && selector != "iframe" {
		iframes, err := obj.querySelectorAll(ctx, "iframe")
		if err != nil {
			return nil, err
		}
		for _, iframe := range iframes {
			dom, err = iframe.querySelector(ctx, selector)
			if err != nil || dom != nil {
				return dom, err
			}
		}
	}
	return dom, err
}
func (obj *Page) querySelector(ctx context.Context, selector string) (*Dom, error) {
	err, _ := obj.initNodeId(ctx)
	if err != nil {
		return nil, err
	}
	rs, err := obj.webSock.DOMQuerySelector(ctx, obj.nodeId, selector)
	if err != nil {
		err2, ok := obj.initNodeId(ctx)
		if err2 != nil {
			return nil, err2
		}
		if ok {
			return nil, err
		} else {
			return obj.querySelector(ctx, selector)
		}
	}
	nodeId := int64(rs.Result["nodeId"].(float64))
	if nodeId == 0 {
		return nil, nil
	}
	dom := &Dom{
		baseUrl: obj.baseUrl,
		webSock: obj.webSock,
		nodeId:  nodeId,
	}
	if re.Search(`^iframe\W|\Wiframe\W|\Wiframe$|^iframe$`, selector) != nil {
		if err = dom.dom2Iframe(ctx); err != nil {
			return nil, err
		}
	}
	return dom, nil
}
func (obj *Page) QuerySelectorAll(ctx context.Context, selector string) ([]*Dom, error) {
	dom, err := obj.querySelectorAll(ctx, selector)
	if err != nil {
		return dom, err
	}
	if dom == nil && selector != "iframe" {
		iframes, err := obj.querySelectorAll(ctx, "iframe")
		if err != nil {
			return nil, err
		}
		doms := []*Dom{}
		for _, iframe := range iframes {
			dom, err = iframe.querySelectorAll(ctx, selector)
			if err != nil {
				return dom, err
			}
			doms = append(doms, dom...)
		}
		return doms, err
	}
	return dom, err
}
func (obj *Page) querySelectorAll(ctx context.Context, selector string) ([]*Dom, error) {
	err, _ := obj.initNodeId(ctx)
	if err != nil {
		return nil, err
	}
	rs, err := obj.webSock.DOMQuerySelectorAll(ctx, obj.nodeId, selector)
	if err != nil {
		err2, ok := obj.initNodeId(ctx)
		if err2 != nil {
			return nil, err2
		}
		if ok {
			return nil, err
		} else {
			return obj.querySelectorAll(ctx, selector)
		}
	}
	doms := []*Dom{}
	for _, nodeId := range tools.Any2json(rs.Result["nodeIds"]).Array() {
		dom := &Dom{
			baseUrl: obj.baseUrl,
			webSock: obj.webSock,
			nodeId:  nodeId.Int(),
		}
		if re.Search(`^iframe\W|\Wiframe\W|\Wiframe$|^iframe$`, selector) != nil {
			if err = dom.dom2Iframe(ctx); err != nil {
				return nil, err
			}
		}
		doms = append(doms, dom)
	}
	return doms, nil
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
func (obj *Page) TouchUp(ctx context.Context) error {
	_, err := obj.webSock.InputDispatchTouchEvent(ctx,
		"touchEnd",
		[]cdp.Point{})
	return err
}

// 移动操作结束
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

func (obj *Page) SetCookies(ctx context.Context, cookies ...cdp.Cookie) error {
	if len(cookies) == 0 {
		return nil
	}
	var err error
	for i := 0; i < len(cookies); i++ {
		if cookies[i].Domain == "" {
			if cookies[i].Url == "" {
				cookies[i].Url = obj.baseUrl
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
func (obj *Page) GetCookies(ctx context.Context, urls ...string) ([]cdp.Cookie, error) {
	if len(urls) == 0 {
		urls = append(urls, obj.baseUrl)
	}
	rs, err := obj.webSock.NetworkGetCookies(ctx, urls...)
	result := []cdp.Cookie{}
	if err != nil {
		return result, err
	}
	jsonData := tools.Any2json(rs.Result)
	for _, cookie := range jsonData.Get("cookies").Array() {
		var cook cdp.Cookie
		if err = json.Unmarshal(tools.StringToBytes(cookie.Raw), &cook); err != nil {
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
