package browser

import (
	"archive/zip"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gospider007/blog"
	"github.com/gospider007/cdp"
	"github.com/gospider007/cmd"
	"github.com/gospider007/conf"
	"github.com/gospider007/gson"
	"github.com/gospider007/gtls"
	"github.com/gospider007/ja3"
	"github.com/gospider007/proxy"
	"github.com/gospider007/re"
	"github.com/gospider007/requests"
	"github.com/gospider007/tools"
	"golang.org/x/exp/slices"
)

// https://github.com/microsoft/playwright/blob/main/packages/playwright-core/src/server/registry/nativeDeps.ts
var libsChrome = []string{
	"libasound2",
	"libatk-bridge2.0-0",
	"libatk1.0-0",
	"libatspi2.0-0",
	"libcairo2",
	"libcups2",
	"libdbus-1-3",
	"libdrm2",
	"libgbm1",
	"libglib2.0-0",
	"libnspr4",
	"libnss3",
	"libpango-1.0-0",
	"libx11-6",
	"libxcb1",
	"libxcomposite1",
	"libxdamage1",
	"libxext6",
	"libxfixes3",
	"libxkbcommon0",
	"libxrandr2",
}

// apt install -y libasound2 libatk-bridge2.0-0 libatk1.0-0 libatspi2.0-0 libcairo2 libcups2 libdbus-1-3 libdrm2 libgbm1 libglib2.0-0 libnspr4 libnss3 libpango-1.0-0 libx11-6 libxcb1 libxcomposite1 libxdamage1 libxext6 libxfixes3 libxkbcommon0 libxrandr2
// yum install -y libasound.so.2 libatk-bridge-2.0.so.0 libatk-1.0.so.0 libatspi.so.0 libcairo.so.2 libcups.so.2 libdbus-1.so.3 libdrm.so.2 libgbm.so.1 libgio-2.0.so.0 libnspr4.so libnss3.so libpango-1.0.so.0 libX11.so.6 libxcb.so.1 libXcomposite.so.1 libXdamage.so.1 libXext.so.6 libXfixes.so.3 libxkbcommon.so.0 libXrandr.so.2

// https://github.com/microsoft/playwright/blob/main/packages/playwright-core/src/server/registry/nativeDeps.ts
var libsPackage = map[string]string{
	"libsoup-3.0.so.0":       "libsoup-3.0-0",
	"libasound.so.2":         "libasound2",
	"libatk-1.0.so.0":        "libatk1.0-0",
	"libatk-bridge-2.0.so.0": "libatk-bridge2.0-0",
	"libatspi.so.0":          "libatspi2.0-0",
	"libcairo.so.2":          "libcairo2",
	"libcups.so.2":           "libcups2",
	"libdbus-1.so.3":         "libdbus-1-3",
	"libdrm.so.2":            "libdrm2",
	"libgbm.so.1":            "libgbm1",
	"libgio-2.0.so.0":        "libglib2.0-0",
	"libglib-2.0.so.0":       "libglib2.0-0",
	"libgobject-2.0.so.0":    "libglib2.0-0",
	"libnspr4.so":            "libnspr4",
	"libnss3.so":             "libnss3",
	"libnssutil3.so":         "libnss3",
	"libpango-1.0.so.0":      "libpango-1.0-0",
	"libsmime3.so":           "libnss3",
	"libX11.so.6":            "libx11-6",
	"libxcb.so.1":            "libxcb1",
	"libXcomposite.so.1":     "libxcomposite1",
	"libXdamage.so.1":        "libxdamage1",
	"libXext.so.6":           "libxext6",
	"libXfixes.so.3":         "libxfixes3",
	"libxkbcommon.so.0":      "libxkbcommon0",
	"libXrandr.so.2":         "libxrandr2",
}

func PrintLibs() {
	log.Print(blog.Color(1, "debian libs\n"), blog.Color(2, "apt install -y ", strings.Join(libsChrome, " ")))
	libsPackage2 := map[string]string{}
	for key, val := range libsPackage {
		libsPackage2[val] = key
	}
	libs2 := []string{}
	for _, val := range libsChrome {
		libs2 = append(libs2, libsPackage2[val])
	}
	log.Print(blog.Color(1, "centos libs\n"), blog.Color(2, "yum install -y ", strings.Join(libs2, " ")))
}

// https://github.com/microsoft/playwright/blob/main/packages/playwright-core/browsers.json
const revision = "1150"

// var playwright_cdn_mirrors = []string{
// 	"playwright.azureedge.net",
// 	"playwright-verizon.azureedge.net",
// 	"playwright-akamai.azureedge.net",
// }

const playwright_cdn_mirror = "playwright.azureedge.net"

// from https://playwright.azureedge.net/builds/chromium/1150/chromium-mac-arm64.zip

// var mac13_arm64 = fmt.Sprintf("https://%s/builds/chromium/%s/chromium-mac-arm64.zip", playwright_cdn_mirror, revision)
// var debian12_arm64 = fmt.Sprintf("https://%s/builds/chromium/%s/chromium-linux-arm64.zip", playwright_cdn_mirror, revision)
var debian12_x64 = fmt.Sprintf("https://%s/builds/chromium/%s/chromium-linux.zip", playwright_cdn_mirror, revision)
var mac13 = fmt.Sprintf("https://%s/builds/chromium/%s/chromium-mac-arm64.zip", playwright_cdn_mirror, revision)
var win64 = fmt.Sprintf("https://%s/builds/chromium/%s/chromium-win64.zip", playwright_cdn_mirror, revision)

type Client struct {
	userAgent        string
	proxy            string
	isReplaceRequest bool
	cmdCli           *cmd.Client
	globalReqCli     *requests.Client
	addr             string
	proxyAddr        string
	ctx              context.Context
	cnl              context.CancelFunc
	webSock          *cdp.WebSock
	stealth          bool //是否开启随机指纹
	browserContextId string
}
type ClientOption struct {
	Host       string
	Port       int
	ChromePath string   //chrome path
	UserDir    string   //user dir
	Args       []string //start args
	Headless   bool     //is headless
	UserAgent  string
	Proxy      string                                                  //support http,https,socks5,ex: http://127.0.0.1:7005
	GetProxy   func(ctx context.Context, url *url.URL) (string, error) //pr
	Width      int64                                                   //browser width,1200
	Height     int64                                                   //browser height,605
	Stealth    bool                                                    //is stealth
	Ja3Spec    ja3.Ja3Spec                                             //ja3
}

type downClient struct {
	sync.Mutex
}

var oneDown = &downClient{}

func verifyEvalPath(path string) error {
	if strings.HasSuffix(path, "chrome.exe") || strings.HasSuffix(path, "Chromium") || strings.HasSuffix(path, "chrome") || strings.HasSuffix(path, "chromium") {
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
	chromeDir, err := conf.GetMainDirPath()
	if err != nil {
		return "", err
	}
	var chromePath string
	var chromeDownUrl string
	switch runtime.GOOS {
	case "windows":
		chromeDir = tools.PathJoin(chromeDir, revision)
		chromePath = tools.PathJoin(chromeDir, "chrome-win", "chrome.exe")
		chromeDownUrl = win64
	case "darwin":
		chromeDir = tools.PathJoin(chromeDir, revision)
		chromePath = tools.PathJoin(chromeDir, "chrome-mac", "Chromium.app", "Contents", "MacOS", "Chromium")
		chromeDownUrl = mac13
	case "linux":
		chromeDir = tools.PathJoin(chromeDir, revision)
		chromePath = tools.PathJoin(chromeDir, "chrome-linux", "chrome")
		chromeDownUrl = debian12_x64
	default:
		return "", errors.New("dont know goos")
	}
	if !tools.PathExist(chromePath) {
		if err = downChrome(preCtx, chromeDir, chromeDownUrl); err != nil {
			return "", err
		}
		if !tools.PathExist(chromePath) {
			return "", errors.New("not found chrome")
		}
	}
	return chromePath, nil
}
func (obj *Client) runChrome(option *ClientOption) error {
	var err error
	if option.Host == "" {
		option.Host = "127.0.0.1"
	}
	if option.Port == 0 {
		option.Port, err = tools.FreePort()
		if err != nil {
			return err
		}
	}
	if option.UserAgent == "" {
		option.UserAgent = tools.UserAgent
	}
	if option.ChromePath == "" {
		option.ChromePath, err = oneDown.getChromePath(obj.ctx)
		if err != nil {
			return err
		}
	}
	if err = verifyEvalPath(option.ChromePath); err != nil {
		return err
	}
	var isDelDir bool
	if option.UserDir == "" {
		option.UserDir, err = conf.GetTempChromeDirPath()
		if err != nil {
			return err
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
			return err
		}
		if proxyUrl.User == nil {
			args = append(args, fmt.Sprintf(`--proxy-server=%s`, proxyUrl.String()))
		} else {
			obj.isReplaceRequest = true
		}
	}
	args = append(args, fmt.Sprintf(`--user-data-dir=%s`, option.UserDir))
	args = append(args, fmt.Sprintf("--remote-debugging-port=%d", option.Port))
	args = append(args, fmt.Sprintf("--window-size=%d,%d", option.Width, option.Height))
	args = append(args, fmt.Sprintf("--parent-pid=%d", os.Getpid()))
	args = append(args, fmt.Sprintf("--custom-parent-pid=%d", os.Getpid()))
	for _, arg := range option.Args {
		if !slices.Contains(args, arg) {
			args = append(args, arg)
		}
	}
	var closeCallBack func()
	if isDelDir {
		closeCallBack = func() {
			for i := 0; i < 10; i++ {
				if os.RemoveAll(option.UserDir) == nil {
					return
				}
				time.Sleep(time.Millisecond * 300)
			}
		}
	}
	obj.cmdCli, err = cmd.NewClient(obj.ctx, cmd.ClientOption{
		Name:          option.ChromePath,
		Args:          args,
		CloseCallBack: closeCallBack,
	})
	if err != nil {
		return err
	}
	go obj.cmdCli.Run()
	return obj.cmdCli.Err()
}

var chromeArgs = []string{
	// "--disable-site-isolation-trials", //被识别
	// "--virtual-time-budget=1000", //缩短setTimeout  setInterval 的时间1000秒:目前不生效，不知道以后会不会生效，等生效了再打开
	//远程调试
	"--remote-allow-origins=*",
	// 自动化选项禁用
	"--useAutomationExtension=false",                //禁用自动化扩展。
	"--excludeSwitches=enable-automation",           //禁用自动化
	"--disable-blink-features=AutomationControlled", //禁用 Blink 引擎的自动化控制。
	// 稳定性选项
	"--no-sandbox",      //禁用 Chrome 的沙盒模式。
	"--set-uid-sandbox", //命令行参数用于设置 Chrome 进程运行时使用的 UID，从而提高 Chrome 浏览器的安全性
	"--set-gid-sandbox", //命令行参数用于设置 Chrome 进程运行时使用的 GID，从而提高 Chrome 浏览器的安全性
	"--enable-features=NetworkService,NetworkServiceInProcess",
	"--disable-features=WebRtcHideLocalIpsWithMdns,EnablePasswordsAccountStorage,FlashDeprecationWarning,UserAgentClientHint,AutoUpdate,site-per-process,Profiles,EasyBakeWebBundler,MultipleCompositingThreads,AudioServiceOutOfProcess,TranslateUI,BlinkGenPropertyTrees,BackgroundSync,ClientHints,NetworkQualityEstimator,PasswordGeneration,PrefetchPrivacyChanges,TabHoverCards,ImprovedCookieControls,LazyFrameLoading,GlobalMediaControls,DestroyProfileOnBrowserClose,MediaRouter,DialMediaRouteProvider,AcceptCHFrame,AutoExpandDetailsElement,CertificateTransparencyComponentUpdater,AvoidUnnecessaryBeforeUnloadCheckSync,Translate,TabFreezing,TabDiscarding,HttpsUpgrades", // 禁用一些 Chrome 功能。
	"--blink-settings=primaryHoverType=2,availableHoverTypes=2,primaryPointerType=4,availablePointerTypes=4,imagesEnabled=true", //Blink 设置。
	"--ignore-ssl-errors=true", //忽略 SSL 错误。
	"--disable-setuid-sandbox", //重要headless

	"--disable-extensions", //禁用所有扩展程序，这可以降低Chrome对内存的占用。
	"--disable-plugins",    //禁用所有已安装的Chrome浏览器插件。
	"--fast-start",         //启用快速启动功能，这可以加快Chrome的启动速度。

	"--disable-background-networking", // 禁用Chrome的后台网络请求，可以降低Chrome对内存的占用。
	"--browser-test",                  //启用浏览器测试模式，这可以对Chrome进行优化以实现更低的内存占用率。
	"--disable-gpu",                   //禁用硬件加速功能，这可以降低一些GPU相关任务的CPU占用，但可能降低图形性能和视频播放能力。
	"--no-pings",                      //禁用 ping。
	"--no-zygote",                     //禁用 zygote 进程。

	"--mute-audio",                                //禁用音频。
	"--no-first-run",                              //不显示欢迎页面。
	"--no-default-browser-check",                  //不检查是否为默认浏览器。
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

	"--disable-session-crashed-bubble",                     //禁用会话崩溃气泡。
	"--font-render-hinting=none",                           //禁用字体渲染提示
	"--disable-logging",                                    //禁用日志记录。
	"--disable-partial-raster",                             //禁用部分光栅化
	"--disable-component-extensions-with-background-pages", //禁用具有后台页面的组件扩展。
	"--disable-translate",                                  //禁用翻译。
	"--password-store=basic",                               //使用基本密码存储。
	"--disable-image-animation-resync",                     //禁用图像动画重新
	"--use-gl=swiftshader",                                 //可以在不支持硬件加速的系统或设备上提供基本的图形渲染功能。
	"--window-position=0,0",                                //窗口起始位置
	"--disable-remote-fonts",                               //禁用远程字体加载。这个参数可以防止Chrome从远程服务器加载字体，从而减少与服务器的连接，增强隐私。
	"--disable-geolocation",                                //禁用地理位置定位功能。这个参数可以防止Chrome获取您的地理位置信息，增强隐私。
	"--disable-media-stream",                               //禁用媒体流功能。这个参数可以防止Chrome访问您的摄像头和麦克风，增强隐私。
	"--disable-preconnect",                                 //禁用预连接。预连接是一种优化技术，可以在您点击链接之前预先建立与目标服务器的连接，以加快页面加载速度。禁用预连接可以减少被追踪的可能性。
	"--force-color-profile=srgb",
	"--disable-dev-shm-usage",                //禁用Chrome在/dev/shm文件系统中分配的共享内存
	"--disable-background-mode",              // 禁用浏览器后台模式。
	"--disable-hardware-acceleration",        //禁用硬件加速功能，这可以在某些旧的计算机和旧的显卡上降低Chrome的资源消耗，但可能会影响一些图形性能和视频播放。
	"--disable-renderer-backgrounding",       //禁用渲染器后台化。,反爬用到
	"--disable-web-security",                 //关闭同源策略，抖音需要
	"--disable-search-engine-choice-screen",  //用于禁用搜索引擎选择屏幕。该选项通常用于自定义 Chrome 浏览器的行为。
	"--renderer",                             //使进程作为渲染器而不是浏览器运行。
	"--disable-renderer-accessibility",       //关闭渲染器中的辅助功能。
	"--disable-renderer-priority-management", //根本不管理渲染器进程优先级。
	"--allow-running-insecure-content",       //在安全页面上加载不安全内容时禁用警告消息，这可以节省测试时间。
	"--disable-add-to-shelf",                 //禁用“添加到工具架”功能，该功能对于自动测试是不必要的。
	"--disable-checker-imaging",              //禁用检查器成像，减少测试期间不必要的图像处理。
	"--disable-datasaver-prompt",             //禁用与测试方案无关的数据保护程序提示
	"--disable-desktop-notifications",        //禁用桌面通知，避免在测试期间中断。
	"--disable-notifications",                //禁用浏览器通知，避免在测试期间中断。
	"--test-type",
}

func downChrome(preCtx context.Context, chromeDir, chromeDownUrl string) error {
	log.Print("download chrome... ", chromeDownUrl)
	resp, err := requests.Get(preCtx, chromeDownUrl, requests.RequestOption{Bar: true})
	if err != nil {
		return err
	}
	zipData, err := zip.NewReader(bytes.NewReader(resp.Content()), int64(len(resp.Content())))
	if err != nil {
		return err
	}
	log.Printf("准备环境中...")
	for _, file := range zipData.File {
		filePath := tools.PathJoin(chromeDir, file.Name)
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
		defer readBody.Close()
		tempBody := bytes.NewBuffer(nil)
		if err = tools.CopyWitchContext(preCtx, tempBody, readBody); err != nil {
			return err
		}
		if err = os.WriteFile(filePath, tempBody.Bytes(), 0777); err != nil {
			return err
		}
	}
	log.Printf("准备环境ok")
	return err
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
	globalReqCli, err := requests.NewClient(preCtx, requests.ClientOption{
		MaxRetries:  2,
		Proxy:       option.Proxy,
		GetProxy:    option.GetProxy,
		Ja3:         true,
		Ja3Spec:     option.Ja3Spec,
		MaxRedirect: -1,
	})
	if err != nil {
		return nil, err
	}
	if runtime.GOOS == "linux" {
		option.Headless = true
	}
	if option.Width == 0 {
		option.Width = 1200
	}
	if option.Height == 0 {
		option.Height = 605
	}
	if option.UserAgent == "" {
		option.UserAgent = tools.UserAgent
	}
	client = &Client{
		userAgent:    option.UserAgent,
		proxy:        option.Proxy,
		globalReqCli: globalReqCli,
		stealth:      option.Stealth,
	}
	client.ctx, client.cnl = context.WithCancel(preCtx)
	if option.Host == "" || option.Port == 0 {
		if err = client.runChrome(&option); err != nil {
			return
		}
		var proxyHost string
		for _, addr := range gtls.GetHosts(4) {
			if addr.IsGlobalUnicast() {
				proxyHost = addr.String()
				break
			}
		}
		if proxyHost == "" {
			return client, errors.New("获取内网地址失败")
		}

		client.addr = net.JoinHostPort(option.Host, strconv.Itoa(option.Port))
		proxCli, err := proxy.NewClient(client.ctx, proxy.ClientOption{
			DisVerify: true,
			HttpConnectCallBack: func(r *http.Request) error {
				r.Host = client.addr
				return nil
			},
		})
		if err != nil {
			return client, err
		}
		go proxCli.Run()
		client.proxyAddr = proxCli.Addr()
	} else {
		client.addr = net.JoinHostPort(option.Host, strconv.Itoa(option.Port))
		client.proxyAddr = client.addr
	}
	go tools.Signal(preCtx, client.Close)
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
		fmt.Sprintf("http://%s/json/version", obj.addr),
		requests.RequestOption{
			Timeout:  time.Second * 3,
			DisProxy: true,
			ErrCallBack: func(ctx context.Context, _ *requests.RequestOption, _ *requests.Response, err error) error {
				select {
				case <-obj.cmdCli.Ctx().Done():
					return context.Cause(obj.cmdCli.Ctx())
				case <-ctx.Done():
					return nil
				case <-time.After(time.Second):
				}
				if obj.cmdCli.Err() != nil {
					return obj.cmdCli.Err()
				}
				return nil
			},
			ResultCallBack: func(ctx context.Context, _ *requests.RequestOption, r *requests.Response) error {
				if r.StatusCode() == 200 {
					return nil
				}
				time.Sleep(time.Second)
				return errors.New("code error")
			},
			MaxRetries: 10,
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
		fmt.Sprintf("ws://%s/devtools/browser/%s", obj.addr, browWsRs.Group(1)),
		cdp.WebSockOption{},
	)
	if err != nil {
		return err
	}
	contextData, err := obj.webSock.TargetCreateBrowserContext(obj.ctx)
	if err != nil {
		return err
	}
	contextResult, err := gson.Decode(contextData.Result)
	if err != nil {
		return err
	}
	obj.browserContextId = contextResult.Get("browserContextId").String()
	if obj.browserContextId == "" {
		return errors.New("not found browserContextId")
	}
	return err
}

// 浏览器是否结束的 chan
func (obj *Client) Done() <-chan struct{} {
	return obj.webSock.Done()
}
func (obj *Client) Error() (err error) {
	return obj.webSock.Error()
}

// 返回浏览器远程控制的地址
func (obj *Client) Addr() string {
	return obj.addr
}

// 返回浏览器远程控制的代理地址
func (obj *Client) ProxyAddr() string {
	return obj.proxyAddr
}

// 关闭浏览器
func (obj *Client) Close() {
	if obj.globalReqCli != nil {
		obj.globalReqCli.Close()
	}
	if obj.webSock != nil {
		obj.webSock.TargetDisposeBrowserContext(obj.browserContextId)
		obj.webSock.BrowserClose()
		obj.webSock.Close()
	}
	if obj.cmdCli != nil {
		obj.cmdCli.Close()
	}
	obj.cnl()
}

type PageOption struct {
	Proxy   string
	Stealth bool //是否开启随机指纹
}

// 新建标签页
func (obj *Client) NewPage(preCtx context.Context, options ...PageOption) (*Page, error) {
	rs, err := obj.webSock.TargetCreateTarget(preCtx, "")
	if err != nil {
		return nil, err
	}
	targetId, ok := rs.Result["targetId"].(string)
	if !ok {
		return nil, errors.New("not found targetId")
	}
	return obj.NewPageWithTargetId(preCtx, targetId, "page", options...)
}

// 设置浏览器的地理位置
func (obj *Client) SetGeolocation(preCtx context.Context, latitude float64, longitude float64) error {
	_, err := obj.webSock.EmulationSetGeolocationOverride(preCtx, latitude, longitude)
	return err
}
func (obj *Client) NewPageWithTargetId(preCtx context.Context, targetId string, targetType string, options ...PageOption) (*Page, error) {
	var option PageOption
	if len(options) > 0 {
		option = options[0]
	}
	isReplaceRequest := obj.isReplaceRequest
	if !isReplaceRequest {
		if option.Proxy != "" && option.Proxy != obj.proxy {
			isReplaceRequest = true
		}
	}
	if !option.Stealth {
		option.Stealth = obj.stealth
	}
	ctx, cnl := context.WithCancel(obj.ctx)
	page := &Page{
		userAgent:        obj.userAgent,
		option:           option,
		targetId:         targetId,
		targetType:       targetType,
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
		if err := page.Request(preCtx, defaultRequestFunc); err != nil {
			return nil, err
		}
	}
	return page, nil
}
