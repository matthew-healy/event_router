package event_router

import (
	"context"
	"errors"
	"reflect"
	"sync"
)

var (
	ErrDuplicateRouteDef = errors.New("duplicate route definition")
	ErrNoSuchEvent       = errors.New("no such event")
	ErrDataTypeMismatch  = errors.New("event data type mismatch")
)

type EventHandler func(context.Context, any) error

type EventKey interface {
	DataType() reflect.Type
}

var handlerMap = map[EventKey]EventHandler{}

var mutex sync.RWMutex

func AddRoute(eventID EventKey, handler EventHandler) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := handlerMap[eventID]; exists {
		return ErrDuplicateRouteDef
	}
	handlerMap[eventID] = handler
	return nil
}

// TODO: event processing middleware to handle conversions from transport types
// to domain types
func HandleEvent(ctx context.Context, eventID EventKey, eventData any) error {
	mutex.RLock()
	handler, ok := handlerMap[eventID]
	mutex.RUnlock()
	if !ok {
		return ErrNoSuchEvent
	}

	data := reflect.ValueOf(eventData)
	if data.Type() != eventID.DataType() {
		return ErrDataTypeMismatch
	}

	return handler(ctx, eventData)
}
