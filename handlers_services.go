package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func servicesHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listServices(w, r, store)
		case http.MethodPost:
			createService(w, r, store)
		default:
			w.Header().Set("Allow", "GET, POST")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
		}
	}
}

func serviceHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		value := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/services/"), "/")
		id, err := strconv.Atoi(value)

		if err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "identifiant invalide"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			getService(w, r, store, id)
		case http.MethodPut:
			updateService(w, r, store, id)
		case http.MethodDelete:
			deleteService(w, r, store, id)
		default:
			w.Header().Set("Allow", "GET, PUT, DELETE")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
		}
	}
}

func listServices(w http.ResponseWriter, r *http.Request, store *Store) {
	filter := ServiceFilter{
		Categorie: r.URL.Query().Get("categorie"),
		Ville:     r.URL.Query().Get("ville"),
		Search:    r.URL.Query().Get("search"),
	}

	services, err := ListServices(r.Context(), store, filter)

	if err != nil {
		writeServiceError(w, err, "liste des services")
		return
	}

	writeJSON(w, http.StatusOK, services)
}

func createService(w http.ResponseWriter, r *http.Request, store *Store) {
	userID, ok := userIDFromHeader(r)

	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "header X-User-ID invalide"})
		return
	}

	var input CreateServiceInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	service, err := CreateService(r.Context(), store, userID, input)

	if err != nil {
		writeServiceError(w, err, "création du service")
		return
	}

	writeJSON(w, http.StatusCreated, service)
}

func getService(w http.ResponseWriter, r *http.Request, store *Store, id int) {
	service, err := GetService(r.Context(), store, id)

	if err != nil {
		writeServiceError(w, err, "lecture du service")
		return
	}

	writeJSON(w, http.StatusOK, service)
}

func updateService(w http.ResponseWriter, r *http.Request, store *Store, id int) {
	userID, ok := userIDFromHeader(r)

	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "header X-User-ID invalide"})
		return
	}

	var input CreateServiceInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	service, err := UpdateService(r.Context(), store, userID, id, input)

	if err != nil {
		writeServiceError(w, err, "modification du service")
		return
	}

	writeJSON(w, http.StatusOK, service)
}

func deleteService(w http.ResponseWriter, r *http.Request, store *Store, id int) {
	userID, ok := userIDFromHeader(r)

	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "header X-User-ID invalide"})
		return
	}

	if err := DeleteService(r.Context(), store, userID, id); err != nil {
		writeServiceError(w, err, "suppression du service")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func userIDFromHeader(r *http.Request) (int, bool) {
	userID, err := strconv.Atoi(r.Header.Get("X-User-ID"))

	if err != nil || userID <= 0 {
		return 0, false
	}

	return userID, true
}

func writeServiceError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, ErrServiceTitleRequired),
		errors.Is(err, ErrServiceCategoryInvalid),
		errors.Is(err, ErrServiceDurationInvalid),
		errors.Is(err, ErrServiceCreditsInvalid),
		errors.Is(err, ErrServiceSkillRequired):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrServiceForbidden):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrServiceNotFound), errors.Is(err, ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	default:
		log.Printf("%s : %v", action, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
	}
}
