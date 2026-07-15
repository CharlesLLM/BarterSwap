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
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			listServices(responseWriter, request, store)
		case http.MethodPost:
			createService(responseWriter, request, store)
		default:
			responseWriter.Header().Set("Allow", "GET, POST")
			writeJSON(responseWriter, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
		}
	}
}

func serviceHandler(store *Store) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		value := strings.Trim(strings.TrimPrefix(request.URL.Path, "/api/services/"), "/")
		id, err := strconv.Atoi(value)

		if err != nil || id <= 0 {
			writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "identifiant invalide"})
			return
		}

		switch request.Method {
		case http.MethodGet:
			getService(responseWriter, request, store, id)
		case http.MethodPut:
			updateService(responseWriter, request, store, id)
		case http.MethodDelete:
			deleteService(responseWriter, request, store, id)
		default:
			responseWriter.Header().Set("Allow", "GET, PUT, DELETE")
			writeJSON(responseWriter, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
		}
	}
}

func listServices(responseWriter http.ResponseWriter, request *http.Request, store *Store) {
	filter := ServiceFilter{
		Categorie: request.URL.Query().Get("categorie"),
		Ville:     request.URL.Query().Get("ville"),
		Search:    request.URL.Query().Get("search"),
	}

	services, err := ListServices(request.Context(), store, filter)

	if err != nil {
		writeServiceError(responseWriter, err, "liste des services")
		return
	}

	writeJSON(responseWriter, http.StatusOK, services)
}

func createService(responseWriter http.ResponseWriter, request *http.Request, store *Store) {
	userID, userIDIsValid := userIDFromHeader(request)

	if !userIDIsValid {
		writeJSON(responseWriter, http.StatusUnauthorized, map[string]string{"error": "header X-User-ID invalide"})
		return
	}

	var input CreateServiceInput
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	service, err := CreateService(request.Context(), store, userID, input)

	if err != nil {
		writeServiceError(responseWriter, err, "création du service")
		return
	}

	writeJSON(responseWriter, http.StatusCreated, service)
}

func getService(responseWriter http.ResponseWriter, request *http.Request, store *Store, id int) {
	service, err := GetService(request.Context(), store, id)

	if err != nil {
		writeServiceError(responseWriter, err, "lecture du service")
		return
	}

	writeJSON(responseWriter, http.StatusOK, service)
}

func updateService(responseWriter http.ResponseWriter, request *http.Request, store *Store, id int) {
	userID, userIDIsValid := userIDFromHeader(request)

	if !userIDIsValid {
		writeJSON(responseWriter, http.StatusUnauthorized, map[string]string{"error": "header X-User-ID invalide"})
		return
	}

	var input CreateServiceInput
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	service, err := UpdateService(request.Context(), store, userID, id, input)

	if err != nil {
		writeServiceError(responseWriter, err, "modification du service")
		return
	}

	writeJSON(responseWriter, http.StatusOK, service)
}

func deleteService(responseWriter http.ResponseWriter, request *http.Request, store *Store, id int) {
	userID, userIDIsValid := userIDFromHeader(request)

	if !userIDIsValid {
		writeJSON(responseWriter, http.StatusUnauthorized, map[string]string{"error": "header X-User-ID invalide"})
		return
	}

	if err := DeleteService(request.Context(), store, userID, id); err != nil {
		writeServiceError(responseWriter, err, "suppression du service")
		return
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}

func userIDFromHeader(request *http.Request) (int, bool) {
	userID, err := strconv.Atoi(request.Header.Get("X-User-ID"))

	if err != nil || userID <= 0 {
		return 0, false
	}

	return userID, true
}

func writeServiceError(responseWriter http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, ErrServiceTitleRequired),
		errors.Is(err, ErrServiceCategoryInvalid),
		errors.Is(err, ErrServiceDurationInvalid),
		errors.Is(err, ErrServiceCreditsInvalid),
		errors.Is(err, ErrServiceSkillRequired):
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrServiceForbidden):
		writeJSON(responseWriter, http.StatusForbidden, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrServiceNotFound), errors.Is(err, ErrUserNotFound):
		writeJSON(responseWriter, http.StatusNotFound, map[string]string{"error": err.Error()})
	default:
		log.Printf("%s : %v", action, err)
		writeJSON(responseWriter, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
	}
}
