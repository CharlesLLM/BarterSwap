package main

const (
	ExchangeStatusPending   = "pending"
	ExchangeStatusAccepted  = "accepted"
	ExchangeStatusRejected  = "rejected"
	ExchangeStatusCancelled = "cancelled"
	ExchangeStatusCompleted = "completed"
)

type Exchange struct {
	ID          int    `json:"id"`
	ServiceID   int    `json:"service_id"`
	RequesterID int    `json:"requester_id"`
	OwnerID     int    `json:"owner_id"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
