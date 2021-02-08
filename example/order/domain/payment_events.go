package domain

type PaymentPrepared struct {
	ID string
}

type PaymentCompleted struct {
	ID string
}

type PaymentCancelled struct {
	ID string
}

type PaymentReverted struct {
	ID string
}
