package domain

type PreparePayment struct {
	OrderID string
	Amount  int
}

type ConfirmPayment struct {
	PaymentID string
}

type CancelPayment struct {
	PaymentID string
}

type RevertPayment struct {
	PaymentID string
}
