package es

import (
	"context"
	"github.com/e-zhydzetski/go-cqrs/pkg/es"
	"sync"
	"time"
)

func New() es.EventStore {
	return &inMemoryEventStore{
		events:      NewList(),
		streamIndex: map[string]*List{},
	}
}

type inMemoryEventStore struct {
	mx          sync.Mutex
	events      *List
	streamIndex map[string]*List
	// TODO subscribers
}

func (s *inMemoryEventStore) PublishEvents(ctx context.Context, events ...*es.EventRecord) error {
	if len(events) == 0 {
		return nil // nothing to do
	}
	var stream string // should be the same for all events
	for i, event := range events {
		if event == nil {
			return es.ErrInvalidEvent
		}
		if i == 0 {
			stream = event.Stream
			continue
		}
		if event.Stream != stream {
			return es.ErrInvalidEvent
		}
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	var prevSeq uint32 = 0
	si, exists := s.streamIndex[stream]
	if exists {
		lastElement := si.GetLastElement()
		if lastElement != nil {
			prevSeq = lastElement.(*Node).Data.(*es.EventRecord).Sequence
		}
	} else {
		si = NewList()
		s.streamIndex[stream] = si
	}

	for _, event := range events {
		if prevSeq+1 != event.Sequence {
			return es.ErrStreamConcurrentModification
		}
		event.Timestamp = time.Now()
		prevSeq = event.Sequence
	}

	for _, event := range events {
		h := s.events.PushToHead(event) // update main list
		si.PushToHead(h)                // update index
	}

	// TODO maybe release mutex before
	s.notifyAllAsync(events...)

	return nil
}

func (s *inMemoryEventStore) GetStreamEvents(ctx context.Context, stream string) ([]*es.EventRecord, error) {
	var events []*es.EventRecord
	si, exists := s.streamIndex[stream]
	if !exists {
		return events, nil
	}
	si.Iterate(func(data interface{}) bool {
		events = append(events, data.(*Node).Data.(*es.EventRecord))
		return true
	}, false)
	return events, nil
}

func (s *inMemoryEventStore) notifyAllAsync(events ...*es.EventRecord) {

}

func (s *inMemoryEventStore) SubscribeOnEvents(ctx context.Context, filter es.EventFilter) (<-chan *es.EventRecord, func(), error) {
	if filter.Stream != "" {
		// use stream index list
	}
	// TODO: read and publish all current events until head
	// TODO: publish all new events by notifyAllAsync

	return nil, nil, nil // TODO: return safe channel, unsubscribe func, error
}
