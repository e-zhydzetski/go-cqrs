package cqrs

import (
	"context"
	"fmt"
	"github.com/e-zhydzetski/go-cqrs/pkg/es"
	"reflect"
)

type App interface {
	Command(aggregateID string, command Command) (SavedEvents, error)
	Query(query Query, result QueryResult) error
}

type MainApp interface {
	App
	BaseContext(ctx context.Context)
	RegisterAggregate(aggregateFactory AggregateFactory)
	RegisterView(view View)
}

type SimpleApp struct {
	ctx                context.Context
	eventStore         es.EventStore
	aggregateFactories map[reflect.Type]AggregateFactory
}

func NewSimpleApp(ctx context.Context, eventStore es.EventStore) *SimpleApp {
	return &SimpleApp{
		ctx:                ctx,
		eventStore:         eventStore,
		aggregateFactories: map[reflect.Type]AggregateFactory{},
	}
}

func (s *SimpleApp) EventStore() es.EventStore {
	return s.eventStore
}

func (s *SimpleApp) BaseContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *SimpleApp) RegisterAggregate(aggregateFactory AggregateFactory) {
	aggregate := aggregateFactory("")

	for _, commandType := range aggregate.CommandTypes() {
		s.aggregateFactories[reflect.TypeOf(commandType)] = aggregateFactory
	}
}

func (s *SimpleApp) RegisterView(view View) {
	panic("implement me")
}

type aggregateActions struct {
	aggregate     Aggregate
	pendingEvents []Event
}

func (a *aggregateActions) Emit(events ...Event) {
	for _, event := range events {
		a.aggregate.Apply(event)
	}
	a.pendingEvents = append(a.pendingEvents, events...)
}

func (s *SimpleApp) aggregateToESStreamName(aggregate Aggregate) string {
	return reflect.TypeOf(aggregate).String() + "!" + aggregate.AggregateID() // TODO make stable aggregate type name
}

func (s *SimpleApp) eventToESEventType(event Event) string {
	return reflect.TypeOf(event).String() // TODO make stable event type name
}

func (s *SimpleApp) Command(aggregateID string, command Command) (SavedEvents, error) { // TODO maybe return full event with aggregate id
	commandType := reflect.TypeOf(command)
	factory, found := s.aggregateFactories[commandType]
	if !found {
		return nil, fmt.Errorf("no aggregate found for command %T", command)
	}

RETRY:
	aggregate := factory(aggregateID)

	var streamSeq uint32 = 0
	if aggregateID != "" { // special case if aggregate not exists before command
		aggregateEvents, err := s.eventStore.GetStreamEvents(s.ctx, s.aggregateToESStreamName(aggregate))
		if err != nil {
			return nil, err
		}
		for _, event := range aggregateEvents {
			aggregate.Apply(event.Data)
			streamSeq = event.Sequence
		}
	}

	actions := &aggregateActions{
		aggregate: aggregate,
	}
	err := aggregate.Handle(command, actions)
	if err != nil {
		return nil, err
	}

	if aggregateID != "" {
		if aggregateID != aggregate.AggregateID() { // additional check that aggregate id not changed
			return nil, fmt.Errorf("aggregate %T id was changed from %s to %s during command %T",
				aggregate, aggregateID, aggregate.AggregateID(), command)
		}
	}

	if aggregate.AggregateID() == "" {
		return nil, fmt.Errorf("aggregate %T has no ID after command %T handling", aggregate, command)
	}

	streamName := s.aggregateToESStreamName(aggregate)
	eventRecords := make([]*es.EventRecord, len(actions.pendingEvents))
	for i, event := range actions.pendingEvents {
		streamSeq++
		eventRecords[i] = &es.EventRecord{
			Stream:   streamName,
			Sequence: streamSeq,
			Type:     s.eventToESEventType(event),
			Data:     event,
		}
	}
	err = s.eventStore.PublishEvents(s.ctx, eventRecords...)
	if err != nil {
		if err == es.ErrStreamConcurrentModification {
			goto RETRY
		}
		return nil, err
	}
	return actions.pendingEvents, nil
}

func (s *SimpleApp) Query(query Query, result QueryResult) error {
	panic("implement me")
}
