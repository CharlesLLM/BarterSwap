package domain

var (
	ErrReviewNoteInvalid        = Error{Kind: ErrorValidation, Message: "la note doit être comprise entre 1 et 5"}
	ErrReviewExchangeIncomplete = Error{Kind: ErrorValidation, Message: "l'échange doit être terminé avant de laisser un avis"}
	ErrReviewForbidden          = Error{Kind: ErrorForbidden, Message: "vous ne pouvez pas évaluer cet échange"}
	ErrReviewAlreadyExists      = Error{Kind: ErrorConflict, Message: "vous avez déjà évalué cet échange"}
)

type Review struct {
	ID          int    `json:"id"`
	ExchangeID  int    `json:"exchange_id"`
	AuthorID    int    `json:"author_id"`
	TargetID    int    `json:"target_id"`
	Note        int    `json:"note"`
	Commentaire string `json:"commentaire,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type CreateReviewInput struct {
	Note        int    `json:"note"`
	Commentaire string `json:"commentaire"`
}
