package domain

const (
	ExchangeStatusPending   = "pending"
	ExchangeStatusAccepted  = "accepted"
	ExchangeStatusRejected  = "rejected"
	ExchangeStatusCancelled = "cancelled"
	ExchangeStatusCompleted = "completed"
)

var (
	ErrExchangeNotFound         = Error{Kind: ErrorNotFound, Message: "échange introuvable"}
	ErrExchangeServiceRequired  = Error{Kind: ErrorValidation, Message: "l'identifiant du service est obligatoire"}
	ErrExchangeStatusInvalid    = Error{Kind: ErrorValidation, Message: "le statut de l'échange est invalide"}
	ErrExchangeTransition       = Error{Kind: ErrorValidation, Message: "cette transition d'échange est impossible"}
	ErrExchangeSelfService      = Error{Kind: ErrorValidation, Message: "un utilisateur ne peut pas demander son propre service"}
	ErrExchangeConflict         = Error{Kind: ErrorConflict, Message: "ce service a déjà un échange en cours"}
	ErrExchangeForbidden        = Error{Kind: ErrorForbidden, Message: "vous ne pouvez pas effectuer cette action sur cet échange"}
	ErrExchangeInsufficientFund = Error{Kind: ErrorValidation, Message: "crédits insuffisants"}
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
