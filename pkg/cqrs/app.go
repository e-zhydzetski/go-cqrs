package cqrs

import (
	"context"
	"errors"
	"reflect"
)

type App interface {
	Command(aggregateID string, command Command) error
	Query(query Query, result QueryResult) error
}

type MainApp interface {
	App
	BaseContext(ctx context.Context)
	RegisterAggregate(aggregateFactory AggregateFactory)
	RegisterView(view View)
}

type EventStore interface {
	GetEventsForAggregate(aggregateType reflect.Type, aggregateID string) ([]Event, error)
	PublishEventsForAggregate(aggregateType reflect.Type, aggregateID string, events []Event) error
}

type SimpleApp struct {
	ctx                context.Context
	eventStore         EventStore
	aggregateFactories map[reflect.Type]AggregateFactory
}

func (s SimpleApp) BaseContext(ctx context.Context) {
	s.ctx = ctx
}

func (s SimpleApp) RegisterAggregate(aggregateFactory AggregateFactory) {
	aggregate := aggregateFactory("")

	for _, commandType := range aggregate.CommandTypes() {
		s.aggregateFactories[reflect.TypeOf(commandType)] = aggregateFactory
	}
}

func (s SimpleApp) RegisterView(view View) {
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

func (s SimpleApp) Command(aggregateID string, command Command) error {
	commandType := reflect.TypeOf(command)
	factory, found := s.aggregateFactories[commandType]
	if !found {
		return errors.New("No Aggregate found for command type")
	}
	aggregate := factory(aggregateID)
	aggregateType := reflect.TypeOf(aggregate)
	aggregateEvents, err := s.eventStore.GetEventsForAggregate(aggregateType, aggregateID)
	if err != nil {
		return err
	}
	for _, event := range aggregateEvents {
		aggregate.Apply(event)
	}

	actions := &aggregateActions{
		aggregate: aggregate,
	}
	aggregate.Handle(command, actions)
	return s.eventStore.PublishEventsForAggregate(aggregateType, aggregateID, actions.pendingEvents)
}

func (s SimpleApp) Query(query Query, result QueryResult) error {
	panic("implement me")
}
