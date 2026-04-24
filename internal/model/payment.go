package model

type Payment struct {
	ID        int64 `json:"id"`
	PaymentId int64 `json:"payment_id"`
	Amount    int64 `json:"amount"`
}

type PaymentDto struct {
	PaymentId int64 `json:"payment_id"`
	Amount    int64 `json:"amount"`
}
