package domain

import "github.com/e-zhydzetski/go-cqrs/pkg/cqrs"

func NewPayment(id string) cqrs.Aggregate {
	return &Payment{
		ID: id,
	}
}

type Payment struct {
	ID        string
	OrderID   string
	Amount    int
	Completed bool
	Reversed  bool
}

func (p *Payment) AggregateID() string {
	return p.ID
}

func (p *Payment) Apply(event cqrs.Event) {
	panic("implement me")
}

func (p *Payment) CommandTypes() []cqrs.Command {
	panic("implement me")
}

func (p *Payment) Handle(command cqrs.Command, actions cqrs.AggregateActions) error {
	panic("implement me")
}
