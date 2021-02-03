package es

import (
	"context"
	"github.com/e-zhydzetski/go-cqrs/pkg/es"
	"gotest.tools/assert"
	"strconv"
	"testing"
)

func TestLoad(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	step1 := make(chan struct{})
	step2 := make(chan struct{})

	store := New()

	for k := 0; k < 10; k++ {
		for i := 1; i <= 10; i++ {
			err := store.PublishEvents(ctx, &es.EventRecord{
				Stream:   "test_stream_" + strconv.Itoa(k),
				Sequence: uint32(i),
				Type:     "test_event",
				Data:     "test_data",
			})
			assert.NilError(t, err)
		}
		events, _ := store.GetStreamEvents(ctx, "test_stream_"+strconv.Itoa(k))
		assert.Equal(t, len(events), 10)
	}

	go func() {
		counter := 0
		_ = store.SubscribeOnEvents(ctx, es.FilterDefault().WithTypes("test_event"), func(event *es.EventRecord) bool {
			counter++
			if counter == 10 {
				close(step1)
			}
			return counter < 200
		})
		close(step2)
	}()

	for k := 10; k < 20; k++ {
		go func(streamNumber int) {
			<-step1
			for i := 1; i <= 10; i++ {
				err := store.PublishEvents(ctx, &es.EventRecord{
					Stream:   "test_stream_" + strconv.Itoa(streamNumber),
					Sequence: uint32(i),
					Type:     "test_event",
					Data:     "test_data",
				})
				assert.NilError(t, err)
			}
			events, _ := store.GetStreamEvents(ctx, "test_stream_"+strconv.Itoa(streamNumber))
			assert.Equal(t, len(events), 10)
		}(k)
	}

	<-step2
}
