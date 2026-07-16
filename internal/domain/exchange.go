package domain

import "errors"

const (
	ExchangeStatusPending   = "pending"
	ExchangeStatusAccepted  = "accepted"
	ExchangeStatusRejected  = "rejected"
	ExchangeStatusCancelled = "cancelled"
	ExchangeStatusCompleted = "completed"
)

var (
	ErrExchangeNotFound         = errors.New("échange introuvable")
	ErrExchangeServiceRequired  = errors.New("l'identifiant du service est obligatoire")
	ErrExchangeStatusInvalid    = errors.New("le statut de l'échange est invalide")
	ErrExchangeTransition       = errors.New("cette transition d'échange est impossible")
	ErrExchangeSelfService      = errors.New("un utilisateur ne peut pas demander son propre service")
	ErrExchangeConflict         = errors.New("ce service a déjà un échange en cours")
	ErrExchangeForbidden        = errors.New("vous ne pouvez pas effectuer cette action sur cet échange")
	ErrExchangeInsufficientFund = errors.New("crédits insuffisants")
)

type Exchange struct {
	ID          int    `json:"id"`
	ServiceID   int    `json:"service_id"`
	RequesterID int    `json:"requester_id"`
	OwnerID     int    `json:"owner_id"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Credits     int    `json:"-"`
}

type CreateExchangeInput struct {
	ServiceID int `json:"service_id"`
}

type ExchangeFilter struct {
	Status string
}

func IsValidExchangeStatus(status string) bool {
	switch status {
	case ExchangeStatusPending,
		ExchangeStatusAccepted,
		ExchangeStatusRejected,
		ExchangeStatusCancelled,
		ExchangeStatusCompleted:
		return true
	default:
		return false
	}
}
