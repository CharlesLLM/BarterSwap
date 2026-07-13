package main

import (
	"fmt"
	"net/http"
)

func (a *API) services(w http.ResponseWriter, r *http.Request, path []string) {
	if len(path) == 0 {
		switch r.Method {
		case http.MethodGet:
			a.listServices(w, r)
		case http.MethodPost:
			a.createService(w, r)
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
	if len(path) == 2 && path[1] == "reviews" {
		if r.Method != http.MethodGet {
			a.methodNotAllowed(w)
			return
		}
		a.writeReviews(w, r, "service", id)
		return
	}
	if len(path) != 1 {
		a.notFound(w)
		return
	}
	service, err := a.store.Service(r.Context(), id)
	if err != nil {
		a.fail(w, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, service)
	case http.MethodPut:
		a.updateService(w, r, service)
	case http.MethodDelete:
		a.deleteService(w, r, service)
	default:
		a.methodNotAllowed(w)
	}
}

func (a *API) listServices(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := ServiceFilter{
		Categorie: query.Get("categorie"),
		Ville:     query.Get("ville"),
		Search:    query.Get("search"),
	}
	services, err := a.store.Services(r.Context(), filter)
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, services)
}

func (a *API) createService(w http.ResponseWriter, r *http.Request) {
	actor, err := authenticatedUserID(r)
	if err != nil {
		a.fail(w, err)
		return
	}
	var service Service
	if !decode(w, r, &service) {
		return
	}
	service.ProviderID = actor
	if err = validateService(service); err != nil {
		a.fail(w, err)
		return
	}
	if err = a.ensureProviderSkill(r, service); err != nil {
		a.fail(w, err)
		return
	}
	created, err := a.store.CreateService(r.Context(), service)
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *API) updateService(w http.ResponseWriter, r *http.Request, current Service) {
	actor, err := authenticatedUserID(r)
	if err != nil || actor != current.ProviderID {
		a.fail(w, ErrForbidden)
		return
	}
	var service Service
	if !decode(w, r, &service) {
		return
	}
	service.ID, service.ProviderID = current.ID, actor
	if err = validateService(service); err != nil {
		a.fail(w, err)
		return
	}
	if err = a.ensureProviderSkill(r, service); err != nil {
		a.fail(w, err)
		return
	}
	updated, err := a.store.UpdateService(r.Context(), service)
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *API) deleteService(w http.ResponseWriter, r *http.Request, service Service) {
	actor, err := authenticatedUserID(r)
	if err != nil || actor != service.ProviderID {
		a.fail(w, ErrForbidden)
		return
	}
	if err = a.store.DeleteService(r.Context(), service.ID); err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (a *API) ensureProviderSkill(r *http.Request, service Service) error {
	ok, err := a.store.HasSkill(r.Context(), service.ProviderID, service.Categorie)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: la catégorie doit correspondre à une compétence", ErrInvalid)
	}
	return nil
}

func validateService(service Service) error {
	_, validCategory := categories[service.Categorie]
	if clean(service.Titre) == "" || !validCategory || service.DureeMinutes < 1 || service.Credits < 1 {
		return fmt.Errorf("%w: titre, catégorie, durée et crédits requis", ErrInvalid)
	}
	return nil
}
