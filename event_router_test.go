package event_router

import (
	"reflect"
	"testing"
)

type event string

func (e event) DataType() reflect.Type {
	return reflect.TypeOf("")
}

func TestHandle(t *testing.T) {
	evt := event("yes")

	expectedData := "the-event-data"

	AddRoute(evt, func(data any) error {
		data, ok := data.(string)
		if !ok {
			t.Fatal("received mistyped event")
		}
		if data != expectedData {
			t.Fatalf("received wrong event data")
		}
		return nil
	})

	err := HandleEvent(evt, expectedData)
	if err != nil {
		t.Fatal("unexpected error")
	}

	err = HandleEvent(evt, 123)
	if err == nil {
		t.Fatal("no error for wrong type")
	}

	err = HandleEvent(event("no"), expectedData)
	if err == nil {
		t.Fatal("no error for unhandled event")
	}
}
