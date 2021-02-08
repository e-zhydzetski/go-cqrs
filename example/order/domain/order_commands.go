package domain

type PlaceOrder struct {
	Amount int
}

type CancelOrder struct {
	OrderID string
	Reason  string
}

type CompleteOrder struct {
	OrderID  string
	Feedback string
}
