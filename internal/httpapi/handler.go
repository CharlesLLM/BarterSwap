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

	mux.HandleFunc("POST /api/users", handler.createUser)
	mux.HandleFunc("GET /api/users", handler.listUsers)
	mux.HandleFunc("GET /api/users/{id}", withID(handler.getUser))
	mux.HandleFunc("PUT /api/users/{id}", withID(handler.updateUser))
	mux.HandleFunc("DELETE /api/users/{id}", withID(handler.deleteUser))
	mux.HandleFunc("GET /api/users/{id}/skills", withID(handler.getUserSkills))
	mux.HandleFunc("PUT /api/users/{id}/skills", withID(handler.replaceUserSkills))
	mux.HandleFunc("GET /api/users/{id}/reviews", withID(handler.listUserReviews))
	mux.HandleFunc("GET /api/users/{id}/stats", withID(handler.getUserStats))

	mux.HandleFunc("GET /api/services", handler.listServices)
	mux.HandleFunc("POST /api/services", handler.createService)
	mux.HandleFunc("GET /api/services/{id}", withID(handler.getService))
	mux.HandleFunc("PUT /api/services/{id}", withID(handler.updateService))
	mux.HandleFunc("DELETE /api/services/{id}", withID(handler.deleteService))
	mux.HandleFunc("GET /api/services/{id}/reviews", withID(handler.listServiceReviews))

	mux.HandleFunc("POST /api/exchanges", handler.createExchange)
	mux.HandleFunc("GET /api/exchanges", handler.listExchanges)
	mux.HandleFunc("GET /api/exchanges/{id}", withID(handler.getExchange))
	mux.HandleFunc("PUT /api/exchanges/{id}/accept", withID(handler.acceptExchange))
	mux.HandleFunc("PUT /api/exchanges/{id}/reject", withID(handler.rejectExchange))
	mux.HandleFunc("PUT /api/exchanges/{id}/complete", withID(handler.completeExchange))
	mux.HandleFunc("PUT /api/exchanges/{id}/cancel", withID(handler.cancelExchange))
	mux.HandleFunc("POST /api/exchanges/{id}/review", withID(handler.createReview))

	mux.HandleFunc("GET /openapi.yaml", openAPIHandler)
	mux.HandleFunc("GET /swagger", swaggerRedirectHandler)
	mux.HandleFunc("GET /swagger/", swaggerHandler)

	return chain(mux, withLogging, withRecovery, withCORS)
}

type handlerWithID func(http.ResponseWriter, *http.Request, int)

func withID(handler handlerWithID) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		id, valid := positiveInteger(request.PathValue("id"))
		if !valid {
			writeError(responseWriter, http.StatusBadRequest, "identifiant invalide")
			return
		}
		handler(responseWriter, request, id)
	}
}
