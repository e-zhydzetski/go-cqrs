package main

import (
	"context"
	"github.com/e-zhydzetski/go-commons/pkg/graceful"
	"github.com/e-zhydzetski/go-cqrs/example/order/domain"
	"github.com/e-zhydzetski/go-cqrs/pkg/cqrs"
	"github.com/e-zhydzetski/go-cqrs/pkg/es/inmemes"
	"log"
	"time"
)

func main() {
	ctx := context.Background()
	ctx = graceful.WrapContext(ctx)
	app := cqrs.NewSimpleApp(ctx, inmemes.New())
	app.RegisterAggregate(domain.NewOrder)
	app.RegisterAggregate(domain.NewPayment)
	app.RegisterView(domain.NewOrdersView())

	log.Println("Ready...")
	events, err := app.Command(&domain.PlaceOrder{Amount: 100})
	if err != nil {
		log.Fatalln(err)
	}
	orderPlaced := domain.OrderPlaced{}
	if !events.Get(&orderPlaced) {
		log.Fatalln("order not placed")
	}
	log.Println(orderPlaced)

	time.Sleep(time.Millisecond * 10)

	res, err := app.Query(&domain.GetCompletedOrdersQuery{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(res)

	_, _ = app.Command(&domain.CompleteOrder{
		AggrID:   cqrs.AggrID(orderPlaced.ID),
		Feedback: "All right!!!",
	})

	time.Sleep(time.Millisecond * 10)

	res, err = app.Query(&domain.GetCompletedOrdersQuery{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(res)
}
