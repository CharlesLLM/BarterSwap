package main

import (
	"fmt"
	"net/http"
)

func (a *API) exchanges(w http.ResponseWriter, r *http.Request, path []string) {
	actor, err := authenticatedUserID(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	if len(path) == 0 {
		switch r.Method {
		case http.MethodGet:
			a.listExchanges(w, r, actor)
		case http.MethodPost:
			a.createExchange(w, r, actor)
		default:
			a.methodNotAllowed(w)
		}
		return
	}
	id, err := pathID(path[0])
	if err != nil {
		a.fail(w, err)
		return
	}
	if len(path) == 1 {
		if r.Method != http.MethodGet {
			a.methodNotAllowed(w)
			return
		}
		a.getExchange(w, r, id, actor)
		return
	}
	if len(path) != 2 {
		a.notFound(w)
		return
	}
	if path[1] == "review" {
		if r.Method != http.MethodPost {
			a.methodNotAllowed(w)
			return
		}
		a.createReview(w, r, id, actor)
		return
	}
	if r.Method != http.MethodPut {
		a.methodNotAllowed(w)
		return
	}
	a.transitionExchange(w, r, id, actor, path[1])
}

func (a *API) listExchanges(w http.ResponseWriter, r *http.Request, actor int64) {
	status := r.URL.Query().Get("status")
	if status != "" && !validStatus(status) {
		a.fail(w, fmt.Errorf("%w: statut inconnu", ErrInvalid))
		return
	}
	exchanges, err := a.store.Exchanges(r.Context(), actor, status)
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, exchanges)
}

func (a *API) createExchange(w http.ResponseWriter, r *http.Request, actor int64) {
	var input struct {
		ServiceID int64 `json:"service_id"`
	}
	if !decode(w, r, &input) {
		return
	}
	exchange, err := a.store.CreateExchange(r.Context(), input.ServiceID, actor)
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, exchange)
}

func (a *API) getExchange(w http.ResponseWriter, r *http.Request, id, actor int64) {
	exchange, err := a.store.Exchange(r.Context(), id)
	if err != nil {
		a.fail(w, err)
		return
	}
	if actor != exchange.OwnerID && actor != exchange.RequesterID {
		a.fail(w, ErrForbidden)
		return
	}
	writeJSON(w, http.StatusOK, exchange)
}

func (a *API) createReview(w http.ResponseWriter, r *http.Request, exchangeID, actor int64) {
	var review Review
	if !decode(w, r, &review) {
		return
	}
	review.ExchangeID, review.AuthorID = exchangeID, actor
	if review.Note < 1 || review.Note > 5 {
		a.fail(w, fmt.Errorf("%w: note entre 1 et 5", ErrInvalid))
		return
	}
	created, err := a.store.CreateReview(r.Context(), review)
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *API) transitionExchange(w http.ResponseWriter, r *http.Request, id, actor int64, action string) {
	if !validExchangeAction(action) {
		a.notFound(w)
		return
	}
	exchange, err := a.store.Transition(r.Context(), id, actor, action)
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, exchange)
}

func validExchangeAction(action string) bool {
	return action == "accept" || action == "reject" || action == "complete" || action == "cancel"
}
