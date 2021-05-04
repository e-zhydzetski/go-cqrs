package inmemes

import (
	"context"
	"github.com/e-zhydzetski/go-commons/pkg/xchan"
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
	mx          sync.RWMutex
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
	seq := es.StorePosition(s.events.Len())
	for _, event := range events {
		seq++
		event.GlobalSequence = seq
		event.Timestamp = time.Now()
		h := s.events.PushToHead(event) // update main list
		si.PushToHead(h)                // update index
	}

	s.notifyAll(events)

	return nil
}

func (s *inMemoryEventStore) GetStreamEvents(ctx context.Context, stream string) ([]*es.EventRecord, error) {
	var events []*es.EventRecord
	s.mx.RLock()
	si, exists := s.streamIndex[stream]
	s.mx.RUnlock()
	if !exists {
		return events, nil
	}
	si.Iterate(func(data interface{}) bool {
		events = append(events, data.(*Node).Data.(*es.EventRecord))
		return true
	}, nil)
	return events, nil
}

func (s *inMemoryEventStore) notifyAll(events []*es.EventRecord) {
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

func checkFilter(filter *es.EventFilter, event *es.EventRecord) bool {
	if filter.Stream != "" && filter.Stream != event.Stream {
		return false
	}
	//if filter.Back {
	//	return event.GlobalSequence <= filter.Pos
	//} else {
	return event.GlobalSequence >= filter.Pos
	//}
}

func (s *inMemoryEventStore) SubscribeOnEvents(ctx context.Context, filter *es.EventFilter, callbackFunc func(event *es.EventRecord) bool) error {
	// prepare for future subscribe
	safeCh := xchan.MakeSafe()
	defer safeCh.Close()

	// iterate by history events
	var err error
	var cont = true
	s.events.Iterate(func(data interface{}) bool {
		err = ctx.Err()
		if err != nil {
			return false
		}
		e := data.(*es.EventRecord)
		if checkFilter(filter, e) {
			//fmt.Println("History:", e.GlobalSequence)
			cont = callbackFunc(e)
			if !cont {
				return false
			}
		}
		return true
	}, func() { // subscribe at the end of list
		s.subsMx.Lock()
		s.subscriptions = append(s.subscriptions, safeCh)
		s.subsMx.Unlock()
	})
	if err != nil {
		return err
	}
	if !cont {
		return nil
	}

	// subscription
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
				//fmt.Println("Subscription:", e.GlobalSequence)
				if !callbackFunc(e) {
					return nil
				}
			}
		}
	}
}
