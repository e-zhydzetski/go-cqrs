package cqrs

import "reflect"

func AggregateToESStreamName(aggregateType AggregateType, aggregateID string) string {
	return aggregateType.String() + "!" + aggregateID // TODO make stable aggregate type name
}

func EventToESEventType(event Event) string {
	return reflect.TypeOf(event).String() // TODO make stable event type name
}
