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
	case *TestEvent:
		t.X = e.NewX
		fmt.Println(t.X)
	}
}

func (t *TestAggregate) CommandTypes() []Command {
	return []Command{&TestCommand{}}
}

func (t *TestAggregate) Handle(command Command, actions AggregateActions) {
	switch c := command.(type) {
	case *TestCommand:
		actions.Emit(&TestEvent{NewX: c.TargetX})
	}
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
	err := app.Command("1", &TestCommand{TargetX: 1})
	err = app.Command("1", &TestCommand{TargetX: 2})
	assert.NilError(t, err)
}
