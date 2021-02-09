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
	AggregateToESStreamName(aggregateType cqrs.AggregateType, aggregateID string) string
	EventToESEventType(event cqrs.Event) string
}

type AggregateEvent struct {
	AggregateType cqrs.AggregateType
	AggregateID   string
	Data          cqrs.Event
}

type TestCase struct {
	store       es.EventStore
	app         TestableApp
	givenEvents []AggregateEvent
	whenCommand cqrs.Command
}

func New(app TestableApp) *TestCase {
	return &TestCase{
		store: app.EventStore(),
		app:   app,
	}
}

func (c *TestCase) Given(events ...AggregateEvent) *TestCase {
	c.givenEvents = events
	return c
}

func (c *TestCase) When(command cqrs.Command) *TestCase {
	c.whenCommand = command
	return c
}

func (c *TestCase) Then(expectedEvents ...cqrs.Event) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		ctx := context.Background()

		for _, event := range c.givenEvents {
			streamName := c.app.AggregateToESStreamName(event.AggregateType, event.AggregateID)
			historyEvents, err := c.store.GetStreamEvents(ctx, streamName)
			if err != nil {
				t.Fatalf("unable to load history events: %v", err)
			}
			seq := uint32(0)
			if len(historyEvents) > 0 {
				seq = historyEvents[len(historyEvents)-1].Sequence
			}

			eventRecord := &es.EventRecord{
				Stream:   streamName,
				Sequence: seq + 1,
				Type:     c.app.EventToESEventType(event.Data),
				Data:     event.Data,
			}
			if err := c.store.PublishEvents(ctx, eventRecord); err != nil {
				t.Fatalf("unable to save given event: %v", err)
			}
		}

		// TODO get events from store by filter/subscription to test not only single aggregate handler results !!!
		events, err := c.app.Command(c.whenCommand)
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
