package browser

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math"
	uurl "net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gospider007/bs4"
	"github.com/gospider007/cdp"
	"github.com/gospider007/gson"
	"github.com/gospider007/requests"
	"github.com/gospider007/tools"
)

type Page struct {
	browserContext *BrowserContext
	option         *PageOption
	targetId       string
	targetType     string
	mouseX         float64
	mouseY         float64
	ctx            context.Context
	cnl            context.CancelFunc
	globalReqCli   *requests.Client

	baseUrl string
	webSock *cdp.WebSock

	pageStop           bool
	domLoad            bool
	loadNotices        chan struct{}
	stopNotices        chan struct{}
	networkNotices     chan struct{}
	networkNoticesSize atomic.Int64
	frames             sync.Map
	storageEnable      bool
}

func defaultRequestFunc(ctx context.Context, r *cdp.Route) {
	r.Continue(ctx)
}
func (obj *Page) delStopNotice() {
	obj.pageStop = false
	obj.delLoadNotice()
	select {
	case <-obj.stopNotices:
	default:
	}
}
func (obj *Page) delLoadNotice() {
	obj.domLoad = false
	select {
	case <-obj.loadNotices:
	default:
	}
}
func (obj *Page) addStopNotice() {
	obj.pageStop = true
	obj.addLoadNotice()
	select {
	case obj.stopNotices <- struct{}{}:
	default:
	}
}
func (obj *Page) addLoadNotice() {
	obj.domLoad = true
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
	if frameId := rd.Params["frameId"].(string); frameId == obj.targetId {
		obj.clearFrames()
		obj.delStopNotice()
	}
}
func (obj *Page) pageEndLoadMain(ctx context.Context, rd cdp.RecvData) {
	if frameId := rd.Params["frameId"].(string); frameId == obj.targetId {
		obj.addStopNotice()
	}
}
func (obj *Page) frameNavigated(ctx context.Context, rd cdp.RecvData) {
	jsonData, _ := gson.Decode(rd.Params)
	targetId := jsonData.Get("frame.id").String()
	if targetId == "" {
		return
	}
	if targetId == obj.targetId {
		obj.baseUrl = jsonData.Get("frame.url").String()
	} else if jsonData.Get("frame.parentId").String() == obj.targetId {
		obj.addIframe("frameNavigated", targetId)
	}
}
func (obj *Page) Url() string {
	return obj.baseUrl
}
func (obj *Page) domLoadMain(ctx context.Context, rd cdp.RecvData) {
	obj.addLoadNotice()
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
		if strings.HasSuffix(route.Url(), "/favicon.ico") {
			route.FulFill(ctx, cdp.FulData{
				StatusCode: 404,
			})
		} else if obj.option.requestFunc != nil {
			obj.option.requestFunc(ctx, route)
		} else {
			defaultRequestFunc(ctx, route)
		}
		if !route.Used() {
			route.Fail(ctx)
		}
	}
}
func (obj *Page) iframeToTargetMain(ctx context.Context, rd cdp.RecvData) {
	jsonData, err := gson.Decode(rd.Params)
	if err != nil {
		return
	}
	if jsonData.Get("waitingForDebugger").Bool() {
		targetId := jsonData.Get("targetInfo.targetId").String()
		sessionId := jsonData.Get("sessionId").String()
		if targetId != "" && sessionId != "" {
			obj.addIframe("iframeToTargetMain", targetId)
			obj.webSock.Cdp(obj.ctx, sessionId, "Runtime.runIfWaitingForDebugger")
		}
	}
}
func (obj *Page) targetToTargetMain(ctx context.Context, rd cdp.RecvData) {
	jsonData, err := gson.Decode(rd.Params)
	if err != nil {
		return
	}
	targetId := jsonData.Get("targetInfo.targetId").String()
	browserContextId := jsonData.Get("targetInfo.browserContextId").String()
	targetType := jsonData.Get("targetInfo.type").String()
	// if obj.browserContext.browserContextId == browserContextId {
	// 	log.Print(jsonData, obj.browserContext.browserContextId, "  ===. ", browserContextId, "  ==. ", obj.targetId)
	// }
	if targetId != "" && browserContextId != "" && targetId != obj.targetId && targetType == "page" {
		obj.addIframe("targetToTargetMain", targetId)
	}
}

func (obj *Page) GetFrame(frameId string) (*Page, bool) {
	frame, ok := obj.frames.Load(frameId)
	if !ok {
		return nil, false
	}
	return frame.(*Page), true
}
func (obj *Page) clearFrames() {
	obj.frames.Range(func(key, value any) bool {
		obj.frames.Delete(key)
		return true
	})
}
func (obj *Page) addIframe(msg string, targetId string) error {
	_, ok := obj.frames.Load(targetId)
	if ok {
		return nil
	}
	// log.Print(msg, " ===. ", targetId)
	iframe, err := newPageWithTargetId(obj.browserContext, obj.ctx, obj.ctx, targetId, obj.option)
	if err != nil {
		return err
	}
	obj.frames.Store(targetId, iframe)
	return nil
}

func (obj *Page) addEvent(method string, fun func(ctx context.Context, rd cdp.RecvData)) {
	obj.webSock.AddEvent(method, fun)
}

//go:embed stealthRaw.js
var stealthRaw string

func (obj *Page) init(ctx context.Context) error {
	var err error
	if obj.webSock, err = cdp.NewWebSock(
		obj.ctx,
		obj.globalReqCli,
		fmt.Sprintf("ws://%s/devtools/page/%s", obj.browserContext.Addr(), obj.targetId),
	); err != nil {
		return err
	}
	obj.addEvent("Page.frameStartedLoading", obj.pageStartLoadMain)
	obj.addEvent("Page.frameStoppedLoading", obj.pageEndLoadMain)
	obj.addEvent("Page.domContentEventFired", obj.domLoadMain)
	obj.addEvent("Page.frameNavigated", obj.frameNavigated)
	obj.addEvent("Fetch.requestPaused", obj.routeMain)
	obj.addEvent("Target.attachedToTarget", obj.iframeToTargetMain)
	if _, err = obj.webSock.PageEnable(ctx); err != nil {
		return err
	}
	if _, err = obj.webSock.FetchRequestEnable(ctx); err != nil {
		return err
	}
	if _, err = obj.webSock.TargetSetAutoAttach(ctx); err != nil {
		return err
	}
	if obj.option.Stealth {
		if err = obj.AddScript(ctx, stealthRaw); err != nil {
			return err
		}
	}
	return nil
}

type FpOption struct {
	Browser         string //("chrome" | "firefox" | "safari" | "edge")
	Device          string //"mobile" | "desktop"
	OperatingSystem string //"windows" | "macos" | "linux" | "android" | "ios"
	UserAgent       string
	Locales         []string
	Locale          string
}

func (obj *Page) AddScript(ctx context.Context, script string) error {
	_, err := obj.webSock.PageAddScriptToEvaluateOnNewDocument(ctx, script)
	return err
}
func (obj *Page) Screenshot(ctx context.Context, rect cdp.Rect, options ...cdp.ScreenshotOption) ([]byte, error) {
	scrollTop, err := obj.Eval(ctx, `()=>{return document.documentElement.scrollTop}`, nil)
	if err != nil {
		return nil, err
	}
	rect.Y += scrollTop.Float()
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
func (obj *Page) WaitDomLoad(ctx context.Context, msTimes ...time.Duration) (err error) {
	var msTime time.Duration
	if len(msTimes) > 0 {
		msTime = msTimes[0]
	}
	return obj.waitMain(ctx, msTime, obj.loadNotices, func() bool { return obj.domLoad })
}
func (obj *Page) WaitDomLoadWithTimeout(preCtx context.Context, timeout time.Duration, msTimes ...time.Duration) (err error) {
	var ctx context.Context
	var cnl context.CancelFunc
	if preCtx == nil {
		preCtx = obj.ctx
	}
	var msTime time.Duration
	if len(msTimes) > 0 {
		msTime = msTimes[0]
	}
	ctx, cnl = context.WithTimeout(preCtx, timeout)
	defer cnl()
	return obj.waitMain(ctx, msTime, obj.loadNotices, func() bool { return obj.domLoad })
}
func (obj *Page) WaitPageStop(ctx context.Context, msTimes ...time.Duration) (err error) {
	var msTime time.Duration
	if len(msTimes) > 0 {
		msTime = msTimes[0]
	}
	return obj.waitMain(ctx, msTime, obj.stopNotices, func() bool { return obj.pageStop })
}
func (obj *Page) WaitPageStopWithTimeout(preCtx context.Context, timeout time.Duration, msTimes ...time.Duration) (err error) {
	var ctx context.Context
	var cnl context.CancelFunc
	if preCtx == nil {
		preCtx = obj.ctx
	}
	var msTime time.Duration
	if len(msTimes) > 0 {
		msTime = msTimes[0]
	}
	ctx, cnl = context.WithTimeout(preCtx, timeout)
	defer cnl()
	return obj.waitMain(ctx, msTime, obj.stopNotices, func() bool { return obj.pageStop })
}

func (obj *Page) WaitNetwork(ctx context.Context, msTimes ...time.Duration) error {
	var msTime time.Duration
	if len(msTimes) > 0 {
		msTime = msTimes[0]
	}
	return obj.waitMain(ctx, msTime, obj.networkNotices, func() bool { return obj.pageStop && obj.networkNoticesSize.Load() <= 0 })
}
func (obj *Page) WaitNetworkWithTimeout(preCtx context.Context, timeout time.Duration, msTimes ...time.Duration) error {
	var ctx context.Context
	var cnl context.CancelFunc
	if preCtx == nil {
		preCtx = obj.ctx
	}
	var msTime time.Duration
	if len(msTimes) > 0 {
		msTime = msTimes[0]
	}
	ctx, cnl = context.WithTimeout(preCtx, timeout)
	defer cnl()
	return obj.waitMain(ctx, msTime, obj.networkNotices, func() bool { return obj.pageStop && obj.networkNoticesSize.Load() <= 0 })
}

func (obj *Page) waitMain(ctx context.Context, msTime time.Duration, notices <-chan struct{}, okFunc func() bool) error {
	if msTime == 0 {
		msTime = time.Millisecond * 1200
	}
	if ctx == nil {
		ctx = obj.ctx
	}
	var ok bool
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-obj.ctx.Done():
			return context.Cause(obj.ctx)
		case <-notices:
			if okFunc() {
				ok = true
			} else {
				ok = false
			}
		case <-time.After(msTime):
			if !okFunc() {
				ok = false
				continue
			}
			if ok {
				return nil
			}
			ok = true
		}
	}
}

func (obj *Page) GoTo(preCtx context.Context, url string, options ...cdp.PageNavigateOption) error {
	obj.baseUrl = url
	_, err := obj.webSock.PageNavigate(preCtx, url, options...)
	return err
}

// ex:   ()=>{}  或者  (params)=>{}
func (obj *Page) Eval(ctx context.Context, expression string, params ...map[string]any) (*gson.Client, error) {
	var value string
	if len(params) > 0 {
		con, err := gson.Encode(params[0])
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
	result, err := gson.Decode(rs.Result)
	if err != nil {
		return nil, err
	}
	if exceptionDetails := result.Get("exceptionDetails"); exceptionDetails.Exists() {
		if exception := exceptionDetails.Get("exception.description"); exception.Exists() {
			return nil, errors.New(exception.String())
		}
		if exception := exceptionDetails.Get("text"); exception.Exists() {
			return nil, errors.New(exception.String())
		}
		return nil, errors.New(exceptionDetails.String())
	}
	if value := result.Get("result.value"); !value.Exists() {
		if result.Get("result.type").String() == "undefined" {
			return nil, nil
		}
		return nil, errors.New(result.String())
	} else {
		return value, nil
	}
}
func (obj *Page) Close() (err error) {
	obj.clearFrames()
	obj.webSock.TargetCloseTarget(obj.targetId)
	err = obj.close()
	obj.webSock.CloseWithError(errors.New("page closed"))
	if obj.globalReqCli != nil {
		obj.globalReqCli.Close()
	}
	obj.cnl()
	return
}
func (obj *Page) Cdp(ctx context.Context, method string, params ...map[string]any) (cdp.RecvData, error) {
	return obj.webSock.Cdp(ctx, "", method, params...)
}
func (obj *Page) close() error {
	_, err := obj.globalReqCli.Request(context.TODO(), "get", fmt.Sprintf("http://%s/json/close/%s", obj.browserContext.Addr(), obj.targetId), requests.RequestOption{
		DisProxy:   true,
		MaxRetries: 10,
		ResultCallBack: func(ctx *requests.Response) error {
			switch ctx.Text() {
			case "Target is closing", fmt.Sprintf("No such target id: %s", obj.targetId):
			}
			if text := ctx.Text(); text == "Target is closing" || text == fmt.Sprintf("No such target id: %s", obj.targetId) {
				return nil
			}
			return errors.New("req close target error")
		},
	})
	return err
}

func (obj *Page) Context() context.Context {
	return obj.webSock.Context()
}
func (obj *Page) Request(ctx context.Context, requestFunc func(context.Context, *cdp.Route)) error {
	obj.option.requestFunc = requestFunc
	return nil
}

func (obj *Page) Html(ctx context.Context) (*bs4.Client, error) {
	r, err := obj.webSock.DOMGetDocuments(ctx)
	if err != nil {
		return nil, err
	}
	data, err := gson.Decode(r.Result)
	if err != nil {
		return nil, err
	}
	parseDom, err := obj.parseJsonDom(ctx, data.Get("root"))
	if err != nil {
		return nil, err
	}
	mainHtml := bs4.NewClientWithNode(parseDom)
	for _, iframe := range mainHtml.Finds("iframe") {
		if gospiderFrameId := iframe.Get("gospiderFrameId"); gospiderFrameId != "" {
			if framePage, ok := obj.GetFrame(gospiderFrameId); ok {
				if frameHtml, err := framePage.Html(ctx); err == nil {
					iframe.SetHtml(frameHtml.String())
				}
			}
		}
	}
	return mainHtml, nil
}
func (obj *Page) GetFrameTree(ctx context.Context) (*bs4.Client, error) {
	r, err := obj.webSock.PageGetFrameTree(ctx)
	if err != nil {
		return nil, err
	}
	data, err := gson.Decode(r.Result)
	if err != nil {
		return nil, err
	}
	// log.Print(data)
	// return nil, nil
	parseDom, err := obj.parseJsonDom(ctx, data.Get("root"))
	if err != nil {
		return nil, err
	}
	mainHtml := bs4.NewClientWithNode(parseDom)
	for _, iframe := range mainHtml.Finds("iframe") {
		if gospiderFrameId := iframe.Get("gospiderFrameId"); gospiderFrameId != "" {
			if framePage, ok := obj.GetFrame(gospiderFrameId); ok {
				if frameHtml, err := framePage.Html(ctx); err == nil {
					iframe.SetHtml(frameHtml.String())
				}
			}
		}
	}
	return mainHtml, nil
}

func (obj *Page) mainHtml(ctx context.Context) (*bs4.Client, error) {
	r, err := obj.webSock.DOMGetDocuments(ctx)
	if err != nil {
		return nil, err
	}
	data, err := gson.Decode(r.Result)
	if err != nil {
		return nil, err
	}
	// log.Print(data)
	parseDom, err := obj.parseJsonDom(ctx, data.Get("root"))
	if err != nil {
		return nil, err
	}
	return bs4.NewClientWithNode(parseDom), nil
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
			return nil, context.Cause(ctx)
		}
	}
	return nil, errors.New("超时")
}
func (obj *Page) WaitSelectors(ctx context.Context, selector string, timeouts ...time.Duration) ([]*Dom, error) {
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
		dom, err := obj.QuerySelectorAll(ctx, selector)
		if err != nil {
			return nil, err
		}
		if len(dom) > 0 {
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
			return nil, context.Cause(ctx)
		}
	}
	return nil, errors.New("超时")
}
func (obj *Page) Frames() []*Page {
	frames := []*Page{}
	obj.frames.Range(func(key, value any) bool {
		frames = append(frames, value.(*Page))
		return true
	})
	return frames
}
func (obj *Page) QuerySelector(ctx context.Context, selector string) (*Dom, error) {
	html, err := obj.mainHtml(ctx)
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
		frameId: ele.Get("gospiderFrameId"),
		ele:     ele,
	}
	return dom, err
}
func (obj *Page) QuerySelectorAll(ctx context.Context, selector string) ([]*Dom, error) {
	html, err := obj.mainHtml(ctx)
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
			frameId: ele.Get("gospiderFrameId"),
			ele:     ele,
		}
		doms = append(doms, dom)
	}
	return doms, err
}

func (obj *Page) Move(ctx context.Context, x, y float64, steps ...int) error {
	return obj.MoveTo(ctx, cdp.Point{X: obj.mouseX + x, Y: obj.mouseY + y}, steps...)
}
func randomPointsWithStartAndMinDist(width, height float64, n int) []cdp.Point {
	midHight := height / 2
	midWidth := width / 2
	points := make([]cdp.Point, 0, n*2)
	for range n {
		point1 := cdp.Point{
			X: tools.RanFloat64(0, int64(midWidth)),
			Y: tools.RanFloat64(int64(midHight), int64(height)),
		}
		point2 := cdp.Point{
			X: tools.RanFloat64(int64(midWidth), int64(width)),
			Y: tools.RanFloat64(0, int64(midHight)),
		}
		points = append(points, point1, point2)
	}
	return points
}
func (obj *Page) RandomMove(ctx context.Context) error {
	layout, err := obj.webSock.PageGetLayoutMetrics(ctx)
	if err != nil {
		return err
	}
	jsonData, err := gson.Decode(layout.Result)
	if err != nil {
		return err
	}
	clientWidth := jsonData.Get("visualViewport.clientWidth").Float()
	clientHeight := jsonData.Get("visualViewport.clientHeight").Float()
	if clientWidth == 0 || clientHeight == 0 {
		return errors.New("not found width or height")
	}
	for _, point := range randomPointsWithStartAndMinDist(clientWidth, clientHeight, 10) {
		err = obj.MoveTo(ctx, point)
		if err != nil {
			return err
		}
	}
	return nil
}
func (obj *Page) MoveTo(ctx context.Context, point cdp.Point, steps ...int) error {
	step := int(math.Max(point.X, point.Y)/10) + 1
	if len(steps) > 0 && steps[0] > 0 {
		step = steps[0]
	}
	if step == 1 {
		return obj.moveTo(ctx, point)
	}
	for _, poi := range tools.GetTrack(
		[2]float64{obj.mouseX, obj.mouseY},
		[2]float64{point.X, point.Y},
		step,
	) {
		err := obj.moveTo(ctx, cdp.Point{X: poi[0], Y: poi[1]})
		if err != nil {
			return err
		}
	}
	return nil
}
func (obj *Page) moveTo(ctx context.Context, point cdp.Point) error {
	_, err := obj.webSock.InputDispatchMouseEvent(ctx,
		cdp.DispatchMouseEventOption{
			Type:   "mouseMoved",
			Button: "left",
			X:      point.X,
			Y:      point.Y,
		})
	if err != nil {
		return err
	}
	obj.mouseX = point.X
	obj.mouseY = point.Y
	return nil
}
func (obj *Page) TouchMoveTo(ctx context.Context, point cdp.Point, steps ...int) error {
	step := int(math.Max(point.X, point.Y)/10) + 1
	if len(steps) > 0 && steps[0] > 0 {
		step = steps[0]
	}
	for _, poi := range tools.GetTrack(
		[2]float64{obj.mouseX, obj.mouseY},
		[2]float64{point.X, point.Y},
		step,
	) {
		_, err := obj.webSock.InputDispatchTouchEvent(ctx, "touchMove", []cdp.Point{
			{
				X: poi[0],
				Y: poi[1],
			},
		})
		if err != nil {
			return err
		}
		obj.mouseX = poi[0]
		obj.mouseY = poi[1]
	}
	return nil
}
func (obj *Page) TouchMove(ctx context.Context, x, y float64, steps ...int) error {
	return obj.TouchMoveTo(ctx, cdp.Point{X: obj.mouseX + x, Y: obj.mouseY + y}, steps...)
}

func (obj *Page) Wheel(ctx context.Context, x, y, deltaX, deltaY float64) error {
	_, err := obj.webSock.InputDispatchMouseEvent(ctx,
		cdp.DispatchMouseEventOption{
			Type:   "mouseWheel",
			X:      x,
			Y:      y,
			DeltaX: deltaX,
			DeltaY: deltaY,
		})
	return err
}

func (obj *Page) Down(ctx context.Context, point cdp.Point) error {
	err := obj.MoveTo(ctx, point)
	if err != nil {
		return err
	}
	_, err = obj.webSock.InputDispatchMouseEvent(ctx,
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
	return err
}
func (obj *Page) TouchDown(ctx context.Context, point cdp.Point) error {
	err := obj.TouchMoveTo(ctx, point)
	if err != nil {
		return err
	}
	_, err = obj.webSock.InputDispatchTouchEvent(ctx, "touchStart",
		[]cdp.Point{
			{
				X: point.X,
				Y: point.Y,
			},
		})
	if err != nil {
		return err
	}
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
	if err := obj.SetUserAgentOverride(ctx, device.UserAgent, ""); err != nil {
		return err
	}
	if err := obj.SetTouch(ctx, device.HasTouch); err != nil {
		return err
	}
	return obj.SetDeviceMetrics(ctx, device)
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
	for i := 0; i < len(cookies); i++ {
		if cookies[i].Url == "" {
			if obj.baseUrl == "" {
				return errors.New("not found base url")
			}
			uu, err := uurl.Parse(obj.baseUrl)
			if err != nil {
				return err
			}
			cookies[i].Url = fmt.Sprintf("%s://%s", uu.Scheme, uu.Host) + "/"
		}
		if cookies[i].Domain == "" {
			us, err := uurl.Parse(cookies[i].Url)
			if err != nil {
				return err
			}
			cookies[i].Domain = us.Hostname()
		}
	}
	_, err := obj.webSock.NetworkSetCookies(ctx, cookies)
	return err
}
func (obj *Page) SetCookiesWitString(ctx context.Context, href string, cookies string) error {
	for _, cookie := range strings.Split(cookies, "; ") {
		kvs := strings.Split(cookie, "=")
		if len(kvs) < 2 {
			continue
		}
		name := kvs[0]
		value := strings.Join(kvs[1:], "=")
		err := obj.SetCookies(ctx, cdp.Cookie{
			Name:  name,
			Value: value,
			Url:   href,
		})
		if err != nil {
			return err
		}
	}
	return nil
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
func (obj *Page) TargetId() string {
	return obj.targetId
}
func (obj *Page) Activate(ctx context.Context) error {
	_, err := obj.webSock.PageBringToFront(ctx)
	return err
}
func (obj *Page) SetHtml(ctx context.Context, html string) error {
	_, err := obj.webSock.PageSetDocumentContent(ctx, obj.targetId, html)
	return err
}

func (obj *Page) SetDOMStorageItem(ctx context.Context, key, val string, isLocalStorage bool) error {
	if obj.baseUrl == "" {
		return errors.New("not found base url")
	}
	uu, err := uurl.Parse(obj.baseUrl)
	if err != nil {
		return err
	}
	securityOrigin := fmt.Sprintf("%s://%s", uu.Scheme, uu.Host) + "/"
	if !obj.storageEnable {
		_, err = obj.webSock.StorageEnable(ctx, securityOrigin)
		if err != nil {
			return err
		}
		obj.storageEnable = true
	}
	_, err = obj.webSock.SetDOMStorageItem(ctx, securityOrigin, key, val, isLocalStorage)
	return err
}

func (obj *Page) SetExtraHTTPHeaders(ctx context.Context, headers map[string]string) error {
	_, err := obj.webSock.NetworkSetExtraHTTPHeaders(ctx, headers)
	return err
}

// 设置浏览器的地理位置
func (obj *Page) SetGeolocation(preCtx context.Context, latitude float64, longitude float64) error {
	_, err := obj.webSock.EmulationSetGeolocationOverride(preCtx, latitude, longitude)
	return err
}

// 设置浏览器的语言
func (obj *Page) SetLocaleOverride(preCtx context.Context, local string) error {
	_, err := obj.webSock.EmulationSetLocaleOverride(preCtx, local)
	return err
}

// 设置浏览器的语言
func (obj *Page) SetUserAgentOverride(preCtx context.Context, userAgent string, language string) error {
	if userAgent == "" {
		userAgent = tools.UserAgent
	}
	_, err := obj.webSock.EmulationSetUserAgentOverride(preCtx, userAgent, language)
	if err != nil {
		return err
	}
	_, err = obj.webSock.NetworkSetUserAgentOverride(preCtx, userAgent, language)
	if err != nil {
		return err
	}
	return err
}

// 设置浏览器的时区
func (obj *Page) SetTimezoneOverride(preCtx context.Context, timezoneId string) error {
	_, err := obj.webSock.EmulationSetTimezoneOverride(preCtx, timezoneId)
	return err
}
