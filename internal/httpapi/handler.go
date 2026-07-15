package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CharlesLLM/BarterSwap/internal/application"
)

type Handler struct {
	users    *application.UserService
	services *application.ServiceService
}

func NewHandler(users *application.UserService, services *application.ServiceService) *Handler {
	return &Handler{users: users, services: services}
}

func (handler *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", handler.usersHandler)
	mux.HandleFunc("/api/users/", handler.userHandler)
	mux.HandleFunc("/api/services", handler.servicesHandler)
	mux.HandleFunc("/api/services/", handler.serviceHandler)
	return mux
}

func writeJSON(responseWriter http.ResponseWriter, status int, value interface{}) {
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(status)
	if err := json.NewEncoder(responseWriter).Encode(value); err != nil {
		log.Printf("écriture de la réponse JSON : %v", err)
	}
}
