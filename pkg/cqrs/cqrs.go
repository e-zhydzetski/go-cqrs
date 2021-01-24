package cqrs

type Command interface{}

type Query interface{}
type QueryResult interface{}

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
