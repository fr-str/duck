package ws

import (
	log "docker-project/logger"
	"fmt"

	"reflect"
)

type HandlerI interface {
	Handle(*Request) Response
}

// Subscription handler.
// Request have context which sends signal Done when subscription is canceled.
type SubHandlerI interface {
	HandleSub(*Request, chan<- Response)
}

type actionHandler struct {
	t      handlerType
	action reflect.Type
}

type handlerType byte

const (
	actionT handlerType = 1 << iota
	subscriptionT
)

var handlers = map[string]actionHandler{}

func RegisterAction[T any](name string) {
	a := (*T)(nil)

	if _, ok := any(a).(HandlerI); !ok {
		log.Fatal(fmt.Sprintf("invalid action type: %s %T", name, a))
	}

	if _, ok := handlers[name]; ok {
		log.Fatal(fmt.Sprintf("handler `%s` already registered", name))
	}

	handlers[name] = actionHandler{
		t:      actionT,
		action: reflect.TypeOf(a).Elem(),
	}
}

func RegisterSubscription[T any](name string) {
	a := (*T)(nil)

	if _, ok := any(a).(SubHandlerI); !ok {
		log.Fatal(fmt.Sprintf("invalid action type: %s %T", name, a))
	}

	if _, ok := handlers[name]; ok {
		log.Fatal(fmt.Sprintf("handler `%s` already registered", name))
	}

	handlers[name] = actionHandler{
		t:      subscriptionT,
		action: reflect.TypeOf(a).Elem(),
	}
}
