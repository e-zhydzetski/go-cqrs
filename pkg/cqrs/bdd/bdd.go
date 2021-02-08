package bdd

import (
	"context"
	"github.com/e-zhydzetski/go-cqrs/pkg/cqrs"
	"github.com/e-zhydzetski/go-cqrs/pkg/es"
	"gotest.tools/assert"
	"reflect"
	"testing"
)

type TestableApp interface {
	cqrs.App
	EventStore() es.EventStore
}

type TestCase struct {
	store             es.EventStore
	app               cqrs.App
	givenEvents       []*es.EventRecord
	targetAggregateID string
	whenCommand       cqrs.Command
}

func New(app TestableApp) *TestCase {
	return &TestCase{
		store: app.EventStore(),
		app:   app,
	}
}

func (c *TestCase) Given(events ...*es.EventRecord) *TestCase {
	c.givenEvents = events
	return c
}

func (c *TestCase) When(targetAggregateID string, command cqrs.Command) *TestCase {
	c.targetAggregateID = targetAggregateID
	c.whenCommand = command
	return c
}

func (c *TestCase) Then(expectedEvents ...cqrs.Event) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		if err := c.store.PublishEvents(context.Background(), c.givenEvents...); err != nil {
			t.Fatalf("unable to save given event: %v", err)
		}

		// TODO get events from store by filter/subscription to test not only single aggregate handler results !!!
		events, err := c.app.Command(c.targetAggregateID, c.whenCommand)
		if err != nil {
			t.Fatalf("command handler error: %v", err)
		}

		for _, expectedEvent := range expectedEvents {
			actualEvent := reflect.New(reflect.TypeOf(expectedEvent).Elem()).Interface() // TODO refactor magic
			if !events.Get(actualEvent) {
				t.Fatalf("not found expected event: %T", actualEvent)
			}
			assert.DeepEqual(t, expectedEvent, actualEvent)
		}
	}
}
