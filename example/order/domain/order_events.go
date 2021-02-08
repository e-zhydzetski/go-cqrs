package domain

type OrderPlaced struct {
	ID     string
	Amount int
}

type OrderCancelled struct {
	ID     string
	Reason string
}

type OrderCompleted struct {
	ID       string
	Feedback string
}
