package domain

import (
	"errors"
	"github.com/e-zhydzetski/go-cqrs/pkg/cqrs"
	"github.com/google/uuid"
)

type OrderStatus int

const (
	OrderStatusNew OrderStatus = iota
	OrderStatusCancelled
	OrderStatusCompleted
)

func NewOrder(id string) cqrs.Aggregate {
	return &Order{
		ID: id,
	}
}

type Order struct {
	ID           string
	Status       OrderStatus
	Amount       int
	CancelReason string
	Feedback     string
}

func (o *Order) AggregateID() string {
	return o.ID
}

func (o *Order) Apply(event cqrs.Event) {
	switch e := event.(type) {
	case *OrderPlaced:
		o.ID = e.ID
		o.Amount = e.Amount
		o.Status = OrderStatusNew
	case *OrderCancelled:
		o.Status = OrderStatusCancelled
		o.CancelReason = e.Reason
	case *OrderCompleted:
		o.Status = OrderStatusCompleted
		o.Feedback = e.Feedback
	}
}

func (o *Order) CommandTypes() []cqrs.Command {
	return []cqrs.Command{
		&PlaceOrder{},
		&CancelOrder{},
		&CompleteOrder{},
		&DoNothing{},
	}
}

func (o *Order) Handle(cmd cqrs.Command, actions cqrs.AggregateActions) error {
	switch c := cmd.(type) {
	case *PlaceOrder:
		return o.place(actions, c)
	case *CancelOrder:
		return o.cancel(actions, c)
	case *CompleteOrder:
		return o.complete(actions, c)
	case *DoNothing:
		return nil
	}
	return errors.New("unexpected command")
}

func (o *Order) complete(actions cqrs.AggregateActions, cmd *CompleteOrder) error {
	if o.Status != OrderStatusNew {
		return errors.New("can't complete cancelled or already completed order")
	}
	actions.Emit(&OrderCompleted{
		ID:       o.ID,
		Feedback: cmd.Feedback,
	})
	return nil
}

func (o *Order) cancel(actions cqrs.AggregateActions, cmd *CancelOrder) error {
	if o.Status != OrderStatusNew {
		return errors.New("can't cancel completed or already cancelled order")
	}
	actions.Emit(&OrderCancelled{
		ID:     o.ID,
		Reason: cmd.Reason,
	})
	return nil
}

func (o *Order) place(actions cqrs.AggregateActions, cmd *PlaceOrder) error {
	if o.ID != "" {
		return errors.New("can't place already exists order")
	}
	id := uuid.New().String()
	actions.Emit(&OrderPlaced{
		ID:     id,
		Amount: cmd.Amount,
	})
	return nil
}
