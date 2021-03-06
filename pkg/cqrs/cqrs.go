package cqrs

import (
	"github.com/e-zhydzetski/go-cqrs/pkg/es"
	"reflect"
)

type Command interface {
	AggregateID() string // target aggregate ID
}

type AggrID string // helper struct for embedding into command structs

func (a AggrID) AggregateID() string {
	return string(a)
}

type Query interface{}
type QueryResult struct {
	Result interface{}
	Seq    es.StorePosition
}

type AggregateType reflect.Type

type Handler interface {
	CommandTypes() []Command
	// if command is invalid -> no state change -> no events emitted, simple return error
	Handle(command Command, actions AggregateActions) error
}

type Aggregate interface {
	Handler
	AggregateID() string
	Apply(event Event)
}

type AggregateActions interface {
	Emit(events ...Event)
}

type AggregateFactory func(id string) Aggregate

type View interface {
	Apply(event Event, globalSequence es.StorePosition)
	GetLastAppliedSeq() es.StorePosition
	QueryTypes() []Query
	Query(query Query) (interface{}, error)
}
