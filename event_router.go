package event_router

import (
	"errors"
	"reflect"
)

type EventHandler func(any) error

type EventKey interface {
	DataType() reflect.Type
}

// TODO: read/write lock
var handlerMap = map[EventKey]EventHandler{}

// TODO: error on existing route override
func AddRoute(eventID EventKey, handler EventHandler) {
	handlerMap[eventID] = handler
}

// TODO: event processing middleware to handle conversions from transport types
// to domain types
func HandleEvent(eventID EventKey, eventData any) error {
	handler, ok := handlerMap[eventID]
	if !ok {
		return errors.New("unhandled event")
	}

	data := reflect.ValueOf(eventData)
	if data.Type() != eventID.DataType() {
		return errors.New("type mismatch")
	}

	return handler(eventData)
}
