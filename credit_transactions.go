package main

const (
	CreditTypeWelcome = "welcome"
	CreditTypeEarn    = "earn"
	CreditTypeSpend   = "spend"
	CreditTypeRefund  = "refund"
)

type CreditTransaction struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	ExchangeID int    `json:"exchange_id"`
	Montant    int    `json:"montant"`
	Type       string `json:"type"`
	CreatedAt  string `json:"created_at"`
}
