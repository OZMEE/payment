package model

type Payment struct {
	ID     int64 `json:"id"`
	Amount int   `json:"amount"`
}

type PaymentDto struct {
	Amount int `json:"amount"`
}
