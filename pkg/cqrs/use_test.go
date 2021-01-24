package cqrs

import (
	"context"
	"fmt"
	"gotest.tools/assert"
	"reflect"
	"testing"
)

type TestCommand struct {
	TargetX int
}

type TestCreatedEvent struct {
	ID string
}

type TestEvent struct {
	NewX int
}

type TestAggregate struct {
	ID string
	X  int
}

func (t *TestAggregate) AggregateID() string {
	return t.ID
}

func (t *TestAggregate) Apply(event Event) {
	switch e := event.(type) {
	case *TestCreatedEvent:
		t.ID = e.ID
		fmt.Println("Created with ID=", t.ID)
	case *TestEvent:
		t.X = e.NewX
		fmt.Println(t.X)
	}
}

func (t *TestAggregate) CommandTypes() []Command {
	return []Command{&TestCommand{}}
}

func (t *TestAggregate) Handle(command Command, actions AggregateActions) error {
	switch c := command.(type) {
	case *TestCommand:
		if t.ID == "" {
			actions.Emit(&TestCreatedEvent{ID: "xyz"})
		}
		actions.Emit(&TestEvent{NewX: c.TargetX})
	}
	return nil
}

func NewTestAggregate(id string) Aggregate {
	return &TestAggregate{
		ID: id,
		X:  0,
	}
}

func TestUsage(t *testing.T) {
	app := SimpleApp{
		ctx:                context.Background(),
		eventStore:         NewInMemoryEventStore(),
		aggregateFactories: map[reflect.Type]AggregateFactory{},
	}
	app.RegisterAggregate(NewTestAggregate)
	events, err := app.Command("", &TestCommand{TargetX: 1})
	assert.NilError(t, err)
	var createdEvent TestCreatedEvent
	assert.Assert(t, events.Get(&createdEvent))
	assert.Assert(t, createdEvent.ID == "xyz")
	fmt.Println(events)
	events, err = app.Command("xyz", &TestCommand{TargetX: 2})
	fmt.Println(events)
	assert.NilError(t, err)
}
