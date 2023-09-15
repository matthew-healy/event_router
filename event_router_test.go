package event_router

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type event string

func (e event) DataType() reflect.Type {
	return reflect.TypeOf("")
}

type event2 string

func (e event2) DataType() reflect.Type {
	return reflect.TypeOf("")
}

func TestHandle(t *testing.T) {
	evt := event("yes")
	expectedData := "the-event-data"
	ctx := context.Background()

	router := NewDependencyFreeRouter()
	err := router.DefineUntransportedEvent(evt, func(ctx context.Context, data any) error {
		data, ok := data.(string)
		if !ok {
			t.Error("received mistyped event")
		}
		if data != expectedData {
			t.Error("received wrong event data")
		}
		return nil
	})

	t.Run("does not return an error when defining a new route", func(t *testing.T) {
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
	})

	t.Run("errors when defining the same route twice", func(t *testing.T) {
		err := router.DefineUntransportedEvent(evt, func(ctx context.Context, data any) error {
			// noop
			return nil
		})
		expectError(t, ErrDuplicateRouteDef, err)
	})

	t.Run("routes handled events correctly", func(t *testing.T) {
		err := router.HandleEvent(ctx, evt, expectedData)
		if err != nil {
			t.Error("unexpected error")
		}
	})

	t.Run("returns error when passing wrong type to handler", func(t *testing.T) {
		err := router.HandleEvent(ctx, evt, 123)
		expectError(t, ErrDataTypeMismatch, err)
	})

	t.Run("returns error for unhandled event key", func(t *testing.T) {
		err := router.HandleEvent(ctx, event("no"), expectedData)
		expectError(t, ErrNoSuchEvent, err)
	})

	t.Run("different event types do not overlap", func(t *testing.T) {
		err := router.HandleEvent(ctx, event2("yes"), expectedData)
		expectError(t, ErrNoSuchEvent, err)
	})
}

func expectError(t *testing.T, expected error, actual error) {
	t.Helper()

	if actual == nil {
		t.Error("expected error but got none")
	}
	if !errors.Is(actual, expected) {
		t.Errorf("wrong error - expected %v but got %v", expected, actual)
	}
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Playlist struct {
	ID    string
	title string
}

type Payment struct {
	ID     string
	amount int
}

type domainEvent string

func (e domainEvent) DataType() reflect.Type {
	switch e {
	case "user.created":
		return reflect.TypeOf(User{})
	case "playlist.created":
		return reflect.TypeOf(Playlist{})
	default:
		return nil
	}
}

type externalEvent string

func (e externalEvent) DataType() reflect.Type {
	switch e {
	case "ext.payment.succeeded":
		return reflect.TypeOf(Payment{})
	default:
		return nil
	}
}

func TestMoreComplexExample(t *testing.T) {
	router := NewDependencyFreeRouter()
	router.DefineJSONEvent(domainEvent("user.created"), func(ctx context.Context, data any) error {
		d, ok := data.(*User)
		if !ok {
			t.Error("got wrong data type", data)
			t.Log("data type", reflect.TypeOf(data).String())
		}

		if d != nil && (d.ID != "usr_001" || d.Name != "Matthew") {
			t.Error("wrong user")
		}
		return nil
	})
	router.DefineJSONEvent(domainEvent("playlist.created"), func(ctx context.Context, data any) error {
		// TODO
		return nil
	})
	router.DefineJSONEvent(externalEvent("ext.payment.succeeded"), func(ctx context.Context, data any) error {
		// TODO
		return nil
	})

	userData := `
	  {
	    "id": "usr_001",
	    "name": "Matthew"
	  }
	`
	err := router.HandleEvent(context.Background(), domainEvent("user.created"), []byte(userData))
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

}
