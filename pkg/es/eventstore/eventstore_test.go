package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/EventStore/EventStore-Client-Go/messages"
	"github.com/EventStore/EventStore-Client-Go/position"
	"github.com/EventStore/EventStore-Client-Go/streamrevision"
	"github.com/gofrs/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"strings"
	"testing"
	"time"
)
import "github.com/EventStore/EventStore-Client-Go/client"

type TestEvent struct {
	Index       int       `json:"index"`
	PublishDate time.Time `json:"publish_date"`
}

func runEventStoreContainerAndGetConnectionString(ctx context.Context, t *testing.T) (string, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "eventstore/eventstore:20.10.2-buster-slim",
		ExposedPorts: []string{"2113/tcp"},
		Cmd:          []string{"--insecure"},
		WaitingFor:   wait.ForListeningPort("2113/tcp"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	ip, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "2113")
	if err != nil {
		t.Fatal(err)
	}

	uri := fmt.Sprintf(`esdb://%s:%s?tls=false`,
		ip,
		port.Port(),
	)

	return uri, func() {
		container.Terminate(ctx)
	}
}

func TestIntegration(t *testing.T) {
	ctx := context.Background()
	connectionString, stopContainer := runEventStoreContainerAndGetConnectionString(ctx, t)
	t.Cleanup(stopContainer)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	config, err := client.ParseConnectionString(connectionString)
	if err != nil {
		t.Fatal(err)
	}
	c, err := client.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	err = c.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	go func() {
		sub, err := c.SubscribeToAll(ctx, position.EndPosition, false,
			func(event messages.RecordedEvent) {
				if strings.HasPrefix(event.EventType, "$") {
					return
				}
				// log.Println("Event", event)
				var te TestEvent
				err = json.Unmarshal(event.Data, &te)
				if err != nil {
					t.Fatal(err)
				}
				log.Println(time.Since(te.PublishDate))

				if te.Index == 99 {
					cancel()
				}

			}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		err = sub.Start()
		if err != nil {
			t.Fatal(err)
		}
		defer sub.Stop()
		<-ctx.Done()
	}()

	for i := 0; i < 100; i++ {
		e := TestEvent{
			Index:       i,
			PublishDate: time.Now(),
		}
		eb, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		id, err := uuid.NewV4()
		if err != nil {
			t.Fatal(err)
		}
		result, err := c.AppendToStream(ctx, "test123", streamrevision.StreamRevisionAny, []messages.ProposedEvent{{
			EventID:      id,
			EventType:    "test-event",
			ContentType:  "application/json",
			Data:         eb,
			UserMetadata: nil,
		}})
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("%d: %+v", i, result)
	}
	<-ctx.Done()
}
