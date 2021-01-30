package cqrs

func NewInMemoryEventStore() EventStore {
	return &inMemoryEventStore{
		allEvents: map[AggregateType]map[string][]Event{},
	}
}

type inMemoryEventStore struct {
	allEvents map[AggregateType]map[string][]Event
}

func (i *inMemoryEventStore) GetEventsForAggregate(aggregateType AggregateType, aggregateID string) ([]Event, error) {
	aggregatesEvents, found := i.allEvents[aggregateType]
	if !found {
		return nil, nil
	}
	return aggregatesEvents[aggregateID], nil
}

func (i *inMemoryEventStore) PublishEventsForAggregate(aggregateType AggregateType, aggregateID string, events ...Event) error {
	aggregatesEvents, found := i.allEvents[aggregateType]
	if !found {
		aggregatesEvents = map[string][]Event{}
		i.allEvents[aggregateType] = aggregatesEvents
	}
	aggregatesEvents[aggregateID] = append(aggregatesEvents[aggregateID], events...)
	return nil
}
