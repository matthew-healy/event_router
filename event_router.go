package event_router

import (
	"bytes"
	"context"
	"encoding/json"
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

type eventHandler struct {
	transport Transport
	handler   EventHandler
}

var handlerMap = map[EventKey]eventHandler{}

var mutex sync.RWMutex

func DefineJSONEvent(eventID EventKey, handler EventHandler) error {
	return DefineEvent(eventID, JSONTransport, handler)
}

func DefineUntransportedEvent(eventID EventKey, handler EventHandler) error {
	return DefineEvent(eventID, IdentityTransport, handler)
}

func DefineEvent(eventID EventKey, transport Transport, handler EventHandler) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := handlerMap[eventID]; exists {
		return ErrDuplicateRouteDef
	}
	handlerMap[eventID] = eventHandler{
		transport: transport,
		handler:   handler,
	}
	return nil
}

func HandleEvent(ctx context.Context, eventID EventKey, eventData any) error {
	mutex.RLock()
	handler, ok := handlerMap[eventID]
	mutex.RUnlock()
	if !ok {
		return ErrNoSuchEvent
	}

	parsedData, err := handler.transport(eventData, eventID.DataType())
	if err != nil {
		return err
	}

	return handler.handler(ctx, parsedData)
}

type Transport func(any, reflect.Type) (any, error)

func JSONTransport(data any, dataType reflect.Type) (any, error) {
	b, ok := data.([]byte)
	if !ok {
		return nil, errors.New("malformed input")
	}

	reader := bytes.NewReader(b)
	output := reflect.New(dataType)
	err := json.NewDecoder(reader).Decode(output.Interface())
	if err != nil {
		return nil, err
	}

	return output.Interface(), nil
}

func IdentityTransport(data any, dataType reflect.Type) (any, error) {
	if reflect.TypeOf(data) != dataType {
		return nil, ErrDataTypeMismatch
	}

	return data, nil
}
