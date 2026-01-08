package browser

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gospider007/blog"
	"github.com/gospider007/cmd"
	"github.com/gospider007/conf"
	"github.com/gospider007/requests"
	"github.com/gospider007/tools"
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
const revision = "1200"

// var playwright_cdn_mirrors = []string{
// 	"playwright.azureedge.net",
// 	"playwright-verizon.azureedge.net",
// 	"playwright-akamai.azureedge.net",
// }

const playwright_cdn_mirror = "playwright.azureedge.net"

// from https://playwright.azureedge.net/builds/chromium/1150/chromium-mac-arm64.zip
// from https://playwright.azureedge.net/builds/chromium/1183/chromium-win64.zip

// var mac13_arm64 = fmt.Sprintf("https://%s/builds/chromium/%s/chromium-mac-arm64.zip", playwright_cdn_mirror, revision)
// var debian12_arm64 = fmt.Sprintf("https://%s/builds/chromium/%s/chromium-linux-arm64.zip", playwright_cdn_mirror, revision)
var debian12_x64 = fmt.Sprintf("https://%s/builds/chromium/%s/chromium-linux.zip", playwright_cdn_mirror, revision)
var mac13 = fmt.Sprintf("https://%s/builds/chromium/%s/chromium-mac-arm64.zip", playwright_cdn_mirror, revision)
var mac13_intel = fmt.Sprintf("https://%s/builds/chromium/%s/chromium-mac.zip", playwright_cdn_mirror, revision)
var win64 = fmt.Sprintf("https://%s/builds/chromium/%s/chromium-win64.zip", playwright_cdn_mirror, revision)

type downClient struct {
	sync.Mutex
}

var oneDown = &downClient{}

func verifyEvalPath(path string) error {
	path = strings.TrimSuffix(path, ".exe")
	if strings.Contains(path, "Chrome") || strings.Contains(path, "Chromium") || strings.Contains(path, "chrome") || strings.Contains(path, "chromium") {
		return nil
	}
	if strings.Contains(path, "msedge") {
		return nil
	}
	return errors.New("请输入正确的浏览器路径,如: c:/chrome.exe")
}
func findChromeApp(dirPath string) (string, error) {
	fdirs, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}
	dirs := []os.DirEntry{}
	for _, fdir := range fdirs {
		if strings.HasPrefix(fdir.Name(), ".") {
			continue
		}
		dirs = append(dirs, fdir)
	}
	// log.Print(dirPath, len(dirs))
	if len(dirs) == 0 {
		return "", errors.New("空目录")
	}
	if len(dirs) == 1 {
		path := tools.PathJoin(dirPath, dirs[0].Name())
		if dirs[0].IsDir() {
			return findChromeApp(path)
		} else {
			return path, nil
		}
	}
	sort.SliceStable(dirs, func(x, y int) bool {
		return len(dirs[x].Name()) < len(dirs[y].Name())
	})
	for _, dir := range dirs {
		path := tools.PathJoin(dirPath, dir.Name())
		name := strings.ToLower(dir.Name())
		if dir.IsDir() {
			if strings.HasSuffix(dir.Name(), ".app") || dir.Name() == "MacOS" {
				return findChromeApp(path)
			}
		} else {
			if strings.Contains(name, "chrome") || strings.Contains(name, "chromium") {
				return path, nil
			}
		}
	}
	return "", errors.New("找不到文件")
}

func (obj *downClient) getChromePath(preCtx context.Context) (string, error) {
	obj.Lock()
	defer obj.Unlock()
	chromeDir, err := conf.GetMainDirPath()
	if err != nil {
		return "", err
	}
	var chromeDownUrl string
	chromeDir = tools.PathJoin(chromeDir, revision)
	chromePath, _ := findChromeApp(chromeDir)
	switch runtime.GOOS {
	case "windows":
		chromeDownUrl = win64
	case "darwin":
		if runtime.GOARCH == "arm64" {
			chromeDownUrl = mac13
		} else {
			chromeDownUrl = mac13_intel
		}
	case "linux":
		chromeDownUrl = debian12_x64
	default:
		return "", errors.New("dont know goos")
	}
	if chromePath == "" || !tools.PathExist(chromePath) {
		if err = downChrome(preCtx, chromeDir, chromeDownUrl); err != nil {
			return "", err
		}
		if chromePath == "" {
			chromePath, err = findChromeApp(chromeDir)
			if err != nil {
				return "", err
			}
			log.Print(chromePath)
		}
		if !tools.PathExist(chromePath) {
			return "", errors.New("not found chrome")
		}
	}
	return chromePath, nil
}
func (obj *Client) runChrome() error {
	var err error
	if obj.option.Host == "" {
		obj.option.Host = "127.0.0.1"
	}
	if obj.option.Port == 0 {
		obj.option.Port, err = tools.FreePort()
		if err != nil {
			return err
		}
	}
	if obj.option.ChromePath == "" {
		obj.option.ChromePath, err = oneDown.getChromePath(obj.ctx)
		if err != nil {
			return err
		}
	} else {
		fileInfo, err := os.Stat(obj.option.ChromePath)
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			obj.option.ChromePath, err = findChromeApp(obj.option.ChromePath)
			if err != nil {
				return err
			}
		}
	}
	if err = verifyEvalPath(obj.option.ChromePath); err != nil {
		return err
	}
	var isDelDir bool
	if obj.option.UserDir == "" {
		obj.option.UserDir, err = conf.GetTempChromeDirPath()
		if err != nil {
			return err
		}
		isDelDir = true
	}
	args := []string{}
	args = append(args, chromeArgs...)
	if obj.option.UserAgent != "" && obj.option.Headless {
		args = append(args, fmt.Sprintf("--user-agent=%s", obj.option.UserAgent))
	}
	if obj.option.Headless {
		args = append(args, "--headless=new")
	}
	args = append(args, fmt.Sprintf(`--user-data-dir=%s`, obj.option.UserDir))
	args = append(args, fmt.Sprintf("--remote-debugging-port=%d", obj.option.Port))
	args = append(args, fmt.Sprintf("--window-size=%d,%d", obj.option.Width, obj.option.Height))
	args = append(args, fmt.Sprintf("--parent-pid=%d", os.Getpid()))
	args = append(args, fmt.Sprintf("--custom-parent-pid=%d", os.Getpid()))
	for _, arg := range obj.option.Args {
		if !slices.Contains(args, arg) {
			args = append(args, arg)
		}
	}
	var closeCallBack func()
	if isDelDir {
		closeCallBack = func() {
			for i := 0; i < 10; i++ {
				if os.RemoveAll(obj.option.UserDir) == nil {
					return
				}
				time.Sleep(time.Millisecond * 300)
			}
		}
	}
	obj.cmdCli, err = cmd.NewClient(obj.ctx, cmd.ClientOption{
		Name:          obj.option.ChromePath,
		Args:          args,
		CloseCallBack: closeCallBack,
	})
	if err != nil {
		return err
	}
	go obj.cmdCli.Run()
	return obj.cmdCli.Err()
}

func downChrome(preCtx context.Context, chromeDir, chromeDownUrl string) error {
	log.Print("download chrome... ", chromeDownUrl)
	resp, err := requests.Get(preCtx, chromeDownUrl, requests.RequestOption{
		Bar: true,
	})
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
