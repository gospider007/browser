package browser

import (
	"archive/zip"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/baixudong/browser/cdp"
	"gitee.com/baixudong/gospider/cmd"
	"gitee.com/baixudong/gospider/conf"
	"gitee.com/baixudong/gospider/db"
	"gitee.com/baixudong/gospider/ja3"
	"gitee.com/baixudong/gospider/re"
	"gitee.com/baixudong/gospider/requests"
	"gitee.com/baixudong/gospider/tools"
)

//go:embed stealth.js
var stealth string

// //go:embed stealth2.js
// var stealth2 string

type Client struct {
	isReplaceRequest bool //是否自定义请求
	proxy            string
	getProxy         func(ctx context.Context, url *url.URL) (string, error)
	db               *db.Client[cdp.FulData]
	cmdCli           *cmd.Client
	globalReqCli     *requests.Client
	port             int
	host             string
	lock             sync.Mutex
	ctx              context.Context
	cnl              context.CancelFunc
	webSock          *cdp.WebSock
	dataCache        bool
	headless         bool
	stealth          bool //是否开启随机指纹
}
type ClientOption struct {
	ChromePath string   //chrome浏览器执行路径
	Host       string   //连接host
	Port       int      //连接port
	UserDir    string   //设置用户目录
	Args       []string //启动参数
	Headless   bool     //是否使用无头
	DataCache  bool     //开启数据缓存
	Ja3Spec    ja3.ClientHelloSpec
	Ja3        bool
	H2Ja3Spec  ja3.H2Ja3Spec
	H2Ja3      bool
	UserAgent  string
	Proxy      string                                                  //代理http,https,socks5,ex: http://127.0.0.1:7005
	GetProxy   func(ctx context.Context, url *url.URL) (string, error) //代理
	Width      int64                                                   //浏览器的宽
	Height     int64                                                   //浏览器的高
	Stealth    bool                                                    //是否开启随机指纹
}

type downClient struct {
	sync.Mutex
}

var oneDown = &downClient{}

var chromeVersion = 1132420

func verifyEvalPath(path string) error {
	if strings.HasSuffix(path, "chrome.exe") || strings.HasSuffix(path, "Chromium.app") || strings.HasSuffix(path, "chrome") || strings.HasSuffix(path, "chromium") {
		return nil
	}
	if strings.HasSuffix(path, "msedge.exe") || strings.HasSuffix(path, "msedge") {
		return nil
	}
	return errors.New("请输入正确的浏览器路径,如: c:/chrome.exe")
}
func (obj *downClient) getChromePath(preCtx context.Context) (string, error) {
	obj.Lock()
	defer obj.Unlock()
	chromDir, err := conf.GetMainDirPath()
	if err != nil {
		return "", err
	}
	chromDir = tools.PathJoin(chromDir, strconv.Itoa(chromeVersion))

	var chromePath string
	switch runtime.GOOS {
	case "windows":
		chromePath = tools.PathJoin(chromDir, "chrome-win", "chrome.exe")
	case "darwin":
		chromePath = tools.PathJoin(chromDir, "chrome-mac", "Chromium.app")
	case "linux":
		chromePath = tools.PathJoin(chromDir, "chrome-linux", "chrome")
	default:
		return "", errors.New("dont know goos")
	}
	if !tools.PathExist(chromePath) {
		if err = DownChrome(preCtx, chromeVersion); err != nil {
			return "", err
		}
		if !tools.PathExist(chromePath) {
			return "", errors.New("not found chrome")
		}
	}
	return chromePath, nil
}
func runChrome(ctx context.Context, option *ClientOption) (*cmd.Client, bool, error) {
	var err error
	var isReplaceRequest bool
	if option.Host == "" {
		option.Host = "127.0.0.1"
	}
	if option.Port == 0 {
		option.Port, err = tools.FreePort()
		if err != nil {
			return nil, isReplaceRequest, err
		}
	}
	if option.UserAgent == "" {
		option.UserAgent = requests.UserAgent
	}
	if option.ChromePath == "" {
		option.ChromePath, err = oneDown.getChromePath(ctx)
		if err != nil {
			return nil, isReplaceRequest, err
		}
	}
	if err = verifyEvalPath(option.ChromePath); err != nil {
		return nil, isReplaceRequest, err
	}
	var isDelDir bool
	if option.UserDir == "" {
		option.UserDir, err = conf.GetTempChromeDirPath()
		if err != nil {
			return nil, isReplaceRequest, err
		}
		isDelDir = true
	}
	args := []string{}
	args = append(args, chromeArgs...)
	if option.UserAgent != "" {
		args = append(args, fmt.Sprintf("--user-agent=%s", option.UserAgent))
	}
	if option.Headless {
		args = append(args, "--headless=new")
	}
	if option.Proxy != "" {
		proxyUrl, err := url.Parse(option.Proxy)
		if err != nil {
			return nil, isReplaceRequest, err
		}
		if proxyUrl.User == nil {
			args = append(args, fmt.Sprintf(`--proxy-server=%s`, proxyUrl.String()))
		} else {
			isReplaceRequest = true
		}
	}
	args = append(args, fmt.Sprintf(`--user-data-dir=%s`, option.UserDir))
	args = append(args, fmt.Sprintf("--remote-debugging-port=%d", option.Port))
	args = append(args, fmt.Sprintf("--window-size=%d,%d", option.Width, option.Height))

	args = append(args, option.Args...)
	var closeCallBack func()
	if isDelDir && option.UserDir != "" {
		closeCallBack = func() {
			for i := 0; i < 10; i++ {
				if os.RemoveAll(option.UserDir) == nil {
					return
				}
				time.Sleep(time.Millisecond * 300)
			}
		}
	}
	cli, err := cmd.NewClient(ctx, cmd.ClientOption{
		Name:          option.ChromePath,
		Args:          args,
		CloseCallBack: closeCallBack,
	})
	if err != nil {
		return cli, isReplaceRequest, err
	}
	go cli.Run()
	return cli, isReplaceRequest, cli.Err()
}

var chromeArgs = []string{
	// "--virtual-time-budget=1000", //缩短setTimeout  setInterval 的时间1000秒:目前不生效，不知道以后会不会生效，等生效了再打开
	//自动化选项禁用
	"--useAutomationExtension=false",                //禁用自动化扩展。
	"--excludeSwitches=enable-automation",           //禁用自动化
	"--disable-blink-features=AutomationControlled", //禁用 Blink 引擎的自动化控制。

	//稳定性选项
	"--no-sandbox",      //禁用 Chrome 的沙盒模式。
	"--set-uid-sandbox", //命令行参数用于设置 Chrome 进程运行时使用的 UID，从而提高 Chrome 浏览器的安全性
	"--set-gid-sandbox", //命令行参数用于设置 Chrome 进程运行时使用的 GID，从而提高 Chrome 浏览器的安全性
	"--incognito",       //隐私模式
	"--blink-settings=primaryHoverType=2,availableHoverTypes=2,primaryPointerType=4,availablePointerTypes=4,imagesEnabled=true", //Blink 设置。
	"--ignore-ssl-errors=true", //忽略 SSL 错误。
	"--disable-setuid-sandbox", //重要headless
	"--disable-extensions",     //禁用所有扩展程序，这可以降低Chrome对内存的占用。
	"--disable-plugins",        //禁用所有已安装的Chrome浏览器插件。

	"--process-per-site",              //为每个站点启动一个新的进程，这可以防止内存泄漏，并降低同一进程中多个标签页的内存占用。
	"--disable-dev-shm-usage",         //禁用Chrome在/dev/shm文件系统中分配的共享内存，这可以减少Chrome进程的内存占用。
	"--fast-start",                    //启用快速启动功能，这可以加快Chrome的启动速度。
	"--disable-hardware-acceleration", //禁用硬件加速功能，这可以在某些旧的计算机和旧的显卡上降低Chrome的资源消耗，但可能会影响一些图形性能和视频播放。

	"--browser-test",                  //启用浏览器测试模式，这可以对Chrome进行优化以实现更低的内存占用率。
	"--disable-background-networking", // 禁用Chrome的后台网络请求，可以降低Chrome对内存的占用。
	"--disable-site-isolation-trials", //禁用站点隔离特性试验，这可以提高Chrome的性能并降低其内存占用。
	"--disable-gpu",                   //禁用硬件加速功能，这可以降低一些GPU相关任务的CPU占用，但可能降低图形性能和视频播放能力。
	"--process-per-tab",               //为每个标签页启动一个新的进程，这可以有效防止内存泄漏，并大幅度降低Chrome进程的内存占用。
	"--no-pings",                      //禁用 ping。
	"--no-zygote",                     //禁用 zygote 进程。

	"--mute-audio",                                //禁用音频。
	"--no-first-run",                              //不显示欢迎页面。
	"--no-default-browser-check",                  //不检查是否为默认浏览器。
	"--disable-software-rasterizer",               //禁用软件光栅化器
	"--disable-cloud-import",                      //禁用云导入。
	"--disable-gesture-typing",                    //禁用手势输入。
	"--disable-offer-store-unmasked-wallet-cards", //禁用钱包卡。
	"--disable-offer-upload-credit-cards",         //禁用上传信用卡。
	"--disable-print-preview",                     //禁用打印预览。
	"--disable-voice-input",                       //禁用语音输入。
	"--disable-wake-on-wifi",                      //禁用 Wi-Fi 唤醒。
	"--disable-cookie-encryption",                 //禁用 cookie 加密
	"--ignore-gpu-blocklist",                      //忽略 GPU 阻止列表。
	"--enable-async-dns",                          //启用异步 DNS。
	"--enable-simple-cache-backend",               //启用简单缓存后端
	"--enable-tcp-fast-open",                      //启用 TCP 快速打开。
	"--prerender-from-omnibox=disabled",           //用于禁用从地址栏预渲染页面
	"--disable-features=AutoUpdate,site-per-process,Profiles,EasyBakeWebBundler,MultipleCompositingThreads,AudioServiceOutOfProcess,TranslateUI,BackgroundSync,ClientHints,NetworkQualityEstimator,PasswordGeneration,PrefetchPrivacyChanges,TabHoverCards,ImprovedCookieControls,LazyFrameLoading,GlobalMediaControls,DestroyProfileOnBrowserClose,MediaRouter,DialMediaRouteProvider,AcceptCHFrame,AutoExpandDetailsElement,CertificateTransparencyComponentUpdater,AvoidUnnecessaryBeforeUnloadCheckSync,Translate", // 禁用一些 Chrome 功能。

	"--disable-field-trial-config", //禁用实验室配置，在禁用情况下，不会向远程服务器报告任何配置或默认设置。
	"--disable-back-forward-cache", //禁用后退/前进缓存。

	"--allow-pre-commit-input", //允许在提交前输入词语。

	"--no-service-autorun", //不自动运行服务。

	"--ignore-certificate-errors",
	"--aggressive-cache-discard",               //启用缓存丢弃。
	"--disable-ipc-flooding-protection",        //禁用 IPC 洪水保护。
	"--disable-default-apps",                   //禁用默认应用
	"--enable-webgl",                           //启用 WebGL。
	"--disable-breakpad",                       //禁用 崩溃报告
	"--disable-component-update",               //禁用组件更新。
	"--disable-domain-reliability",             //禁用域可靠性。
	"--disable-sync",                           //禁用同步。
	"--disable-client-side-phishing-detection", //禁用客户端钓鱼检测。
	"--disable-hang-monitor",                   //禁用挂起监视器
	"--disable-popup-blocking",                 //禁用弹出窗口阻止。

	"--disable-crash-reporter",                                        //禁用崩溃报告器。
	"--disable-background-timer-throttling",                           //禁用后台计时器限制。
	"--disable-backgrounding-occluded-windows",                        //禁用后台窗口。
	"--disable-infobars",                                              //禁用信息栏。
	"--hide-scrollbars",                                               //隐藏滚动条。
	"--disable-prompt-on-repost",                                      //禁用重新提交提示。
	"--metrics-recording-only",                                        //仅记录指标。
	"--safebrowsing-disable-auto-update",                              //禁用安全浏览自动更新。
	"--use-mock-keychain",                                             //使用模拟钥匙串。
	"--force-webrtc-ip-handling-policy=default_public_interface_only", //强制 WebRTC IP 处理策略。
	"--enable-webrtc-stun-origin=false",                               //用于禁用WebRTC的STUN源，而
	"--enforce-webrtc-ip-permission-check=false",                      //用于禁用WebRTC的IP权限检查。

	"--disable-session-crashed-bubble", //禁用会话崩溃气泡。
	"--disable-renderer-backgrounding", //禁用渲染器后台化。
	"--font-render-hinting=none",       //禁用字体渲染提示
	"--disable-logging",                //禁用日志记录。

	"--ssl-protocol=any",                                   //使用任何 SSL 协议。
	"--disable-partial-raster",                             //禁用部分光栅化
	"--disable-component-extensions-with-background-pages", //禁用具有后台页面的组件扩展。
	"--disable-translate",                                  //禁用翻译。
	"--password-store=basic",                               //使用基本密码存储。
	"--disable-image-animation-resync",                     //禁用图像动画重新
}

//go:embed devices.json
var devicesData []byte

var Devices = loadDevicesData()

func loadDevicesData() map[string]cdp.Device {
	var result map[string]cdp.Device
	if err := json.Unmarshal(devicesData, &result); err != nil {
		log.Panic(err)
	}
	return result
}
func downLoadChrome(preCtx context.Context, dirUrl string, version int) error {
	reqCli, err := requests.NewClient(preCtx)
	if err != nil {
		return err
	}
	resp, err := reqCli.Request(preCtx, "get", dirUrl)
	if err != nil {
		return err
	}
	var fileDir string
	var fileTime int64
	var ver int
	jsonData, err := resp.Json()
	if err != nil {
		return err
	}
	for _, dir := range jsonData.Array() {
		if tempTime, err := time.Parse(fmt.Sprintf("%sT%sZ", time.DateOnly, time.TimeOnly), dir.Get("date").String()); err == nil {
			if versionRe := re.Search(`\d+`, dir.Get("name").String()); versionRe != nil {
				if versionInt, err := strconv.Atoi(versionRe.Group()); err == nil {
					if versionInt == version || tempTime.Unix() > fileTime {
						fileDir = dir.Get("url").String()
						fileTime = tempTime.Unix()
						ver = versionInt
						if versionInt == version {
							break
						}
					}
				}
			}
		}
	}
	if fileTime == 0 {
		return errors.New("not found chrome dir")
	}
	resp, err = reqCli.Request(preCtx, "get", fileDir)
	if err != nil {
		return err
	}
	if jsonData, err = resp.Json(); err != nil {
		return err
	}

	fileUrl := jsonData.Get("0.url").String()
	resp, err = reqCli.Request(preCtx, "get", fileUrl, requests.RequestOption{Bar: true})
	if err != nil {
		return err
	}
	zipData, err := zip.NewReader(bytes.NewReader(resp.Content()), int64(len(resp.Content())))
	if err != nil {
		return err
	}
	mainDir, err := conf.GetMainDirPath()
	if err != nil {
		return err
	}
	mainDir = tools.PathJoin(mainDir, strconv.Itoa(ver))
	for _, file := range zipData.File {
		filePath := tools.PathJoin(mainDir, file.Name)
		log.Printf("解压文件: %s", filePath)
		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, 0777)
			continue
		}
		fileDirPath := tools.PathJoin(filePath, "..")
		if !tools.PathExist(fileDirPath) {
			if err = os.MkdirAll(fileDirPath, 0777); err != nil {
				return err
			}
		}
		readBody, err := file.Open()
		if err != nil {
			return err
		}
		tempBody := bytes.NewBuffer(nil)
		if err = tools.CopyWitchContext(preCtx, tempBody, readBody); err != nil {
			return err
		}
		if err = os.WriteFile(filePath, tempBody.Bytes(), 0777); err != nil {
			return err
		}
	}
	return err
}
func DownChrome(preCtx context.Context, versions ...int) error {
	var version int
	if len(versions) > 0 {
		version = versions[0]
	}
	switch runtime.GOOS {
	case "windows":
		return downLoadChrome(preCtx, "https://registry.npmmirror.com/-/binary/chromium-browser-snapshots/Win_x64/", version)
	case "darwin":
		return downLoadChrome(preCtx, "https://registry.npmmirror.com/-/binary/chromium-browser-snapshots/Mac/", version)
	case "linux":
		return downLoadChrome(preCtx, "https://registry.npmmirror.com/-/binary/chromium-browser-snapshots/Linux_x64/", version)
	default:
		return errors.New("dont know goos")
	}
}

// 新建浏览器
func NewClient(preCtx context.Context, options ...ClientOption) (client *Client, err error) {
	var option ClientOption
	if len(options) > 0 {
		option = options[0]
	}
	if preCtx == nil {
		preCtx = context.TODO()
	}
	ctx, cnl := context.WithCancel(preCtx)
	defer func() {
		if err != nil {
			cnl()
		}
	}()
	if option.Width == 0 {
		option.Width = 1000
	}
	if option.Height == 0 {
		option.Height = 1200
	}
	globalReqCli, err := requests.NewClient(preCtx, requests.ClientOption{
		Proxy:       option.Proxy,
		GetProxy:    option.GetProxy,
		Ja3Spec:     option.Ja3Spec,
		Ja3:         option.Ja3,
		H2Ja3Spec:   option.H2Ja3Spec,
		H2Ja3:       option.H2Ja3,
		RedirectNum: -1,
		DisDecode:   true,
	})
	if err != nil {
		return nil, err
	}
	var cli *cmd.Client
	var isReplaceRequest bool
	if option.Host == "" || option.Port == 0 {
		if cli, isReplaceRequest, err = runChrome(ctx, &option); err != nil {
			return
		}
	}
	if option.Ja3Spec.IsSet() || option.H2Ja3Spec.IsSet() || option.DataCache {
		isReplaceRequest = true
	}
	client = &Client{
		isReplaceRequest: isReplaceRequest,
		proxy:            option.Proxy,
		getProxy:         option.GetProxy,
		dataCache:        option.DataCache,
		headless:         option.Headless,
		ctx:              ctx,
		cnl:              cnl,
		cmdCli:           cli,
		db:               db.NewClient[cdp.FulData](ctx, cnl),
		host:             option.Host,
		port:             option.Port,
		globalReqCli:     globalReqCli,
		stealth:          option.Stealth,
	}
	go tools.Signal(ctx, client.Close)
	return client, client.init()
}
func (obj *Client) RequestClient() *requests.Client {
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
		fmt.Sprintf("http://%s:%d/json/version", obj.host, obj.port),
		requests.RequestOption{
			Timeout:  3,
			DisProxy: true,
			ErrCallBack: func(ctx context.Context, err error) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second):
				}
				if obj.cmdCli.Err() != nil {
					return obj.cmdCli.Err()
				}
				return nil
			},
			ResponseCallBack: func(ctx context.Context, r *requests.Response) error {
				if r.StatusCode() == 200 {
					return nil
				}
				return errors.New("code error")
			},
			TryNum: 10,
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
	obj.webSock, err = cdp.NewWebSock(
		obj.ctx,
		obj.globalReqCli,
		fmt.Sprintf("ws://%s:%d/devtools/browser/%s", obj.host, obj.port, browWsRs.Group(1)),
		cdp.WebSockOption{},
		obj.db,
	)
	return err
}

// 浏览器是否结束的 chan
func (obj *Client) Done() <-chan struct{} {
	return obj.webSock.Done()
}

// 返回浏览器远程控制的地址
func (obj *Client) Addr() string {
	return net.JoinHostPort(obj.host, strconv.Itoa(obj.port))
}

// 关闭浏览器
func (obj *Client) Close() {
	if obj.globalReqCli != nil {
		obj.globalReqCli.Close()
	}
	if obj.webSock != nil {
		obj.webSock.BrowserClose()
	}
	if obj.cmdCli != nil {
		obj.cmdCli.Close()
	}
	obj.db.Close()
	obj.cnl()
}

type PageOption struct {
	Proxy     string
	DataCache bool //开启数据缓存
	Stealth   bool //是否开启随机指纹
}

// 新建标签页
func (obj *Client) NewPage(preCtx context.Context, options ...PageOption) (*Page, error) {
	var option PageOption
	if len(options) > 0 {
		option = options[0]
	}
	if !option.DataCache {
		option.DataCache = obj.dataCache
	}
	isReplaceRequest := obj.isReplaceRequest
	if !isReplaceRequest {
		if option.DataCache {
			isReplaceRequest = true
		} else if option.Proxy != "" && option.Proxy != obj.proxy {
			isReplaceRequest = true
		}
	}

	rs, err := obj.webSock.TargetCreateTarget(preCtx, "")
	if err != nil {
		return nil, err
	}
	targetId, ok := rs.Result["targetId"].(string)
	if !ok {
		return nil, errors.New("not found targetId")
	}
	ctx, cnl := context.WithCancel(obj.ctx)
	page := &Page{
		id:               targetId,
		preWebSock:       obj.webSock,
		port:             obj.port,
		host:             obj.host,
		ctx:              ctx,
		cnl:              cnl,
		headless:         obj.headless,
		globalReqCli:     obj.globalReqCli,
		stealth:          obj.stealth,
		isReplaceRequest: isReplaceRequest,
		pageAfterTime:    time.NewTimer(0),
		domAfterTime:     time.NewTimer(0),
	}
	if err = page.init(obj.globalReqCli, option, obj.db); err != nil {
		return nil, err
	}
	if isReplaceRequest {
		if err = page.Request(preCtx, defaultRequestFunc); err != nil {
			return nil, err
		}
	}
	return page, nil
}
