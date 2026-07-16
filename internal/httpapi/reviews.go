package httpapi

import (
	"net/http"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

func (handler Handler) createReview(responseWriter http.ResponseWriter, request *http.Request, exchangeID int) {
	authorID, valid := requireUserID(responseWriter, request)
	if !valid {
		return
	}

	var input domain.CreateReviewInput
	if !decodeJSON(responseWriter, request, &input) {
		return
	}
	review, err := handler.reviews.Create(request.Context(), exchangeID, authorID, input)
	if err != nil {
		writeApplicationError(responseWriter, err, "création de l'avis")
		return
	}
	writeJSON(responseWriter, http.StatusCreated, review)
}

func (handler Handler) listUserReviews(responseWriter http.ResponseWriter, request *http.Request, userID int) {
	reviews, err := handler.reviews.ListForUser(request.Context(), userID)
	if err != nil {
		writeApplicationError(responseWriter, err, "liste des avis de l'utilisateur")
		return
	}
	writeJSON(responseWriter, http.StatusOK, reviews)
}

func (handler Handler) listServiceReviews(responseWriter http.ResponseWriter, request *http.Request, serviceID int) {
	reviews, err := handler.reviews.ListForService(request.Context(), serviceID)
	if err != nil {
		writeApplicationError(responseWriter, err, "liste des avis du service")
		return
	}
	writeJSON(responseWriter, http.StatusOK, reviews)
}
