package cqrs

import (
	"context"
	"fmt"
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

type EventStore interface {
	GetEventsForAggregate(aggregateType AggregateType, aggregateID string) ([]Event, error)
	PublishEventsForAggregate(aggregateType AggregateType, aggregateID string, events ...Event) error
}

// Event wrapper with metadata
type EventRecord struct {
	AggregateType AggregateType
	AggregateID   string
	Data          Event
}

type SimpleApp struct {
	ctx                context.Context
	eventStore         EventStore
	aggregateFactories map[reflect.Type]AggregateFactory
}

func NewSimpleApp(ctx context.Context, eventStore EventStore) *SimpleApp {
	return &SimpleApp{
		ctx:                ctx,
		eventStore:         eventStore,
		aggregateFactories: map[reflect.Type]AggregateFactory{},
	}
}

func (s *SimpleApp) EventStore() EventStore {
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

func (s *SimpleApp) Command(aggregateID string, command Command) (SavedEvents, error) { // TODO maybe return full event with aggregate id
	commandType := reflect.TypeOf(command)
	factory, found := s.aggregateFactories[commandType]
	if !found {
		return nil, fmt.Errorf("no aggregate found for command %T", command)
	}
	aggregate := factory(aggregateID)
	aggregateType := reflect.TypeOf(aggregate)

	if aggregateID != "" { // special case if aggregate not exists before command
		aggregateEvents, err := s.eventStore.GetEventsForAggregate(aggregateType, aggregateID)
		if err != nil {
			return nil, err
		}
		for _, event := range aggregateEvents {
			aggregate.Apply(event)
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

	err = s.eventStore.PublishEventsForAggregate(aggregateType, aggregate.AggregateID(), actions.pendingEvents)
	if err != nil {
		return nil, err
	}
	return actions.pendingEvents, nil
}

func (s *SimpleApp) Query(query Query, result QueryResult) error {
	panic("implement me")
}
