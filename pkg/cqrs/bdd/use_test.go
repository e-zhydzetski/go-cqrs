package bdd

import (
	"context"
	"fmt"
	"github.com/e-zhydzetski/go-cqrs/pkg/cqrs"
	"github.com/e-zhydzetski/go-cqrs/pkg/es/inmemes"
	"reflect"
	"testing"
)

type TestCommand struct {
	cqrs.AggrID
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

func (t *TestAggregate) Apply(event cqrs.Event) {
	switch e := event.(type) {
	case *TestCreatedEvent:
		t.ID = e.ID
		fmt.Println("Created with ID=", t.ID)
	case *TestEvent:
		t.X = e.NewX
		fmt.Println(t.X)
	}
}

func (t *TestAggregate) CommandTypes() []cqrs.Command {
	return []cqrs.Command{&TestCommand{}}
}

func (t *TestAggregate) Handle(command cqrs.Command, actions cqrs.AggregateActions) error {
	switch c := command.(type) {
	case *TestCommand:
		if t.ID == "" {
			actions.Emit(&TestCreatedEvent{ID: "xyz"})
		}
		actions.Emit(&TestEvent{NewX: c.TargetX})
	}
	return nil
}

func NewTestAggregate(id string) cqrs.Aggregate {
	return &TestAggregate{
		ID: id,
		X:  0,
	}
}

func TestUsage(t *testing.T) {
	app := cqrs.NewSimpleApp(context.Background(), inmemes.New())
	app.RegisterAggregate(NewTestAggregate)

	testCase := New(app)
	t.Run("simple test", testCase.
		When(&TestCommand{TargetX: 1}).
		Then(&TestCreatedEvent{
			ID: "xyz",
		}),
	)

	t.Run("simple test2", testCase.
		Given(AggregateEvent{
			AggregateType: reflect.TypeOf(&TestAggregate{}),
			AggregateID:   "xyz",
			Data: &TestCreatedEvent{
				ID: "xyz",
			},
		}).
		When(&TestCommand{
			AggrID:  "xyz",
			TargetX: 2,
		}).
		Then(&TestEvent{
			NewX: 2,
		}),
	)
}
