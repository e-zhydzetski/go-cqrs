package domain

import "github.com/e-zhydzetski/go-cqrs/pkg/cqrs"

type PreparePayment struct {
	OrderID string
	Amount  int
}

func (p PreparePayment) AggregateID() string {
	return ""
}

type ConfirmPayment struct {
	cqrs.AggregateID
}

type CancelPayment struct {
	cqrs.AggregateID
}

type RevertPayment struct {
	cqrs.AggregateID
}
