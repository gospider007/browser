package cdp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"gitee.com/baixudong/gospider/re"
	"gitee.com/baixudong/gospider/requests"
	"gitee.com/baixudong/gospider/tools"
	"golang.org/x/exp/maps"
)

type RequestOption struct {
	Url      string      `json:"url"`
	Method   string      `json:"method"`
	PostData string      `json:"postData"`
	Headers  http.Header `json:"headers"`
}
type RequestData struct {
	Url              string            `json:"url"`
	UrlFragment      string            `json:"urlFragment"`
	Method           string            `json:"method"`
	Headers          map[string]string `json:"headers"`
	PostData         string            `json:"postData"`
	HasPostData      bool              `json:"hasPostData"`
	MixedContentType string            `json:"mixedContentType"`
	InitialPriority  string            `json:"initialPriority"` //初始优先级
	ReferrerPolicy   string            `json:"referrerPolicy"`
	IsLinkPreload    bool              `json:"isLinkPreload"`   //是否通过链路预加载加载。
	PostDataEntries  []DataEntrie      `json:"postDataEntries"` //是否通过链路预加载加载。
}
type RouteData struct {
	RequestId    string      `json:"requestId"`
	Request      RequestData `json:"request"`
	FrameId      string      `json:"frameId"`
	NetworkId    string      `json:"networkId"`
	ResourceType string      `json:"resourceType"`

	ResponseErrorReason string   `json:"responseErrorReason"`
	ResponseStatusCode  int      `json:"responseStatusCode"`
	ResponseStatusText  string   `json:"responseStatusText"`
	ResponseHeaders     []Header `json:"responseHeaders"`
}

type Route struct {
	webSock  *WebSock
	recvData RouteData
	isRoute  bool
}

func (obj *Route) IsResponse() bool {
	if obj.recvData.ResponseErrorReason != "" ||
		obj.recvData.ResponseStatusCode != 0 || obj.recvData.ResponseStatusText != "" ||
		obj.recvData.ResponseHeaders != nil {
		return true
	}
	return false
}
func (obj *Route) Error() error {
	if obj.recvData.ResponseErrorReason != "" {
		return errors.New(obj.recvData.ResponseErrorReason)
	}
	return nil
}
func (obj *Route) StatusCode() int {
	return obj.recvData.ResponseStatusCode
}
func (obj *Route) StatusText() string {
	return obj.recvData.ResponseStatusText
}
func (obj *Route) ResponseHeaders() http.Header {
	head := http.Header{}
	for _, hd := range obj.recvData.ResponseHeaders {
		head.Add(hd.Name, hd.Value)
	}
	return head
}

func (obj *Route) NewRequestOption() RequestOption {
	return RequestOption{
		Url:      obj.Url(),
		Method:   obj.Method(),
		PostData: obj.PostData(),
		Headers:  obj.Headers(),
	}
}
func (obj *Route) NewFulData(ctx context.Context) (fulData FulData, err error) {
	if !obj.IsResponse() {
		err = errors.New("not response route")
		return
	}
	if err = obj.Error(); err != nil {
		return
	}
	fulData.Body, err = obj.ResponseBody(ctx)
	fulData.StatusCode = obj.StatusCode()
	fulData.Headers = obj.ResponseHeaders()
	fulData.ResponsePhrase = obj.StatusText()
	return
}

// Document, Stylesheet, Image, Media, Font, Script, TextTrack, XHR, Fetch, Prefetch, EventSource, WebSocket, Manifest, SignedExchange, Ping, CSPViolationReport, Preflight, Other
func (obj *Route) ResourceType() string {
	return obj.recvData.ResourceType
}
func (obj *Route) Url() string {
	return obj.recvData.Request.Url
}
func (obj *Route) Method() string {
	return obj.recvData.Request.Method
}
func (obj *Route) PostData() string {
	return obj.recvData.Request.PostData
}
func (obj *Route) Headers() http.Header {
	if _, ok := obj.recvData.Request.Headers["If-Modified-Since"]; ok {
		delete(obj.recvData.Request.Headers, "If-Modified-Since")
	}
	head := http.Header{}
	for kk, vv := range obj.recvData.Request.Headers {
		head.Add(kk, vv)
	}
	return head
}
func (obj *Route) Cookies() (requests.Cookies, error) {
	return requests.ReadCookies(obj.Headers())
}
func keyMd5(key RequestOption, resourceType string) [16]byte {
	var md5Str string
	nt := strconv.Itoa(int(time.Now().Unix() / 1000))
	key.Url = re.Sub(fmt.Sprintf(`=%s\d*?&`, nt), "=&", key.Url)
	key.Url = re.Sub(fmt.Sprintf(`=%s\d*?$`, nt), "=", key.Url)

	key.Url = re.Sub(fmt.Sprintf(`=%s\d*?\.\d+?&`, nt), "=&", key.Url)
	key.Url = re.Sub(fmt.Sprintf(`=%s\d*?\.\d+?$`, nt), "=", key.Url)

	key.Url = re.Sub(`=0\.\d{10,}&`, "=&", key.Url)
	key.Url = re.Sub(`=0\.\d{10,}$`, "=", key.Url)
	md5Str += fmt.Sprintf("%s,%s,%s", key.Method, key.Url, key.PostData)

	switch resourceType {
	case "Document", "XHR", "Fetch", "Other":
		kks := maps.Keys(key.Headers)
		sort.Strings(kks)
		for _, k := range kks {
			md5Str += fmt.Sprintf("%s,%s", k, key.Headers[k])
		}
	}
	return tools.Md5(md5Str)
}
func (obj *Route) Request(ctx context.Context, routeOption RequestOption, options ...requests.RequestOption) (FulData, error) {
	option := requests.RequestOption{
		Proxy: obj.webSock.option.Proxy,
	}
	if len(options) > 0 {
		option = options[0]
	}
	if routeOption.PostData != "" {
		option.Raw = routeOption.PostData
	}
	option.Headers = routeOption.Headers
	resourceType := obj.ResourceType()
	switch resourceType {
	case "Document", "XHR", "Script", "Fetch", "Other":
		option.TryNum = 2
	default:
		option.TryNum = 1
	}
	var fulData FulData
	var err error
	routeKey := keyMd5(routeOption, resourceType)
	if obj.webSock.option.DataCache {
		if fulData, err = obj.webSock.db.Get(routeKey); err == nil { //如果有緩存
			return fulData, err
		}
	}
	rs, err := obj.webSock.reqCli.Request(ctx, routeOption.Method, routeOption.Url, option)
	if err != nil {
		return fulData, err
	}
	fulData.StatusCode = rs.StatusCode()
	fulData.Body = rs.Text()
	fulData.Headers = rs.Headers()
	fulData.ResponsePhrase = rs.Status()
	if obj.webSock.option.DataCache && fulData.StatusCode == 200 && fulData.Body != "" && routeOption.Method == "GET" {
		obj.webSock.db.Put(routeKey, fulData)
	}
	return fulData, nil
}
func (obj *Route) FulFill(ctx context.Context, fulDatas ...FulData) error {
	obj.isRoute = true
	var fulData FulData
	if len(fulDatas) > 0 {
		fulData = fulDatas[0]
	}
	_, err := obj.webSock.FetchFulfillRequest(ctx, obj.recvData.RequestId, fulData)
	if err != nil {
		obj.Fail(nil)
	}
	return err
}
func (obj *Route) RequestContinue(ctx context.Context) (FulData, error) {
	obj.isRoute = true
	fulData, err := obj.Request(ctx, obj.NewRequestOption())
	if err != nil {
		obj.Fail(ctx)
	} else {
		err = obj.FulFill(ctx, fulData)
	}
	return fulData, err
}

func (obj *Route) Continue(ctx context.Context, options ...RequestOption) error {
	obj.isRoute = true
	_, err := obj.webSock.FetchContinueRequest(ctx, obj.recvData.RequestId, options...)
	if err != nil {
		obj.Fail(ctx)
	}
	return err
}
func (obj *Route) ResponseBody(ctx context.Context) (string, error) {
	if err := obj.Error(); err != nil {
		obj.Continue(ctx)
		return "", err
	}
	rs, err := obj.webSock.FetchGetResponseBody(ctx, obj.recvData.RequestId)
	if err != nil {
		return "", err
	}
	jsonData := tools.Any2json(rs.Result)
	body := jsonData.Get("body").String()
	if body == "" {
		return body, nil
	}
	if jsonData.Get("base64Encoded").Bool() {
		bodyByte, err := tools.Base64Decode(body)
		if err != nil {
			return body, err
		}
		body = tools.BytesToString(bodyByte)
	}
	return body, nil
}

// Failed, Aborted, TimedOut, AccessDenied, ConnectionClosed, ConnectionReset, ConnectionRefused, ConnectionAborted, ConnectionFailed, NameNotResolved, InternetDisconnected, AddressUnreachable, BlockedByClient, BlockedByResponse
func (obj *Route) Fail(ctx context.Context, errorReasons ...string) error {
	obj.isRoute = true
	var errorReason string
	if len(errorReasons) > 0 {
		errorReason = errorReasons[0]
	}
	if errorReason == "" {
		errorReason = "Failed"
	}
	_, err := obj.webSock.FetchFailRequest(ctx, obj.recvData.RequestId, errorReason)
	return err
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type FulData struct {
	StatusCode     int         `json:"responseCode"`
	Headers        http.Header `json:"responseHeaders"`
	Body           string      `json:"body"`
	ResponsePhrase string      `json:"responsePhrase"`
}

func (obj FulData) Cookies() (requests.Cookies, error) {
	return requests.ReadSetCookies(obj.Headers)
}
