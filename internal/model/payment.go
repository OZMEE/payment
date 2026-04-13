package model

type PaymentDb struct {
	ID     string `json:"id"`
	Amount int    `json:"amount"`
}

type PaymentDto struct {
	Amount int `json:"amount"`
}
