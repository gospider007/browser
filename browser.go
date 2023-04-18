package browser

import (
	"archive/zip"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/baixudong/gospider/cdp"
	"gitee.com/baixudong/gospider/cmd"
	"gitee.com/baixudong/gospider/conf"
	"gitee.com/baixudong/gospider/db"
	"gitee.com/baixudong/gospider/ja3"
	"gitee.com/baixudong/gospider/proxy"
	"gitee.com/baixudong/gospider/re"
	"gitee.com/baixudong/gospider/requests"
	"gitee.com/baixudong/gospider/tools"
)

var version = "020"

func delTempDir(dir string) {
	timeout := 10 * 1000         //10s
	sleep := 100                 //每次睡眠0.1s
	totalSize := timeout / sleep //总共100次
	for i := 0; i < totalSize; i++ {
		if i > 0 {
			time.Sleep(time.Millisecond * time.Duration(sleep))
		}
		if os.RemoveAll(dir) == nil {
			return
		}
	}
}

// go build -ldflags="-H windowsgui" -o browser/browserCmd.exe main.go
// go build -o browser/browserCmd main.go
func BrowserCmdMain() (err error) {
	preCtx := context.Background()
	ctx, cnl := context.WithCancelCause(preCtx)
	pipData := make(chan struct{})
	data := map[string]any{}
	args := []string{}
	var cmdCli *cmd.Client
	go func() (err error) {
		defer cnl(err)
		jsonDecode := json.NewDecoder(os.Stdin)
		if err = jsonDecode.Decode(&data); err != nil || data["name"] == nil {
			return
		}
		jsonData := tools.Any2json(data)
		for _, arg := range jsonData.Get("args").Array() {
			args = append(args, arg.String())
		}
		cmdCli = cmd.NewClient(ctx, cmd.ClientOption{Name: jsonData.Get("name").String(), Args: args})
		go func() {
			err = cmdCli.Run()
		}()
		if err != nil {
			return
		}
		close(pipData)
		return jsonDecode.Decode(&data)
	}()
	select {
	case <-cmdCli.Done():
		return cmdCli.Err()
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-pipData:
	case <-time.After(time.Second * 2):
		return
	}
	//dong some thing
	for _, arg := range args {
		if strings.Contains(arg, "--user-data-dir=") {
			rs := re.Search(`--user-data-dir="(.*?)"`, arg)
			if rs != nil {
				defer delTempDir(rs.Group(1))
			} else {
				rs = re.Search(`--user-data-dir=(\S*)`, arg)
				if rs != nil {
					defer delTempDir(rs.Group(1))
				}
			}
		}
	}
	//join
	defer cmdCli.Close()
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-cmdCli.Done():
		return cmdCli.Err()
	}
}

//go:embed stealth.min.js
var stealth string

//go:embed stealth2.min.js
var stealth2 string

//go:embed stealth3.min.js
var stealth3 string

type Client struct {
	proxy        string
	getProxy     func(ctx context.Context, url *url.URL) (string, error)
	db           *db.Client[cdp.FulData]
	cmdCli       *cmd.Client
	globalReqCli *requests.Client
	port         int
	host         string
	lock         sync.Mutex
	ctx          context.Context
	cnl          context.CancelFunc
	webSock      *cdp.WebSock
	proxyCli     *proxy.Client
	dataCache    bool
	headless     bool
	stealth      bool //是否开启随机指纹
}
type ClientOption struct {
	ChromePath  string   //chrome浏览器执行路径
	Host        string   //连接host
	Port        int      //连接port
	UserDir     string   //设置用户目录
	Args        []string //启动参数
	Headless    bool     //是否使用无头
	DataCache   bool     //开启数据缓存
	Ja3Spec     ja3.ClientHelloSpec
	Ja3         bool
	UserAgent   string
	Proxy       string                                                  //代理http,https,socks5,ex: http://127.0.0.1:7005
	GetProxy    func(ctx context.Context, url *url.URL) (string, error) //代理
	Width       int64                                                   //浏览器的宽
	Height      int64                                                   //浏览器的高
	Stealth     bool                                                    //是否开启随机指纹
	DisDnsCache bool                                                    //关闭dns 解析
}

//go:embed browserCmd.exe
var browserCmdWindows []byte

//go:embed browserCmd
var browserCmdLinux []byte

func getCmdName() (string, error) {
	mainDir, err := conf.GetMainDirPath()
	if err != nil {
		return "", err
	}
	fileName := tools.PathJoin(mainDir, fmt.Sprintf("browserCmd%s", version))
	if runtime.GOOS == "windows" {
		fileName += ".exe"
	}
	if !tools.PathExist(fileName) {
		os.MkdirAll(mainDir, 0777)
		if runtime.GOOS == "windows" {
			if err = os.WriteFile(fileName, browserCmdWindows, 0777); err != nil {
				return "", err
			}
		} else {
			if err = os.WriteFile(fileName, browserCmdLinux, 0777); err != nil {
				return "", err
			}
		}
	}
	return fileName, nil
}

type downClient struct {
	sync.Mutex
}

var oneDown = &downClient{}
var chromeVersion = 1108766

func verifyEvalPath(path string) error {
	if !tools.PathExist(path) {
		return errors.New("路径不存在")
	}

	if strings.HasSuffix(path, "chrome.exe") || strings.HasSuffix(path, "Chromium.app") || strings.HasSuffix(path, "chrome") {
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
		chromeArgs = append(chromeArgs,
			"--use-gl=swiftshader",
			"--disable-gpu",
		)
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

func clearTemp() {
	tempDir := os.TempDir()
	dirs, err := os.ReadDir(tempDir)
	if err != nil {
		return
	}
	for _, dir := range dirs {
		if re.Search(fmt.Sprintf(`%s\d+$`, conf.TempChromeDir), dir.Name()) != nil {
			os.RemoveAll(tools.PathJoin(tempDir, dir.Name()))
		}
	}
}
func runChrome(ctx context.Context, option *ClientOption) (*cmd.Client, error) {
	fileName, err := getCmdName()
	if err != nil {
		return nil, err
	}
	if option.Host == "" {
		option.Host = "127.0.0.1"
	}
	if option.Port == 0 {
		option.Port, err = tools.FreePort()
		if err != nil {
			return nil, err
		}
	}
	if option.UserAgent == "" {
		option.UserAgent = requests.UserAgent
	}
	if option.ChromePath == "" {
		option.ChromePath, err = oneDown.getChromePath(ctx)
		if err != nil {
			return nil, err
		}
	}
	if err = verifyEvalPath(option.ChromePath); err != nil {
		return nil, err
	}
	if option.UserDir == "" {
		option.UserDir, err = os.MkdirTemp(os.TempDir(), conf.TempChromeDir)
		if err != nil {
			return nil, err
		}
	}

	cli := cmd.NewLeakClient(ctx, cmd.ClientOption{
		Name: fileName,
	})
	inP, err := cli.StdInPipe()
	if err != nil {
		return nil, err
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
		args = append(args, fmt.Sprintf(`--proxy-server=%s`, option.Proxy))
	}
	args = append(args, fmt.Sprintf(`--user-data-dir=%s`, option.UserDir))
	args = append(args, fmt.Sprintf("--remote-debugging-port=%d", option.Port))
	args = append(args, fmt.Sprintf("--window-size=%d,%d", option.Width, option.Height))

	args = append(args, option.Args...)
	_, err = inP.Write(tools.StringToBytes(tools.Any2json(map[string]any{
		"name": option.ChromePath,
		"args": args,
	}).Raw))
	if err != nil {
		return nil, err
	}
	go cli.Run()
	return cli, cli.Err()
}

var chromeArgs = []string{
	"--no-sandbox",                                  //禁用 Chrome 的沙盒模式。
	"--useAutomationExtension=false",                //禁用自动化扩展。
	"--excludeSwitches=enable-automation",           //禁用自动化
	"--disable-blink-features=AutomationControlled", //禁用 Blink 引擎的自动化控制。
	"--blink-settings=primaryHoverType=2,availableHoverTypes=2,primaryPointerType=4,availablePointerTypes=4,imagesEnabled=true", //Blink 设置。
	"--ignore-ssl-errors=true", //忽略 SSL 错误。
	// "--virtual-time-budget=1000", //缩短setTimeout  setInterval 的时间1000秒:目前不生效，不知道以后会不会生效，等生效了再打开

	"--no-pings",                                  //禁用 ping。
	"--no-zygote",                                 //禁用 zygote 进程。
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
	"--disable-features=AudioServiceOutOfProcess,IsolateOrigins,site-per-process,TranslateUI,BlinkGenPropertyTrees", // 禁用一些 Chrome 功能。
	"--aggressive-cache-discard",                                      //启用缓存丢弃。
	"--disable-ipc-flooding-protection",                               //禁用 IPC 洪水保护。
	"--disable-default-apps",                                          //禁用默认应用
	"--enable-webgl",                                                  //启用 WebGL。
	"--disable-breakpad",                                              //禁用 Breakpad。
	"--disable-component-update",                                      //禁用组件更新。
	"--disable-domain-reliability",                                    //禁用域可靠性。
	"--disable-sync",                                                  //禁用同步。
	"--disable-client-side-phishing-detection",                        //禁用客户端钓鱼检测。
	"--disable-hang-monitor",                                          //禁用挂起监视器
	"--disable-popup-blocking",                                        //禁用弹出窗口阻止。
	"--disable-crash-reporter",                                        //禁用崩溃报告器。
	"--disable-dev-shm-usage",                                         //禁用 /dev/shm 使用。
	"--disable-background-networking",                                 //禁用后台网络。
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
	for _, dir := range resp.Json().Array() {
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
	fileUrl := resp.Json().Get("0.url").String()
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
		fileDirPath := tools.PathJoin(filePath, "..")
		log.Printf("解压文件: %s", filePath)
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
		if _, err = io.Copy(tempBody, readBody); err != nil {
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
	clearTemp()
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
		option.Width = 1492
	}
	if option.Height == 0 {
		option.Height = 843
	}
	var cli *cmd.Client
	if option.Host == "" || option.Port == 0 {
		if cli, err = runChrome(ctx, &option); err != nil {
			return
		}
	}
	globalReqCli, err := requests.NewClient(preCtx, requests.ClientOption{
		Proxy:       option.Proxy,
		GetProxy:    option.GetProxy,
		Ja3Spec:     option.Ja3Spec,
		Ja3:         option.Ja3,
		DisDnsCache: option.DisDnsCache,
	})
	if err != nil {
		return nil, err
	}
	globalReqCli.RedirectNum = -1
	globalReqCli.DisDecode = true
	client = &Client{
		proxy:        option.Proxy,
		getProxy:     option.GetProxy,
		dataCache:    option.DataCache,
		headless:     option.Headless,
		ctx:          ctx,
		cnl:          cnl,
		cmdCli:       cli,
		db:           db.NewClient[cdp.FulData](ctx, cnl),
		host:         option.Host,
		port:         option.Port,
		globalReqCli: globalReqCli,
		stealth:      option.Stealth,
	}
	return client, client.init()
}
func (obj *Client) RequestClient() *requests.Client {
	return obj.globalReqCli
}

// 浏览器初始化
func (obj *Client) init() error {
	rs, err := obj.globalReqCli.Request(obj.ctx, "get",
		fmt.Sprintf("http://%s:%d/json/version", obj.host, obj.port),
		requests.RequestOption{
			DisProxy: true,
			ErrCallBack: func(err error) bool {
				time.Sleep(time.Millisecond * 1000)
				return false
			},
			AfterCallBack: func(r *requests.Response) error {
				if r.StatusCode() == 200 {
					return nil
				}
				return errors.New("code error")
			},
			TryNum: 10,
		})
	if err != nil {
		obj.cmdCli.Err()
		return err
	}
	wsUrl := rs.Json().Get("webSocketDebuggerUrl").String()
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
	if err != nil {
		return err
	}
	var host string
	if runtime.GOOS == "windows" {
		host = "0.0.0.0"
	} else {
		host = tools.GetHost(4).String()
	}
	obj.proxyCli, err = proxy.NewClient(obj.ctx, proxy.ClientOption{
		Host:  host,
		Port:  obj.port,
		Proxy: fmt.Sprintf("http://%s:%d", obj.host, obj.port),
	})
	if err != nil {
		return err
	}
	obj.proxyCli.DisVerify = true
	go obj.proxyCli.Run()
	return obj.proxyCli.Err
}

// 浏览器是否结束的 chan
func (obj *Client) Done() <-chan struct{} {
	return obj.webSock.Done()
}

// 返回浏览器远程控制的地址
func (obj *Client) Addr() string {
	return obj.proxyCli.Addr()
}

// 关闭浏览器
func (obj *Client) Close() (err error) {
	if obj.webSock != nil {
		if err = obj.webSock.BrowserClose(); err != nil {
			return err
		}
	}
	if obj.cmdCli != nil {
		obj.cmdCli.Close()
	}
	obj.cnl()
	obj.db.Close()
	return
}

type PageOption struct {
	Proxy     string
	DataCache bool //开启数据缓存
	Ja3Spec   ja3.ClientHelloSpec
	Ja3       bool
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
	if option.Ja3 || option.Ja3Spec.IsSet() {
		option.DataCache = true
	} else if option.Proxy != "" && option.Proxy != obj.proxy {
		option.DataCache = true
	} else if obj.getProxy != nil {
		option.DataCache = true
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
		id:           targetId,
		preWebSock:   obj.webSock,
		port:         obj.port,
		host:         obj.host,
		ctx:          ctx,
		cnl:          cnl,
		headless:     obj.headless,
		globalReqCli: obj.globalReqCli,
		stealth:      obj.stealth,
		dataCache:    option.DataCache,
	}
	if err = page.init(obj.globalReqCli, option, obj.db); err != nil {
		return nil, err
	}
	if option.DataCache {
		if err = page.Request(preCtx, defaultRequestFunc); err != nil {
			return nil, err
		}
	}
	return page, nil
}
