package httpapi

import (
	"net/http"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

func (handler Handler) servicesHandler(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		handler.listServices(responseWriter, request)
	case http.MethodPost:
		handler.createService(responseWriter, request)
	default:
		methodNotAllowed(responseWriter, http.MethodGet, http.MethodPost)
	}
}

func (handler Handler) serviceHandler(responseWriter http.ResponseWriter, request *http.Request) {
	parts := pathSegments(request.URL.Path, "/api/services/")
	if len(parts) != 1 {
		writeError(responseWriter, http.StatusNotFound, "route introuvable")
		return
	}

	id, valid := positiveInteger(parts[0])
	if !valid {
		writeError(responseWriter, http.StatusBadRequest, "identifiant invalide")
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
		methodNotAllowed(responseWriter, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func (handler Handler) listServices(responseWriter http.ResponseWriter, request *http.Request) {
	filter := domain.ServiceFilter{
		Categorie: request.URL.Query().Get("categorie"),
		Ville:     request.URL.Query().Get("ville"),
		Search:    request.URL.Query().Get("search"),
	}
	services, err := handler.catalog.List(request.Context(), filter)
	if err != nil {
		writeApplicationError(responseWriter, err, "liste des services")
		return
	}
	writeJSON(responseWriter, http.StatusOK, services)
}

func (handler Handler) createService(responseWriter http.ResponseWriter, request *http.Request) {
	userID, valid := requireUserID(responseWriter, request)
	if !valid {
		return
	}

	var input domain.CreateServiceInput
	if !decodeJSON(responseWriter, request, &input) {
		return
	}

	service, err := handler.catalog.Create(request.Context(), userID, input)
	if err != nil {
		writeApplicationError(responseWriter, err, "création du service")
		return
	}
	writeJSON(responseWriter, http.StatusCreated, service)
}

func (handler Handler) getService(responseWriter http.ResponseWriter, request *http.Request, id int) {
	service, err := handler.catalog.Get(request.Context(), id)
	if err != nil {
		writeApplicationError(responseWriter, err, "lecture du service")
		return
	}
	writeJSON(responseWriter, http.StatusOK, service)
}

func (handler Handler) updateService(responseWriter http.ResponseWriter, request *http.Request, id int) {
	userID, valid := requireUserID(responseWriter, request)
	if !valid {
		return
	}

	var input domain.CreateServiceInput
	if !decodeJSON(responseWriter, request, &input) {
		return
	}

	service, err := handler.catalog.Update(request.Context(), userID, id, input)
	if err != nil {
		writeApplicationError(responseWriter, err, "modification du service")
		return
	}
	writeJSON(responseWriter, http.StatusOK, service)
}

func (handler Handler) deleteService(responseWriter http.ResponseWriter, request *http.Request, id int) {
	userID, valid := requireUserID(responseWriter, request)
	if !valid {
		return
	}
	if err := handler.catalog.Delete(request.Context(), userID, id); err != nil {
		writeApplicationError(responseWriter, err, "suppression du service")
		return
	}
	responseWriter.WriteHeader(http.StatusNoContent)
}
