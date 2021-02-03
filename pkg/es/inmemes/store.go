package es

import (
	"context"
	"github.com/e-zhydzetski/go-commons/pkg/xchan"
	"github.com/e-zhydzetski/go-cqrs/pkg/es"
	"sync"
	"time"
)

func New() es.EventStore {
	return &inMemoryEventStore{
		seq:         es.StorePosBegin,
		events:      NewList(),
		streamIndex: map[string]*List{},
	}
}

type inMemoryEventStore struct {
	mx          sync.Mutex
	seq         es.StorePosition
	events      *List
	streamIndex map[string]*List

	subsMx        sync.RWMutex
	subscriptions []*xchan.Safe
}

func (s *inMemoryEventStore) PublishEvents(ctx context.Context, events ...*es.EventRecord) error {
	if len(events) == 0 {
		return nil // nothing to do
	}
	var stream string // should be the same for all events, TODO maybe change API to pass stream once
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

	if err := func() error {
		s.mx.Lock()
		defer s.mx.Unlock()

		// find/create index, detect current stream sequence
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

		// optimistic lock check
		for _, event := range events {
			if prevSeq+1 != event.Sequence {
				return es.ErrStreamConcurrentModification
			}
			prevSeq = event.Sequence
		}

		// set global sequence and store events
		for _, event := range events {
			s.seq++
			event.GlobalSequence = s.seq
			event.Timestamp = time.Now()
			h := s.events.PushToHead(event) // update main list
			si.PushToHead(h)                // update index
		}
		return nil
	}(); err != nil {
		return err
	}

	s.notifyAllAsync(events)

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

func checkFilter(filter es.EventFilter, event *es.EventRecord) bool {
	if filter.Stream != "" && filter.Stream != event.Stream {
		return false
	}
	if len(filter.Types) != 0 {
		contains := false
		for _, t := range filter.Types {
			if t == event.Type {
				contains = true
				break
			}
		}
		if !contains {
			return false
		}
	}
	if filter.Back {
		return event.GlobalSequence <= filter.Pos
	} else {
		return event.GlobalSequence >= filter.Pos
	}
}

func (s *inMemoryEventStore) notifyAllAsync(events []*es.EventRecord) {
	s.subsMx.RLock()
	defer s.subsMx.RUnlock()
	for _, sub := range s.subscriptions {
		func() {
			for _, e := range events {
				if !sub.Send(e) { // subscription closed, TODO remove closed subscriptions
					return
				}
			}
		}()
	}
}

func (s *inMemoryEventStore) SubscribeOnEvents(ctx context.Context, filter es.EventFilter, callbackFunc func(event *es.EventRecord)) error {
	// iterate by history events
	var err error
	s.events.Iterate(func(data interface{}) bool {
		err = ctx.Err()
		if err != nil {
			return false
		}
		e := data.(*es.EventRecord)
		if checkFilter(filter, e) {
			callbackFunc(e)
		}
		return true
	}, filter.Back)
	if err != nil {
		return err
	}

	// subscribe on new events
	if filter.Back {
		return nil // no new events will be in the past
	}
	// TODO fix lost events saved after history read and before subscription create !
	safeCh := xchan.MakeSafe()
	defer safeCh.Close()
	s.subsMx.Lock()
	s.subscriptions = append(s.subscriptions, safeCh)
	s.subsMx.Unlock()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case raw, ok := <-safeCh.Ch():
			if !ok {
				return ctx.Err()
			}
			e := raw.(*es.EventRecord)
			if checkFilter(filter, e) {
				callbackFunc(e)
			}
		}
	}
}
