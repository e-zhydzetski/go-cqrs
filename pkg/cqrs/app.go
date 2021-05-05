package cqrs

import (
	"context"
	"fmt"
	"github.com/e-zhydzetski/go-cqrs/pkg/es"
	"reflect"
	"time"
)

type CommandResult struct {
	AggregateID             string
	LastEventSequenceNumber es.StorePosition
	Events                  SavedEvents
}

type App interface {
	Command(ctx context.Context, command Command) (CommandResult, error)
	Query(ctx context.Context, query Query, minimalSeq es.StorePosition) (QueryResult, error)
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
	queryHandlers      map[reflect.Type]View
}

func NewSimpleApp(ctx context.Context, eventStore es.EventStore) *SimpleApp {
	return &SimpleApp{
		ctx:                ctx,
		eventStore:         eventStore,
		aggregateFactories: map[reflect.Type]AggregateFactory{},
		queryHandlers:      map[reflect.Type]View{},
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
	go func() {
		_ = s.eventStore.SubscribeOnEvents(s.ctx, view.EventFilter(), func(event *es.EventRecord) bool {
			view.Apply(event.Data, event.GlobalSequence)
			return true
		})
	}()

	for _, query := range view.QueryTypes() {
		s.queryHandlers[reflect.TypeOf(query)] = view
	}
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

func (s *SimpleApp) Command(ctx context.Context, command Command) (CommandResult, error) { // TODO maybe return full event with aggregate id
	aggregateID := command.AggregateID()
	commandType := reflect.TypeOf(command)
	factory, found := s.aggregateFactories[commandType]
	if !found {
		return CommandResult{}, fmt.Errorf("no aggregate found for command %T", command)
	}

RETRY:
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}

	aggregate := factory(aggregateID)

	var streamSeq uint32 = 0
	if aggregateID != "" { // special case if aggregate not exists before command
		aggregateEvents, err := s.eventStore.GetStreamEvents(ctx, AggregateToESStreamName(reflect.TypeOf(aggregate), aggregateID))
		if err != nil {
			return CommandResult{}, err
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
		return CommandResult{}, err
	}

	if aggregateID != "" {
		if aggregateID != aggregate.AggregateID() { // additional check that aggregate id not changed
			return CommandResult{}, fmt.Errorf("aggregate %T id was changed from %s to %s during command %T",
				aggregate, aggregateID, aggregate.AggregateID(), command)
		}
	}

	if aggregate.AggregateID() == "" {
		return CommandResult{}, fmt.Errorf("aggregate %T has no ID after command %T handling", aggregate, command)
	}

	streamName := AggregateToESStreamName(reflect.TypeOf(aggregate), aggregate.AggregateID())
	eventRecords := make([]*es.EventRecord, len(actions.pendingEvents))
	for i, event := range actions.pendingEvents {
		streamSeq++
		eventRecords[i] = &es.EventRecord{
			Stream:   streamName,
			Sequence: streamSeq,
			Type:     EventToESEventType(event),
			Data:     event,
		}
	}
	lastPublishedEventPosition, err := s.eventStore.PublishEvents(ctx, eventRecords...)
	if err != nil {
		if err == es.ErrStreamConcurrentModification {
			goto RETRY
		}
		return CommandResult{}, err
	}
	return CommandResult{
		AggregateID:             aggregate.AggregateID(),
		LastEventSequenceNumber: lastPublishedEventPosition,
		Events:                  actions.pendingEvents,
	}, nil
}

func (s *SimpleApp) Query(ctx context.Context, query Query, minimalSeq es.StorePosition) (QueryResult, error) {
	qt := reflect.TypeOf(query)
	view, found := s.queryHandlers[qt]
	if !found {
		return QueryResult{}, fmt.Errorf("no view found for query %T", query)
	}

	lastAppliedSeq := view.GetLastAppliedSeq()
	for lastAppliedSeq < minimalSeq {
		time.Sleep(100 * time.Millisecond)
		if err := ctx.Err(); err != nil {
			return QueryResult{}, err
		}
		lastAppliedSeq = view.GetLastAppliedSeq()
	}

	res, err := view.Query(query)
	if err != nil {
		return QueryResult{}, err
	}
	return QueryResult{
		Result: res,
		Seq:    lastAppliedSeq,
	}, nil
}
