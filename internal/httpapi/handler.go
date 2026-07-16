package httpapi

import (
	"net/http"

	"github.com/CharlesLLM/BarterSwap/internal/application"
)

type Handler struct {
	users   *application.UserService
	catalog *application.CatalogService
}

func NewHandler(users *application.UserService, catalog *application.CatalogService) *Handler {
	return &Handler{users: users, catalog: catalog}
}

func (handler *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", handler.usersHandler)
	mux.HandleFunc("/api/users/", handler.userHandler)
	mux.HandleFunc("/api/services", handler.servicesHandler)
	mux.HandleFunc("/api/services/", handler.serviceHandler)
	return mux
}
