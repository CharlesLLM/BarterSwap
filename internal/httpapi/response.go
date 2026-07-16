package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

var statusByErrorKind = map[domain.ErrorKind]int{
	domain.ErrorValidation: http.StatusBadRequest,
	domain.ErrorConflict:   http.StatusConflict,
	domain.ErrorForbidden:  http.StatusForbidden,
	domain.ErrorNotFound:   http.StatusNotFound,
}

func writeJSON(responseWriter http.ResponseWriter, status int, value any) {
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(status)
	if err := json.NewEncoder(responseWriter).Encode(value); err != nil {
		fmt.Printf("écriture de la réponse JSON : %v\n", err)
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
		fmt.Printf("%s : %v\n", action, err)
		writeError(responseWriter, status, "erreur interne")
		return
	}
	writeError(responseWriter, status, err.Error())
}

func statusForError(err error) int {
	var domainError domain.Error
	if !errors.As(err, &domainError) {
		return http.StatusInternalServerError
	}

	status, found := statusByErrorKind[domainError.Kind]
	if !found {
		return http.StatusInternalServerError
	}
	return status
}
