package browser

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gospider007/cdp"
	"github.com/gospider007/gson"
	"github.com/gospider007/re"
	"github.com/gospider007/tools"
)

//go:embed devices.json
var devicesJson string

func Json2Devices() {
	lls, err := gson.Decode(devicesJson)
	if err != nil {
		log.Panic(err)
	}
	var txt string

	for key, device := range lls.Map() {
		key = strings.ToTitle(key)
		key := re.Sub(`\s|\(|\)|\+`, "", key)
		deviceStr := fmt.Sprintf(`var %s = cdp.Device{
		UserAgent: "%s",
		Viewport: cdp.Viewport{
			Width:  %d,
			Height: %d,
		},
		DeviceScaleFactor: %d,
		IsMobile:          %t,
		HasTouch:          %t,
	}`, key, device.Get("user_agent").String(), device.Get("viewport.width").Int(), device.Get("viewport.height").Int(),
			device.Get("device_scale_factor").Int(), device.Get("is_mobile").Bool(), device.Get("has_touch").Bool())

		txt += deviceStr + "\n"
	}
	err = os.WriteFile("devices.txt", tools.StringToBytes(txt), 0777)
	if err != nil {
		log.Panic(err)
	}
}

var BLACKBERRYPLAYBOOKLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (PlayBook; U; RIM Tablet OS 2.1.0; en-US) AppleWebKit/536.2+ (KHTML like Gecko) Version/15.4 Safari/536.2+",
	Viewport: cdp.Viewport{
		Width:  1024,
		Height: 600,
	},
	DeviceScaleFactor: 1,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONEXR = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  414,
		Height: 896,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var DESKTOPFIREFOXHIDPI = cdp.Device{
	UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0.1) Gecko/20100101 Firefox/94.0.1",
	Viewport: cdp.Viewport{
		Width:  1280,
		Height: 720,
	},
	DeviceScaleFactor: 2,
	IsMobile:          false,
	HasTouch:          false,
}
var GALAXYS9 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.0.0; SM-G965U Build/R16NW) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  320,
		Height: 658,
	},
	DeviceScaleFactor: 4,
	IsMobile:          true,
	HasTouch:          true,
}
var IPADMINILANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPad; CPU OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  1024,
		Height: 768,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE8PLUSLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  736,
		Height: 414,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONEX = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  375,
		Height: 812,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var JIOPHONE2LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Mobile; LYF/F300B/LYF-F300B-001-01-15-130718-i;Android; rv:94.0.1) Gecko/48.0 Firefox/94.0.1 KAIOS/2.5",
	Viewport: cdp.Viewport{
		Width:  320,
		Height: 240,
	},
	DeviceScaleFactor: 1,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL2LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.0; Pixel 2 Build/OPD3.170816.012) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  731,
		Height: 411,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL2XL = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.0.0; Pixel 2 XL Build/OPD1.170816.004) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  411,
		Height: 823,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYNOTE3LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; U; Android 4.3; en-us; SM-N900T Build/JSS15J) AppleWebKit/534.30 (KHTML, like Gecko) Version/15.4 Mobile Safari/534.30",
	Viewport: cdp.Viewport{
		Width:  640,
		Height: 360,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE13 = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  390,
		Height: 664,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var NOKIAN9LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (MeeGo; NokiaN9) AppleWebKit/534.13 (KHTML, like Gecko) NokiaBrowser/8.5.0 Mobile Safari/534.13",
	Viewport: cdp.Viewport{
		Width:  854,
		Height: 480,
	},
	DeviceScaleFactor: 1,
	IsMobile:          true,
	HasTouch:          true,
}
var DESKTOPFIREFOX = cdp.Device{
	UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0.1) Gecko/20100101 Firefox/94.0.1",
	Viewport: cdp.Viewport{
		Width:  1280,
		Height: 720,
	},
	DeviceScaleFactor: 1,
	IsMobile:          false,
	HasTouch:          false,
}
var IPHONE11PROLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  724,
		Height: 325,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var BLACKBERRYZ30LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (BB10; Touch) AppleWebKit/537.10+ (KHTML, like Gecko) Version/15.4 Mobile Safari/537.10+",
	Viewport: cdp.Viewport{
		Width:  640,
		Height: 360,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPADPRO11 = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPad; CPU OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  834,
		Height: 1194,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE13PROMAXLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  832,
		Height: 380,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var JIOPHONE2 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Mobile; LYF/F300B/LYF-F300B-001-01-15-130718-i;Android; rv:94.0.1) Gecko/48.0 Firefox/94.0.1 KAIOS/2.5",
	Viewport: cdp.Viewport{
		Width:  240,
		Height: 320,
	},
	DeviceScaleFactor: 1,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYS5 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  360,
		Height: 640,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPADGEN7 = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPad; CPU OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  810,
		Height: 1080,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE12PRO = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  390,
		Height: 664,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS6P = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.0.0; Nexus 6P Build/OPP3.170518.006) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  412,
		Height: 732,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL5 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 11; Pixel 5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  393,
		Height: 727,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE7 = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  375,
		Height: 667,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var MICROSOFTLUMIA550LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Windows Phone 10.0; Android 4.2.1; Microsoft; Lumia 550) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36 Edge/14.14263",
	Viewport: cdp.Viewport{
		Width:  360,
		Height: 640,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL4 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 10; Pixel 4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  353,
		Height: 745,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE7LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  667,
		Height: 375,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE11PROMAX = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  414,
		Height: 715,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE13PRO = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  390,
		Height: 664,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE13MINI = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  375,
		Height: 629,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL3 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 9; Pixel 3 Build/PQ1A.181105.017.A1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  393,
		Height: 786,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL4A5G = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 11; Pixel 4a (5G)) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  412,
		Height: 765,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYS5LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  640,
		Height: 360,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS10LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 10 Build/MOB31T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  1280,
		Height: 800,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var BLACKBERRYZ30 = cdp.Device{
	UserAgent: "Mozilla/5.0 (BB10; Touch) AppleWebKit/537.10+ (KHTML, like Gecko) Version/15.4 Mobile Safari/537.10+",
	Viewport: cdp.Viewport{
		Width:  360,
		Height: 640,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE12PROLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  750,
		Height: 340,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE13PROLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  750,
		Height: 342,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYNOTEIILANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; U; Android 4.1; en-us; GT-N7100 Build/JRO03C) AppleWebKit/534.30 (KHTML, like Gecko) Version/15.4 Mobile Safari/534.30",
	Viewport: cdp.Viewport{
		Width:  640,
		Height: 360,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE8 = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  375,
		Height: 667,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS6 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 7.1.1; Nexus 6 Build/N6F26U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  412,
		Height: 732,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL2 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.0; Pixel 2 Build/OPD3.170816.012) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  411,
		Height: 731,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var MOTOG4 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 7.0; Moto G (4)) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  360,
		Height: 640,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var DESKTOPEDGEHIDPI = cdp.Device{
	UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Safari/537.36 Edg/98.0.4695.0",
	Viewport: cdp.Viewport{
		Width:  1280,
		Height: 720,
	},
	DeviceScaleFactor: 2,
	IsMobile:          false,
	HasTouch:          false,
}
var IPHONE11 = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  414,
		Height: 715,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var KINDLEFIREHDXLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; U; en-us; KFAPWI Build/JDQ39) AppleWebKit/535.19 (KHTML, like Gecko) Silk/3.13 Safari/535.19 Silk-Accelerated=true",
	Viewport: cdp.Viewport{
		Width:  1280,
		Height: 800,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL2XLLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.0.0; Pixel 2 XL Build/OPD1.170816.004) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  823,
		Height: 411,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYNOTE3 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; U; Android 4.3; en-us; SM-N900T Build/JSS15J) AppleWebKit/534.30 (KHTML, like Gecko) Version/15.4 Mobile Safari/534.30",
	Viewport: cdp.Viewport{
		Width:  360,
		Height: 640,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var MICROSOFTLUMIA950 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Windows Phone 10.0; Android 4.2.1; Microsoft; Lumia 950) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36 Edge/14.14263",
	Viewport: cdp.Viewport{
		Width:  360,
		Height: 640,
	},
	DeviceScaleFactor: 4,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS10 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 10 Build/MOB31T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  800,
		Height: 1280,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL5LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 11; Pixel 5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  802,
		Height: 293,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var DESKTOPEDGE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Safari/537.36 Edg/98.0.4695.0",
	Viewport: cdp.Viewport{
		Width:  1280,
		Height: 720,
	},
	DeviceScaleFactor: 1,
	IsMobile:          false,
	HasTouch:          false,
}
var GALAXYNOTEII = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; U; Android 4.1; en-us; GT-N7100 Build/JRO03C) AppleWebKit/534.30 (KHTML, like Gecko) Version/15.4 Mobile Safari/534.30",
	Viewport: cdp.Viewport{
		Width:  360,
		Height: 640,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPADGEN7LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPad; CPU OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  1080,
		Height: 810,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE6PLUSLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  736,
		Height: 414,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE13MINILANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  712,
		Height: 327,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS7 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 7 Build/MOB30X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  600,
		Height: 960,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var KINDLEFIREHDX = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; U; en-us; KFAPWI Build/JDQ39) AppleWebKit/535.19 (KHTML, like Gecko) Silk/3.13 Safari/535.19 Silk-Accelerated=true",
	Viewport: cdp.Viewport{
		Width:  800,
		Height: 1280,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS5LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  640,
		Height: 360,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var MOTOG4LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 7.0; Moto G (4)) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  640,
		Height: 360,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var DESKTOPCHROME = cdp.Device{
	UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  1280,
		Height: 720,
	},
	DeviceScaleFactor: 1,
	IsMobile:          false,
	HasTouch:          false,
}
var IPHONEXRLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  896,
		Height: 414,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYTABS4 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.1.0; SM-T837A) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  712,
		Height: 1138,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONESELANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/15.4 Mobile/14E304 Safari/602.1",
	Viewport: cdp.Viewport{
		Width:  568,
		Height: 320,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS4 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 4.4.2; Nexus 4 Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  384,
		Height: 640,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS4LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 4.4.2; Nexus 4 Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  640,
		Height: 384,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPADMINI = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPad; CPU OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  768,
		Height: 1024,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE11PROMAXLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  808,
		Height: 364,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE12LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  750,
		Height: 340,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS5 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  360,
		Height: 640,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS7LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 7 Build/MOB30X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  960,
		Height: 600,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL4A5GLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 11; Pixel 4a (5G)) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  840,
		Height: 312,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYS8LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 7.0; SM-G950U Build/NRD90M) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  740,
		Height: 360,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE8LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  667,
		Height: 375,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE12 = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  390,
		Height: 664,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE13PROMAX = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  428,
		Height: 746,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var DESKTOPSAFARI = cdp.Device{
	UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Safari/605.1.15",
	Viewport: cdp.Viewport{
		Width:  1280,
		Height: 720,
	},
	DeviceScaleFactor: 2,
	IsMobile:          false,
	HasTouch:          false,
}
var IPADGEN6LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPad; CPU OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  1024,
		Height: 768,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE12PROMAXLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  832,
		Height: 378,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var BLACKBERRYPLAYBOOK = cdp.Device{
	UserAgent: "Mozilla/5.0 (PlayBook; U; RIM Tablet OS 2.1.0; en-US) AppleWebKit/536.2+ (KHTML like Gecko) Version/15.4 Safari/536.2+",
	Viewport: cdp.Viewport{
		Width:  600,
		Height: 1024,
	},
	DeviceScaleFactor: 1,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYS9LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.0.0; SM-G965U Build/R16NW) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  658,
		Height: 320,
	},
	DeviceScaleFactor: 4,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONESE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/15.4 Mobile/14E304 Safari/602.1",
	Viewport: cdp.Viewport{
		Width:  320,
		Height: 568,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var LGOPTIMUSL70LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; U; Android 4.4.2; en-us; LGMS323 Build/KOT49I.MS32310c) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  640,
		Height: 384,
	},
	DeviceScaleFactor: 1,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS5XLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.0.0; Nexus 5X Build/OPR4.170623.006) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  732,
		Height: 412,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS6PLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.0.0; Nexus 6P Build/OPP3.170518.006) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  732,
		Height: 412,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE7PLUSLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  736,
		Height: 414,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE11PRO = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  375,
		Height: 635,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var LGOPTIMUSL70 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; U; Android 4.4.2; en-us; LGMS323 Build/KOT49I.MS32310c) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  384,
		Height: 640,
	},
	DeviceScaleFactor: 1,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYSIIILANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; U; Android 4.0; en-us; GT-I9300 Build/IMM76D) AppleWebKit/534.30 (KHTML, like Gecko) Version/15.4 Mobile Safari/534.30",
	Viewport: cdp.Viewport{
		Width:  640,
		Height: 360,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPADGEN6 = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPad; CPU OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  768,
		Height: 1024,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL3LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 9; Pixel 3 Build/PQ1A.181105.017.A1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  786,
		Height: 393,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYTABS4LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.1.0; SM-T837A) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  1138,
		Height: 712,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE7PLUS = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  414,
		Height: 736,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS5X = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 8.0.0; Nexus 5X Build/OPR4.170623.006) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  412,
		Height: 732,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var DESKTOPCHROMEHIDPI = cdp.Device{
	UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  1280,
		Height: 720,
	},
	DeviceScaleFactor: 2,
	IsMobile:          false,
	HasTouch:          false,
}
var IPHONE6LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  667,
		Height: 375,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var MICROSOFTLUMIA950LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Windows Phone 10.0; Android 4.2.1; Microsoft; Lumia 950) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36 Edge/14.14263",
	Viewport: cdp.Viewport{
		Width:  640,
		Height: 360,
	},
	DeviceScaleFactor: 4,
	IsMobile:          true,
	HasTouch:          true,
}
var NOKIALUMIA520 = cdp.Device{
	UserAgent: "Mozilla/5.0 (compatible; MSIE 10.0; Windows Phone 8.0; Trident/6.0; IEMobile/10.0; ARM; Touch; NOKIA; Lumia 520)",
	Viewport: cdp.Viewport{
		Width:  320,
		Height: 533,
	},
	DeviceScaleFactor: 1,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYSIII = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; U; Android 4.0; en-us; GT-I9300 Build/IMM76D) AppleWebKit/534.30 (KHTML, like Gecko) Version/15.4 Mobile Safari/534.30",
	Viewport: cdp.Viewport{
		Width:  360,
		Height: 640,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var GALAXYS8 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 7.0; SM-G950U Build/NRD90M) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  360,
		Height: 740,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE6PLUS = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  414,
		Height: 736,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE12PROMAX = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  428,
		Height: 746,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var NOKIALUMIA520LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (compatible; MSIE 10.0; Windows Phone 8.0; Trident/6.0; IEMobile/10.0; ARM; Touch; NOKIA; Lumia 520)",
	Viewport: cdp.Viewport{
		Width:  533,
		Height: 320,
	},
	DeviceScaleFactor: 1,
	IsMobile:          true,
	HasTouch:          true,
}
var NOKIAN9 = cdp.Device{
	UserAgent: "Mozilla/5.0 (MeeGo; NokiaN9) AppleWebKit/534.13 (KHTML, like Gecko) NokiaBrowser/8.5.0 Mobile Safari/534.13",
	Viewport: cdp.Viewport{
		Width:  480,
		Height: 854,
	},
	DeviceScaleFactor: 1,
	IsMobile:          true,
	HasTouch:          true,
}
var IPADPRO11LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPad; CPU OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  1194,
		Height: 834,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE8PLUS = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  414,
		Height: 736,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONEXLANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  812,
		Height: 375,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE11LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  800,
		Height: 364,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE6 = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/15.4 Mobile/15A372 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  375,
		Height: 667,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
var IPHONE13LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
	Viewport: cdp.Viewport{
		Width:  750,
		Height: 342,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var NEXUS6LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 7.1.1; Nexus 6 Build/N6F26U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  732,
		Height: 412,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var PIXEL4LANDSCAPE = cdp.Device{
	UserAgent: "Mozilla/5.0 (Linux; Android 10; Pixel 4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36",
	Viewport: cdp.Viewport{
		Width:  745,
		Height: 353,
	},
	DeviceScaleFactor: 3,
	IsMobile:          true,
	HasTouch:          true,
}
var MICROSOFTLUMIA550 = cdp.Device{
	UserAgent: "Mozilla/5.0 (Windows Phone 10.0; Android 4.2.1; Microsoft; Lumia 550) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4695.0 Mobile Safari/537.36 Edge/14.14263",
	Viewport: cdp.Viewport{
		Width:  640,
		Height: 360,
	},
	DeviceScaleFactor: 2,
	IsMobile:          true,
	HasTouch:          true,
}
