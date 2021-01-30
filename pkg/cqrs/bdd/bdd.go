package bdd

import (
	"github.com/e-zhydzetski/go-cqrs/pkg/cqrs"
	"gotest.tools/assert"
	"reflect"
	"testing"
)

type TestableApp interface {
	cqrs.App
	EventStore() cqrs.EventStore
}

type TestCase struct {
	store             cqrs.EventStore
	app               cqrs.App
	givenEvents       []cqrs.EventRecord
	targetAggregateID string
	whenCommand       cqrs.Command
}

func New(app TestableApp) *TestCase {
	return &TestCase{
		store: app.EventStore(),
		app:   app,
	}
}

func (c *TestCase) Given(events ...cqrs.EventRecord) *TestCase {
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

		for _, er := range c.givenEvents {
			if err := c.store.PublishEventsForAggregate(er.AggregateType, er.AggregateID, er.Data); err != nil {
				t.Fatalf("unable to save given event: %v", err)
			}
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
