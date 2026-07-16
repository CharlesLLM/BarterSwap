package httpapi

import (
	"net/http"

	"github.com/CharlesLLM/BarterSwap/internal/application"
)

type Handler struct {
	users     application.UserService
	catalog   application.CatalogService
	exchanges application.ExchangeService
	reviews   application.ReviewService
}

func NewHandler(
	users application.UserService,
	catalog application.CatalogService,
	exchanges application.ExchangeService,
	reviews application.ReviewService,
) Handler {
	return Handler{users: users, catalog: catalog, exchanges: exchanges, reviews: reviews}
}

func (handler Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", handler.usersHandler)
	mux.HandleFunc("/api/users/", handler.userHandler)
	mux.HandleFunc("/api/services", handler.servicesHandler)
	mux.HandleFunc("/api/services/", handler.serviceHandler)
	mux.HandleFunc("/api/exchanges", handler.exchangesHandler)
	mux.HandleFunc("/api/exchanges/", handler.exchangeHandler)
	mux.HandleFunc("/openapi.yaml", openAPIHandler)
	mux.HandleFunc("/swagger", swaggerRedirectHandler)
	mux.HandleFunc("/swagger/", swaggerHandler)
	return mux
}
