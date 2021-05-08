package bdd

import (
	"context"
	"github.com/e-zhydzetski/go-cqrs/pkg/cqrs"
	"github.com/e-zhydzetski/go-cqrs/pkg/es"
	"gotest.tools/assert"
	"reflect"
	"testing"
)

type AggregateEvent struct {
	AggregateType cqrs.AggregateType
	AggregateID   string
	Data          cqrs.Event
}

type TestCase struct {
	store       es.EventStore
	app         cqrs.App
	givenEvents []AggregateEvent
	whenCommand cqrs.Command
}

func New(app cqrs.App) *TestCase {
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
			streamName := cqrs.AggregateToESStreamName(event.AggregateType, event.AggregateID)
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
				Type:     cqrs.EventToESEventType(event.Data),
				Data:     event.Data,
			}
			if _, err := c.store.PublishEvents(ctx, eventRecord); err != nil {
				t.Fatalf("unable to save given event: %v", err)
			}
		}

		// TODO get events from store by filter/subscription to test not only single aggregate handler results !!!
		result, err := c.app.Command(ctx, c.whenCommand)
		if err != nil {
			t.Fatalf("command handler error: %v", err)
		}

		for _, expectedEvent := range expectedEvents {
			actualEvent := reflect.New(reflect.TypeOf(expectedEvent).Elem()).Interface() // TODO refactor magic
			if !result.Events.Get(actualEvent) {
				t.Fatalf("not found expected event: %T", actualEvent)
			}
			assert.DeepEqual(t, expectedEvent, actualEvent)
		}
	}
}
