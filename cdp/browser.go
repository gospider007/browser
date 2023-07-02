package cdp

func (obj *WebSock) BrowserClose() error {
	_, err := obj.send(obj.ctx, commend{
		Method: "Browser.close",
	})
	return err
}
