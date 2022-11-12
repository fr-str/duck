package ws

import (
	"context"
	log "docker-project/logger"
	"docker-project/wsc"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/timoni-io/go-utils"
	"github.com/timoni-io/go-utils/math"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type websocketSession struct {
	Ctx                context.Context
	CancelSubscription context.CancelFunc
}

type Request struct {
	RequestID string
	Action    string
	Args      json.RawMessage
	Timeout   uint // in seconds

	ResultCh chan<- Response
	// action - request ctx, with timeout
	//
	// subscription - session ctx, canceled when subscription is changed or disconnected
	Ctx context.Context `json:"-"`
}

type Response struct {
	RequestID string
	Code      wsc.Type
	Data      any
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// TODO
	// if auth(r) {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	w.Write([]byte("Unauthorized"))
	// }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error during connection upgrade:", err)
		w.Write([]byte("Error Upgrading"))
		return
	}
	var service string = ""
	go func() {
		ip := utils.RequestIP(r)
		service = utils.DNSLookup(ip)
		if service == "" {
			service = ip
		}
		log.Info("New connection " + service)
	}()

	socketHandler(conn, &service)
}

func auth(r *http.Request) bool {
	return true
}

func socketHandler(conn *websocket.Conn, service *string) {
	// session ctx
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	webSess := &websocketSession{
		Ctx:                ctx,
		CancelSubscription: func() {},
	}

	// Response writer
	w := make(chan Response, 4)
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Debug("Closing writer", *service)
				return

			case res := <-w:
				if err := conn.WriteJSON(res); err != nil {
					log.Error(err)
				}
			}
		}
	}()

	// The event loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Error("Error during message reading:", err)
			}
			defer conn.Close()
			log.Info("Closing connection", *service)
			return // will cancel ctx
		}

		// decode
		r, code := decodeRequest(message)
		if code != wsc.OK {
			conn.WriteJSON(Error(r, code))
			continue
		}
		// handle
		go func(r *Request) {
			requestCtx, cancelRequest := context.WithTimeout(
				ctx,
				time.Duration(math.Clamp(r.Timeout, 5, 60))*time.Second,
			)
			defer cancelRequest()
			requestHandler(requestCtx, r, w, webSess)
		}(r)
	}
}

// requestHandler handles every request sent from client.
func requestHandler(ctx context.Context, r *Request, w chan<- Response, sess *websocketSession) {
	r.ResultCh = w
	prefix, _, _ := strings.Cut(r.Action, ".")

	// Get action handler
	handler, exists := handlers[prefix]
	if !exists {
		log.Debug(wsc.Action.String() + wsc.Missing.String())
		w <- Error(r, wsc.Action+wsc.Missing)
		return
	}

	// Decode action data
	action, err := handler.decode(r.Args)
	if err != nil {
		log.Error(wsc.Invalid.String() + wsc.ActionArgs.String())
		w <- Error(r, wsc.Invalid+wsc.ActionArgs, map[string]interface{}{
			"Error":      wsc.Invalid.String() + wsc.ActionArgs.String(),
			"FieldTypes": prettyPrintActionFieldsForFrontend(reflect.New(handler.action)),
		})
		return
	}

	// Exec action handler
	result := make(chan Response)

	go func() {
		defer panicHandler(r, result)

		switch handler.t {

		case actionT:
			// action has this request context
			r.Ctx = ctx
			result <- action.(HandlerI).Handle(r)

		case subscriptionT:
			// stop previous subscription
			sess.CancelSubscription()
			// subscription has new context with session context as parent, will get canceled when disconnected
			r.Ctx, sess.CancelSubscription = context.WithCancel(sess.Ctx)

			// Start subscription writer
			go action.(SubHandlerI).HandleSub(r, w)
			result <- Ok(r, "Subscription updated")
		}
	}()

	// Wait for result with timeout
	select {
	case res := <-result:
		w <- res
	case <-ctx.Done():
		log.Debug("request timeout")
		w <- Error(r, wsc.Timeout)
	}
}

func decodeRequest(dataIn []byte) (r *Request, code wsc.Type) {
	r = &Request{
		Args: dataIn,
	}
	err := json.Unmarshal(dataIn, r)
	if err != nil {
		log.Error(err)
		return r, wsc.Error + wsc.Decode
	}
	if r.RequestID == "" {
		log.Debug(wsc.Missing.String() + wsc.ReqID.String())
		return r, wsc.Missing + wsc.ReqID
	}
	if r.Action == "" {
		log.Debug(wsc.Missing.String() + wsc.Action.String())
		return r, wsc.Missing + wsc.Action
	}
	return r, wsc.OK
}

// decode returns action interface
func (h *actionHandler) decode(data json.RawMessage) (act any, err error) {
	// Create action struct pointer
	action := reflect.New(h.action).Interface()
	// Decode action
	if len(data) > 0 {
		err = json.Unmarshal(data, action)
		if err != nil {
			return nil, fmt.Errorf("invalid request: %s", err)
		}
	}
	// Extract action from pointer
	return reflect.ValueOf(action).Interface(), nil
}

func panicHandler(r *Request, w chan<- Response) {
	if err := recover(); err != nil {
		log.Error(err)
		w <- Error(r, wsc.InternalServerError)
	}
}

func Error(r *Request, code wsc.Type, d ...any) Response {
	var data any
	data = (code % 100).String() + (code - (code % 100)).String()
	if len(d) > 0 {
		data = d[0]
	}
	return Response{
		RequestID: r.RequestID,
		Code:      code,
		Data:      data,
	}
}

func GoError(r *Request, code wsc.Type, fns ...func() error) {
	go func() {
		for _, fn := range fns {
			err := fn()
			if err != nil {
				data := (code % 100).String() + (code ^ (code % 100)).String()
				r.ResultCh <- Response{
					RequestID: r.RequestID,
					Code:      code,
					Data:      fmt.Sprintf("%s: %s", data, err),
				}
			}
		}
	}()
}

func Ok(r *Request, data any) Response {
	return Response{
		RequestID: r.RequestID,
		Code:      wsc.OK,
		Data:      data,
	}
}

func Custom(r *Request, code wsc.Type, data any) Response {
	return Response{
		RequestID: r.RequestID,
		Code:      code,
		Data:      data,
	}
}

func Live(requestID string, data any) Response {
	return Response{
		RequestID: requestID,
		Code:      wsc.OK,
		Data:      data,
	}
}

func prettyPrintActionFieldsForFrontend(v reflect.Value) map[string]string {
	val := reflect.Indirect(v)
	m := map[string]string{}
	for i := 0; i < val.NumField(); i++ {
		m[val.Type().Field(i).Name] = val.Type().Field(i).Type.String()
	}
	return m
}
