package cdp

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"gitee.com/baixudong/gospider/db"
	"gitee.com/baixudong/gospider/requests"
	"gitee.com/baixudong/gospider/websocket"

	"go.uber.org/atomic"
)

type commend struct {
	Id     int64          `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}
type event struct {
	Ctx      context.Context
	Cnl      context.CancelFunc
	RecvData chan RecvData
}
type RecvData struct {
	Id     int64          `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
	Result map[string]any `json:"result"`
}

type WebSock struct {
	option       WebSockOption
	db           *db.Client[FulData]
	ids          sync.Map
	conn         *websocket.Conn
	ctx          context.Context
	cnl          context.CancelCauseFunc
	id           atomic.Int64
	RequestFunc  func(context.Context, *Route)
	ResponseFunc func(context.Context, *Route)
	reqCli       *requests.Client
	onEvents     sync.Map
}

type DataEntrie struct {
	Bytes string `json:"bytes"`
}

func (obj *WebSock) Done() <-chan struct{} {
	return obj.ctx.Done()
}
func (obj *WebSock) routeMain(ctx context.Context, recvData RecvData) {
	routeData := RouteData{}
	temData, err := json.Marshal(recvData.Params)
	if err == nil && json.Unmarshal(temData, &routeData) == nil {
		route := &Route{
			webSock:  obj,
			recvData: routeData,
		}
		if route.IsResponse() {
			if obj.ResponseFunc != nil {
				obj.ResponseFunc(ctx, route)
				if !route.isRoute {
					if obj.option.IsReplaceRequest {
						route.RequestContinue(ctx)
					} else {
						route.Continue(ctx)
					}
				}
			} else {
				route.Continue(ctx)
			}
		} else {
			if obj.RequestFunc != nil {
				obj.RequestFunc(ctx, route)
				if !route.isRoute {
					if obj.option.IsReplaceRequest {
						route.RequestContinue(ctx)
					} else {
						route.Continue(ctx)
					}
				}
			} else {
				route.Continue(ctx)
			}
		}
	}
}

func (obj *WebSock) recv(ctx context.Context, rd RecvData) error {
	defer recover()
	cmdDataAny, ok := obj.ids.LoadAndDelete(rd.Id)
	if ok {
		cmdData := cmdDataAny.(*event)
		select {
		case <-obj.Done():
			return errors.New("websocks closed")
		case <-ctx.Done():
			return ctx.Err()
		case <-cmdData.Ctx.Done():
		case cmdData.RecvData <- rd:
		}
	}
	methodFuncAny, ok := obj.onEvents.Load(rd.Method)
	if ok && methodFuncAny != nil {
		methodFuncAny.(func(ctx context.Context, rd RecvData))(ctx, rd)
	}
	return nil
}
func (obj *WebSock) recvMain() (err error) {
	defer obj.Close(err)
	for {
		select {
		case <-obj.ctx.Done():
			return obj.ctx.Err()
		default:
			rd := RecvData{}
			if err := obj.conn.RecvJson(obj.ctx, &rd); err != nil {
				return err
			}
			if rd.Id == 0 {
				rd.Id = obj.id.Add(1)
			}
			go obj.recv(obj.ctx, rd)
		}
	}
}

type WebSockOption struct {
	Proxy            string
	DataCache        bool //开启数据缓存
	IsReplaceRequest bool
}

func NewWebSock(preCtx context.Context, globalReqCli *requests.Client, ws string, option WebSockOption, db *db.Client[FulData]) (*WebSock, error) {
	response, err := globalReqCli.Request(preCtx, "get", ws, requests.RequestOption{DisProxy: true})
	if err != nil {
		return nil, err
	}
	response.WebSocket().SetReadLimit(1024 * 1024 * 1024) //1G
	cli := &WebSock{
		conn:   response.WebSocket(),
		db:     db,
		reqCli: globalReqCli,
		option: option,
	}
	cli.ctx, cli.cnl = context.WithCancelCause(preCtx)
	go cli.recvMain()
	cli.AddEvent("Fetch.requestPaused", cli.routeMain)
	return cli, err
}
func (obj *WebSock) AddEvent(method string, fun func(ctx context.Context, rd RecvData)) {
	obj.onEvents.Store(method, fun)
}
func (obj *WebSock) DelEvent(method string) {
	obj.onEvents.Delete(method)
}
func (obj *WebSock) Close(err error) error {
	obj.cnl(err)
	return obj.conn.Close("close")
}

func (obj *WebSock) regId(preCtx context.Context, ids ...int64) *event {
	data := new(event)
	data.Ctx, data.Cnl = context.WithCancel(preCtx)
	data.RecvData = make(chan RecvData)
	for _, id := range ids {
		obj.ids.Store(id, data)
	}
	return data
}
func (obj *WebSock) send(preCtx context.Context, cmd commend) (RecvData, error) {
	var cnl context.CancelFunc
	var ctx context.Context
	if preCtx == nil {
		ctx, cnl = context.WithTimeout(obj.ctx, time.Second*60)
	} else {
		ctx, cnl = context.WithTimeout(preCtx, time.Second*60)
	}
	defer cnl()
	select {
	case <-obj.Done():
		return RecvData{}, context.Cause(obj.ctx)
	case <-ctx.Done():
		return RecvData{}, obj.ctx.Err()
	default:
		cmd.Id = obj.id.Add(1)
		idEvent := obj.regId(ctx, cmd.Id)
		defer idEvent.Cnl()
		if err := obj.conn.SendJson(ctx, cmd); err != nil {
			return RecvData{}, err
		}
		select {
		case <-obj.Done():
			return RecvData{}, context.Cause(obj.ctx)
		case <-ctx.Done():
			return RecvData{}, ctx.Err()
		case idRecvData := <-idEvent.RecvData:
			return idRecvData, nil
		}
	}
}
