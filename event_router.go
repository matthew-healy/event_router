package event_router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
)

// TODO: split into separate packages?

var (
	ErrDuplicateRouteDef = errors.New("duplicate route definition")
	ErrNoSuchEvent       = errors.New("no such event")
	ErrDataTypeMismatch  = errors.New("event data type mismatch")
)

type Router[Deps any] struct {
	routeMap     map[EventKey]eventHandler[Deps]
	dependencies Deps
	mutex        sync.RWMutex
}

func NewRouter[Deps any](dependencies Deps) *Router[Deps] {
	return &Router[Deps]{
		routeMap:     map[EventKey]eventHandler[Deps]{},
		dependencies: dependencies,
	}
}

type DependencyFreeRouter struct {
	r *Router[struct{}]
}

type DependencyFreeHandler func(context.Context, any) error

func NewDependencyFreeRouter() *DependencyFreeRouter {
	return &DependencyFreeRouter{
		r: NewRouter[struct{}](struct{}{}),
	}
}

type EventHandler[Deps any] func(context.Context, Deps, any) error

type EventKey interface {
	DataType() reflect.Type
}

type eventHandler[Deps any] struct {
	transport Transport
	handler   EventHandler[Deps]
}

func (r *Router[Deps]) DefineJSONEvent(eventID EventKey, handler EventHandler[Deps]) error {
	return r.DefineEvent(eventID, JSONTransport, handler)
}

func (d *DependencyFreeRouter) DefineJSONEvent(eventID EventKey, handler DependencyFreeHandler) error {
	return d.r.DefineJSONEvent(eventID, func(ctx context.Context, _ struct{}, data any) error {
		return handler(ctx, data)
	})
}

func (r *Router[Deps]) DefineUntransportedEvent(eventID EventKey, handler EventHandler[Deps]) error {
	return r.DefineEvent(eventID, IdentityTransport, handler)
}

func (d *DependencyFreeRouter) DefineUntransportedEvent(eventID EventKey, handler DependencyFreeHandler) error {
	return d.r.DefineUntransportedEvent(eventID, func(ctx context.Context, _ struct{}, data any) error {
		return handler(ctx, data)
	})
}

func (r *Router[Deps]) DefineEvent(eventID EventKey, transport Transport, handler EventHandler[Deps]) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.routeMap[eventID]; exists {
		return ErrDuplicateRouteDef
	}
	r.routeMap[eventID] = eventHandler[Deps]{
		transport: transport,
		handler:   handler,
	}
	return nil
}

func (r *Router[Deps]) HandleEvent(ctx context.Context, eventID EventKey, eventData any) error {
	r.mutex.RLock()
	handler, ok := r.routeMap[eventID]
	r.mutex.RUnlock()
	if !ok {
		return ErrNoSuchEvent
	}

	parsedData, err := handler.transport(eventData, eventID.DataType())
	if err != nil {
		return err
	}

	return handler.handler(ctx, r.dependencies, parsedData)
}

func (d *DependencyFreeRouter) HandleEvent(ctx context.Context, eventID EventKey, eventData any) error {
	return d.r.HandleEvent(ctx, eventID, eventData)
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
