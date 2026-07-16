package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

func writeJSON(responseWriter http.ResponseWriter, status int, value any) {
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(status)
	if err := json.NewEncoder(responseWriter).Encode(value); err != nil {
		log.Printf("écriture de la réponse JSON : %v", err)
	}
}

func writeError(responseWriter http.ResponseWriter, status int, message string) {
	writeJSON(responseWriter, status, map[string]string{"error": message})
}

func methodNotAllowed(responseWriter http.ResponseWriter, methods ...string) {
	responseWriter.Header().Set("Allow", strings.Join(methods, ", "))
	writeError(responseWriter, http.StatusMethodNotAllowed, "méthode non autorisée")
}

func writeApplicationError(responseWriter http.ResponseWriter, err error, action string) {
	status := statusForError(err)
	if status == http.StatusInternalServerError {
		log.Printf("%s : %v", action, err)
		writeError(responseWriter, status, "erreur interne")
		return
	}
	writeError(responseWriter, status, err.Error())
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, domain.ErrPseudoRequired),
		errors.Is(err, domain.ErrSkillNameRequired),
		errors.Is(err, domain.ErrSkillLevelInvalid),
		errors.Is(err, domain.ErrSkillDuplicate),
		errors.Is(err, domain.ErrServiceTitleRequired),
		errors.Is(err, domain.ErrServiceCategoryInvalid),
		errors.Is(err, domain.ErrServiceDurationInvalid),
		errors.Is(err, domain.ErrServiceCreditsInvalid),
		errors.Is(err, domain.ErrServiceSkillRequired),
		errors.Is(err, domain.ErrExchangeServiceRequired),
		errors.Is(err, domain.ErrExchangeStatusInvalid),
		errors.Is(err, domain.ErrExchangeTransition),
		errors.Is(err, domain.ErrExchangeSelfService),
		errors.Is(err, domain.ErrExchangeInsufficientFund):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrPseudoAlreadyExists), errors.Is(err, domain.ErrExchangeConflict):
		return http.StatusConflict
	case errors.Is(err, domain.ErrServiceForbidden), errors.Is(err, domain.ErrExchangeForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrServiceNotFound),
		errors.Is(err, domain.ErrExchangeNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
