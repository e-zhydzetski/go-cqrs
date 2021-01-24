package cqrs

import (
	"fmt"
	"reflect"
)

type Event interface{}

type SavedEvents []Event

func (se SavedEvents) Get(event interface{}) bool {
	eventType := reflect.TypeOf(event)
	for _, savedEvent := range se {
		if eventType == reflect.TypeOf(savedEvent) {
			eventVal := reflect.ValueOf(event)
			savedEventVal := reflect.ValueOf(savedEvent)

			if savedEventVal.Kind() == reflect.Ptr {
				savedEventVal = savedEventVal.Elem()
			}

			if savedEventVal.Type().AssignableTo(eventVal.Type().Elem()) {
				eventVal.Elem().Set(savedEventVal)
				return true
			}
		}
	}
	return false
}

func (se SavedEvents) String() string {
	str := ""
	for i, event := range se {
		if i > 0 {
			str = str + "\n"
		}
		str = str + fmt.Sprintf("%d: %#v", i, event)
	}
	return str
}
