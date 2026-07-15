package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

func (handler *Handler) servicesHandler(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		handler.listServices(responseWriter, request)
	case http.MethodPost:
		handler.createService(responseWriter, request)
	default:
		responseWriter.Header().Set("Allow", "GET, POST")
		writeJSON(responseWriter, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
	}
}

func (handler *Handler) serviceHandler(responseWriter http.ResponseWriter, request *http.Request) {
	value := strings.Trim(strings.TrimPrefix(request.URL.Path, "/api/services/"), "/")
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "identifiant invalide"})
		return
	}

	switch request.Method {
	case http.MethodGet:
		handler.getService(responseWriter, request, id)
	case http.MethodPut:
		handler.updateService(responseWriter, request, id)
	case http.MethodDelete:
		handler.deleteService(responseWriter, request, id)
	default:
		responseWriter.Header().Set("Allow", "GET, PUT, DELETE")
		writeJSON(responseWriter, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
	}
}

func (handler *Handler) listServices(responseWriter http.ResponseWriter, request *http.Request) {
	filter := domain.ServiceFilter{
		Categorie: request.URL.Query().Get("categorie"),
		Ville:     request.URL.Query().Get("ville"),
		Search:    request.URL.Query().Get("search"),
	}
	services, err := handler.services.List(request.Context(), filter)
	if err != nil {
		writeServiceError(responseWriter, err, "liste des services")
		return
	}
	writeJSON(responseWriter, http.StatusOK, services)
}

func (handler *Handler) createService(responseWriter http.ResponseWriter, request *http.Request) {
	userID, valid := userIDFromHeader(request)
	if !valid {
		writeJSON(responseWriter, http.StatusUnauthorized, map[string]string{"error": "header X-User-ID invalide"})
		return
	}

	var input domain.CreateServiceInput
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	service, err := handler.services.Create(request.Context(), userID, input)
	if err != nil {
		writeServiceError(responseWriter, err, "création du service")
		return
	}
	writeJSON(responseWriter, http.StatusCreated, service)
}

func (handler *Handler) getService(responseWriter http.ResponseWriter, request *http.Request, id int) {
	service, err := handler.services.Get(request.Context(), id)
	if err != nil {
		writeServiceError(responseWriter, err, "lecture du service")
		return
	}
	writeJSON(responseWriter, http.StatusOK, service)
}

func (handler *Handler) updateService(responseWriter http.ResponseWriter, request *http.Request, id int) {
	userID, valid := userIDFromHeader(request)
	if !valid {
		writeJSON(responseWriter, http.StatusUnauthorized, map[string]string{"error": "header X-User-ID invalide"})
		return
	}

	var input domain.CreateServiceInput
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	service, err := handler.services.Update(request.Context(), userID, id, input)
	if err != nil {
		writeServiceError(responseWriter, err, "modification du service")
		return
	}
	writeJSON(responseWriter, http.StatusOK, service)
}

func (handler *Handler) deleteService(responseWriter http.ResponseWriter, request *http.Request, id int) {
	userID, valid := userIDFromHeader(request)
	if !valid {
		writeJSON(responseWriter, http.StatusUnauthorized, map[string]string{"error": "header X-User-ID invalide"})
		return
	}
	if err := handler.services.Delete(request.Context(), userID, id); err != nil {
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
	case errors.Is(err, domain.ErrServiceTitleRequired),
		errors.Is(err, domain.ErrServiceCategoryInvalid),
		errors.Is(err, domain.ErrServiceDurationInvalid),
		errors.Is(err, domain.ErrServiceCreditsInvalid),
		errors.Is(err, domain.ErrServiceSkillRequired):
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrServiceForbidden):
		writeJSON(responseWriter, http.StatusForbidden, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrServiceNotFound), errors.Is(err, domain.ErrUserNotFound):
		writeJSON(responseWriter, http.StatusNotFound, map[string]string{"error": err.Error()})
	default:
		log.Printf("%s : %v", action, err)
		writeJSON(responseWriter, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
	}
}
