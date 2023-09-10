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
	return DefineEvent(eventID, JSONTransport(), handler)
}

func DefineUntransportedEvent(eventID EventKey, handler EventHandler) error {
	return DefineEvent(eventID, NoTransport(), handler)
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

// TODO: event processing middleware to handle conversions from transport types
// to domain types
func HandleEvent(ctx context.Context, eventID EventKey, eventData any) error {
	mutex.RLock()
	handler, ok := handlerMap[eventID]
	mutex.RUnlock()
	if !ok {
		return ErrNoSuchEvent
	}

	parsedData, err := handler.transport.Decode(eventData, eventID.DataType())
	if err != nil {
		return err
	}

	return handler.handler(ctx, parsedData)
}

type Transport interface {
	Decode(data any, dataType reflect.Type) (any, error)
}

type jsonTransport struct{}

func (t *jsonTransport) Decode(data any, dataType reflect.Type) (any, error) {
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

func JSONTransport() Transport {
	return &jsonTransport{}
}

func NoTransport() Transport {
	return &noTransport{}
}

type noTransport struct{}

func (t *noTransport) Decode(data any, dataType reflect.Type) (any, error) {
	if reflect.TypeOf(data) != dataType {
		return nil, ErrDataTypeMismatch
	}

	return data, nil
}
