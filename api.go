package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// API contient les dépendances de la couche HTTP.
type API struct {
	store  *Store
	logger *log.Logger
}

func NewAPI(store *Store, logger *log.Logger) http.Handler {
	a := &API{store: store, logger: logger}
	return a.middleware(http.HandlerFunc(a.route))
}

type apiError struct {
	Error string `json:"error"`
}

func (a *API) route(w http.ResponseWriter, r *http.Request) {
	segments := pathSegments(r.URL.Path)
	if len(segments) < 2 || segments[0] != "api" {
		a.notFound(w)
		return
	}
	switch segments[1] {
	case "users":
		a.users(w, r, segments[2:])
	case "services":
		a.services(w, r, segments[2:])
	case "exchanges":
		a.exchanges(w, r, segments[2:])
	default:
		a.notFound(w)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if value != nil {
		_ = json.NewEncoder(w).Encode(value)
	}
}

func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{"JSON invalide: " + err.Error()})
		return false
	}
	return true
}

func (a *API) fail(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrInsufficientCredits):
		status = http.StatusBadRequest
	}
	if status == http.StatusInternalServerError {
		a.logger.Printf("erreur interne: %v", err)
		writeJSON(w, status, apiError{"erreur interne"})
		return
	}
	writeJSON(w, status, apiError{err.Error()})
}

func pathID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 1 {
		return 0, ErrNotFound
	}
	return id, nil
}

func authenticatedUserID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("%w: header X-User-ID manquant ou invalide", ErrForbidden)
	}
	return id, nil
}

func pathSegments(path string) []string { return strings.Split(strings.Trim(path, "/"), "/") }
func (a *API) methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, apiError{"méthode non autorisée"})
}
func (a *API) notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, apiError{"route introuvable"})
}
