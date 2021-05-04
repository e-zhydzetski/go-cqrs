package inmemes

import (
	"context"
	"github.com/e-zhydzetski/go-cqrs/pkg/es"
	"gotest.tools/assert"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := sync.WaitGroup{}

	store := New()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			counter := 0
			_ = store.SubscribeOnEvents(ctx, es.FilterDefault(), func(event *es.EventRecord) bool {
				counter++
				return counter < 2000
			})
			wg.Done()
		}()
	}

	publishNewEvent := func(streamID int) {
		retry := 0
		for {
			events, _ := store.GetStreamEvents(ctx, "test_stream_"+strconv.Itoa(streamID))
			var streamSeq uint32 = 0
			if l := len(events); l > 0 {
				streamSeq = events[l-1].Sequence
			}
			_, err := store.PublishEvents(ctx, &es.EventRecord{
				Stream:   "test_stream_" + strconv.Itoa(streamID),
				Sequence: streamSeq + 1,
				Type:     "test_event",
				Data:     "test_data",
			})
			if err == es.ErrStreamConcurrentModification {
				retry++
				//fmt.Println("Retry", retry)
				continue
			}
			assert.NilError(t, err)
			break
		}
	}

	for i := 0; i < 10; i++ {
		for j := 0; j < 100; j++ {
			go publishNewEvent(i)
		}
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			counter := 0
			_ = store.SubscribeOnEvents(ctx, es.FilterDefault(), func(event *es.EventRecord) bool {
				counter++
				return counter < 2000
			})
			wg.Done()
		}()
	}

	for i := 0; i < 10; i++ {
		go func(streamNumber int) {
			for j := 0; j < 100; j++ {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				go publishNewEvent(streamNumber)
			}
		}(i)
	}

	wg.Wait()
}
