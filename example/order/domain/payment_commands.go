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
	cqrs.AggrID
}

type CancelPayment struct {
	cqrs.AggrID
}

type RevertPayment struct {
	cqrs.AggrID
}
