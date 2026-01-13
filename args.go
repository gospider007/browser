package browser

var chromeArgs = []string{
	"--disable-translate",
	"--no-startup-window", //是否多开一个窗口
	"--use-mock-keychain", //使用模拟钥匙串。
	"--force-webrtc-ip-handling-policy",
	"--webrtc-ip-handling-policy=disable_non_proxied_udp",
	"--remote-allow-origins=*",
	"--useAutomationExtension=false",                //禁用自动化扩展。
	"--excludeSwitches=enable-automation",           //禁用自动化
	"--disable-blink-features=AutomationControlled", //禁用 Blink 引擎的自动化控制。
	"--no-sandbox",                                  //禁用 Chrome 的沙盒模式。
	"--enable-features=NetworkService,NetworkServiceInProcess",
	"--disable-features=VizDisplayCompositor,WebRtcHideLocalIpsWithMdns,EnablePasswordsAccountStorage,FlashDeprecationWarning,UserAgentClientHint,AutoUpdate,site-per-process,Profiles,EasyBakeWebBundler,MultipleCompositingThreads,AudioServiceOutOfProcess,TranslateUI,BlinkGenPropertyTrees,BackgroundSync,ClientHints,NetworkQualityEstimator,PasswordGeneration,PrefetchPrivacyChanges,TabHoverCards,ImprovedCookieControls,LazyFrameLoading,GlobalMediaControls,DestroyProfileOnBrowserClose,MediaRouter,DialMediaRouteProvider,AcceptCHFrame,AutoExpandDetailsElement,CertificateTransparencyComponentUpdater,AvoidUnnecessaryBeforeUnloadCheckSync,Translate,TabFreezing,TabDiscarding,HttpsUpgrades", // 禁用一些 Chrome 功能。
	"--blink-settings=primaryHoverType=2,availableHoverTypes=2,primaryPointerType=4,availablePointerTypes=4,imagesEnabled=true", //Blink 设置。
	"--ignore-ssl-errors=true", //忽略 SSL 错误。
	"--disable-setuid-sandbox", //重要headless
	// "--disable-web-security",          //关闭同源策略，抖音需要, 开启会导致 cloudflare 验证不过
	// "--disable-site-isolation-trials", // 开启会导致 cloudflare 验证不过
	//==============================

	"--disable-3d-apis",
	"--disable-webgl",
	"--disable-gpu",        //禁用硬件加速功能，这可以降低一些GPU相关任务的CPU占用，但可能降低图形性能和视频播放能力。
	"--use-gl=swiftshader", //可以在不支持硬件加速的系统或设备上提供基本的图形渲染功能。
	//==================

	"--disable-hidpi-scaling",
	"--disable-perfetto",
	"--disable-hardware-acceleration", //禁用硬件加速功能，这可以在某些旧的计算机和旧的显卡上降低Chrome的资源消耗，但可能会影响一些图形性能和视频播放。
	"--virtual-time-budget=1",         //缩短setTimeout  setInterval 的时间1000秒:目前不生效，不知道以后会不会生效，等生效了再打开

	//远程调试
	"--set-uid-sandbox", //命令行参数用于设置 Chrome 进程运行时使用的 UID，从而提高 Chrome 浏览器的安全性
	"--set-gid-sandbox", //命令行参数用于设置 Chrome 进程运行时使用的 GID，从而提高 Chrome 浏览器的安全性

	"--disable-extensions", //禁用所有扩展程序，这可以降低Chrome对内存的占用。
	"--disable-plugins",    //禁用所有已安装的Chrome浏览器插件。
	"--fast-start",         //启用快速启动功能，这可以加快Chrome的启动速度。

	"--disable-background-networking", // 禁用Chrome的后台网络请求，可以降低Chrome对内存的占用。
	"--browser-test",                  //启用浏览器测试模式，这可以对Chrome进行优化以实现更低的内存占用率。
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
	"--disable-breakpad",                       //禁用 崩溃报告
	"--disable-component-update",               //禁用组件更新。
	"--disable-domain-reliability",             //禁用域可靠性。
	"--disable-sync",                           //禁用同步。
	"--disable-client-side-phishing-detection", //禁用客户端钓鱼检测。
	"--disable-hang-monitor",                   //禁用挂起监视器
	"--disable-popup-blocking",                 //禁用弹出窗口阻止。

	"--disable-crash-reporter",                                  //禁用崩溃报告器。
	"--disable-background-timer-throttling",                     //禁用后台计时器限制。
	"--disable-backgrounding-occluded-windows",                  //禁用后台窗口。
	"--disable-infobars",                                        //禁用信息栏。
	"--hide-scrollbars",                                         //隐藏滚动条。
	"--disable-prompt-on-repost",                                //禁用重新提交提示。
	"--metrics-recording-only",                                  //仅记录指标。
	"--safebrowsing-disable-auto-update",                        //禁用安全浏览自动更新。
	"--force-webrtc-ip-handling-policy=disable_non_proxied_udp", //强制 WebRTC IP 处理策略。
	"--enable-webrtc-stun-origin=false",                         //用于禁用WebRTC的STUN源，而
	"--enforce-webrtc-ip-permission-check=false",                //用于禁用WebRTC的IP权限检查。

	"--disable-session-crashed-bubble",                     //禁用会话崩溃气泡。
	"--font-render-hinting=none",                           //禁用字体渲染提示
	"--disable-logging",                                    //禁用日志记录。
	"--disable-partial-raster",                             //禁用部分光栅化
	"--disable-component-extensions-with-background-pages", //禁用具有后台页面的组件扩展。
	"--disable-translate",                                  //禁用翻译。
	"--password-store=basic",                               //使用基本密码存储。
	"--disable-image-animation-resync",                     //禁用图像动画重新
	"--window-position=0,0",                                //窗口起始位置
	"--disable-remote-fonts",                               //禁用远程字体加载。这个参数可以防止Chrome从远程服务器加载字体，从而减少与服务器的连接，增强隐私。
	"--disable-geolocation",                                //禁用地理位置定位功能。这个参数可以防止Chrome获取您的地理位置信息，增强隐私。
	"--disable-media-stream",                               //禁用媒体流功能。这个参数可以防止Chrome访问您的摄像头和麦克风，增强隐私。
	"--disable-preconnect",                                 //禁用预连接。预连接是一种优化技术，可以在您点击链接之前预先建立与目标服务器的连接，以加快页面加载速度。禁用预连接可以减少被追踪的可能性。

	"--force-color-profile=srgb",
	"--disable-dev-shm-usage",   //禁用Chrome在/dev/shm文件系统中分配的共享内存
	"--disable-background-mode", // 禁用浏览器后台模式。

	"--disable-renderer-backgrounding",       //禁用渲染器后台化。,反爬用到
	"--disable-search-engine-choice-screen",  //用于禁用搜索引擎选择屏幕。该选项通常用于自定义 Chrome 浏览器的行为。
	"--renderer",                             //使进程作为渲染器而不是浏览器运行。
	"--disable-renderer-accessibility",       //关闭渲染器中的辅助功能。
	"--disable-renderer-priority-management", //根本不管理渲染器进程优先级。

	"--allow-running-insecure-content", //在安全页面上加载不安全内容时禁用警告消息，这可以节省测试时间。
	"--disable-add-to-shelf",           //禁用“添加到工具架”功能，该功能对于自动测试是不必要的。
	"--disable-checker-imaging",        //禁用检查器成像，减少测试期间不必要的图像处理。
	"--disable-datasaver-prompt",       //禁用与测试方案无关的数据保护程序提示
	"--disable-desktop-notifications",  //禁用桌面通知，避免在测试期间中断。
	"--test-type",
}
