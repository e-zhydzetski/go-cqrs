package domain

import "github.com/e-zhydzetski/go-cqrs/pkg/cqrs"

type PlaceOrder struct {
	Amount int
}

func (p PlaceOrder) AggregateID() string {
	return "" // always create new aggregate
}

type CancelOrder struct {
	cqrs.AggregateID
	Reason string
}

type CompleteOrder struct {
	cqrs.AggregateID
	Feedback string
}
