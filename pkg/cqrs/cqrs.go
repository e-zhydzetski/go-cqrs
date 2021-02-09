package cqrs

import "reflect"

type Command interface {
	AggregateID() string // target aggregate ID
}

type AggregateID string // helper struct for embedding into command structs

func (a AggregateID) AggregateID() string {
	return string(a)
}

type Query interface{}
type QueryResult interface{}

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
	EventTypes() []Event
	Apply(event Event)
	QueryTypes() []Query
	Query(query Query) QueryResult
}
