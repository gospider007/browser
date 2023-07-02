package cdp

import "context"

func (obj *WebSock) EmulationSetUserAgentOverride(preCtx context.Context, userAgent string) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Emulation.setUserAgentOverride",
		Params: map[string]any{
			"userAgent": userAgent,
		},
	})
}

type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}
type Device struct {
	UserAgent         string   `json:"user_agent"`
	Viewport          Viewport `json:"viewport"`
	DeviceScaleFactor float64  `json:"device_scale_factor"`
	IsMobile          bool     `json:"is_mobile"`
	HasTouch          bool     `json:"has_touch"`
}

func (obj *WebSock) EmulationSetDeviceMetricsOverride(preCtx context.Context, device Device) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Emulation.setDeviceMetricsOverride",
		Params: map[string]any{
			"width":             device.Viewport.Width,
			"height":            device.Viewport.Height,
			"deviceScaleFactor": device.DeviceScaleFactor,
			"mobile":            device.IsMobile,
		},
	})
}
func (obj *WebSock) EmulationSetTouchEmulationEnabled(preCtx context.Context, hasTouch bool) (RecvData, error) {
	return obj.send(preCtx, commend{
		Method: "Emulation.setTouchEmulationEnabled",
		Params: map[string]any{
			"enabled": hasTouch,
		},
	})
}
