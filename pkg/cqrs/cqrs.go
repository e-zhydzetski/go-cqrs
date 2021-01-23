package cqrs

type Event interface{}

type Command interface{}

type Query interface{}
type QueryResult interface{}

type Aggregate interface {
	AggregateID() string
	Apply(event Event)
	CommandTypes() []Command
	Handle(command Command, actions AggregateActions)
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
