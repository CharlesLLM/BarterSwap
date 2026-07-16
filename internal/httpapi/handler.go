package httpapi

import (
	"net/http"

	"github.com/CharlesLLM/BarterSwap/internal/application"
)

type Handler struct {
	users     *application.UserService
	catalog   *application.CatalogService
	exchanges *application.ExchangeService
}

func NewHandler(users *application.UserService, catalog *application.CatalogService, exchanges *application.ExchangeService) *Handler {
	return &Handler{users: users, catalog: catalog, exchanges: exchanges}
}

func (handler *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", handler.usersHandler)
	mux.HandleFunc("/api/users/", handler.userHandler)
	mux.HandleFunc("/api/services", handler.servicesHandler)
	mux.HandleFunc("/api/services/", handler.serviceHandler)
	mux.HandleFunc("/api/exchanges", handler.exchangesHandler)
	mux.HandleFunc("/api/exchanges/", handler.exchangeHandler)
	return mux
}
