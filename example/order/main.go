package main

import (
	"context"
	"github.com/e-zhydzetski/go-commons/pkg/graceful"
	"github.com/e-zhydzetski/go-cqrs/example/order/domain"
	"github.com/e-zhydzetski/go-cqrs/pkg/cqrs"
)

func main() {
	ctx := context.Background()
	ctx = graceful.WrapContext(ctx)
	app := cqrs.NewSimpleApp(ctx, nil)
	app.RegisterAggregate(domain.NewOrder)
	app.RegisterAggregate(domain.NewPayment)
}
