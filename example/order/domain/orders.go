package domain

import (
	"errors"
	"github.com/e-zhydzetski/go-cqrs/pkg/cqrs"
	"github.com/e-zhydzetski/go-cqrs/pkg/es"
)

func NewOrdersView() *OrdersView {
	return &OrdersView{
		TotalOrders:     0,
		CompletedOrders: map[string]string{},
	}
}

type OrdersView struct {
	Seq             es.StorePosition
	TotalOrders     uint32
	CompletedOrders map[string]string
}

func (o *OrdersView) Apply(event cqrs.Event, globalSequence es.StorePosition) {
	defer func() { o.Seq = globalSequence }() // update after the sequence event applied
	switch e := event.(type) {
	case *OrderPlaced:
		o.TotalOrders++
	case *OrderCancelled:
	case *OrderCompleted:
		o.CompletedOrders[e.ID] = e.Feedback
	}
}

func (o *OrdersView) GetLastAppliedSeq() es.StorePosition {
	return o.Seq
}

func (o *OrdersView) QueryTypes() []cqrs.Query {
	return []cqrs.Query{&GetCompletedOrdersQuery{}}
}

func (o *OrdersView) Query(query cqrs.Query) (interface{}, error) {
	switch q := query.(type) {
	case *GetCompletedOrdersQuery:
		_ = q
		return &GetCompletedOrdersQueryResult{
			Total:     o.TotalOrders,
			Completed: o.CompletedOrders,
		}, nil
	}
	return nil, errors.New("unexpected query")
}

type GetCompletedOrdersQuery struct {
}

type GetCompletedOrdersQueryResult struct {
	Total     uint32
	Completed map[string]string
}
