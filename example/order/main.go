package main

import (
	"context"
	"github.com/e-zhydzetski/go-commons/pkg/graceful"
	"github.com/e-zhydzetski/go-cqrs/example/order/domain"
	"github.com/e-zhydzetski/go-cqrs/pkg/cqrs"
	"github.com/e-zhydzetski/go-cqrs/pkg/es/inmemes"
	"log"
)

func main() {
	ctx := context.Background()
	ctx = graceful.WrapContext(ctx)
	app := cqrs.NewSimpleApp(ctx, inmemes.New())
	app.RegisterAggregate(domain.NewOrder)
	app.RegisterAggregate(domain.NewPayment)
	app.RegisterView(domain.NewOrdersView())

	log.Println("Ready...")
	result, err := app.Command(&domain.PlaceOrder{Amount: 100})
	if err != nil {
		log.Fatalln(err)
	}
	orderPlaced := domain.OrderPlaced{}
	if !result.Events.Get(&orderPlaced) {
		log.Fatalln("order not placed")
	}
	log.Println(orderPlaced)

	res, err := app.Query(ctx, &domain.GetCompletedOrdersQuery{}, result.LastEventSequenceNumber)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(res)

	result, err = app.Command(&domain.CompleteOrder{
		AggrID:   cqrs.AggrID(orderPlaced.ID),
		Feedback: "All right!!!",
	})
	if err != nil {
		log.Fatalln(err)
	}

	res, err = app.Query(ctx, &domain.GetCompletedOrdersQuery{}, result.LastEventSequenceNumber)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(res)
}
