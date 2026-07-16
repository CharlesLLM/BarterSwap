package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CharlesLLM/BarterSwap/internal/application"
	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type reviewHTTPRepositoryStub struct{}

func (reviewHTTPRepositoryStub) FindExchange(context.Context, int) (domain.Exchange, error) {
	return domain.Exchange{ID: 8, ServiceID: 5, RequesterID: 1, OwnerID: 2, Status: domain.ExchangeStatusCompleted}, nil
}

func (reviewHTTPRepositoryStub) FindUser(context.Context, int) (domain.User, error) {
	return domain.User{ID: 2}, nil
}

func (reviewHTTPRepositoryStub) FindService(context.Context, int) (domain.Service, error) {
	return domain.Service{ID: 5}, nil
}

func (reviewHTTPRepositoryStub) CreateReview(
	_ context.Context,
	exchangeID int,
	authorID int,
	targetID int,
	input domain.CreateReviewInput,
) (domain.Review, error) {
	return domain.Review{
		ID: 1, ExchangeID: exchangeID, AuthorID: authorID,
		TargetID: targetID, Note: input.Note, Commentaire: input.Commentaire,
	}, nil
}

func (reviewHTTPRepositoryStub) ListUserReviews(context.Context, int) ([]domain.Review, error) {
	return []domain.Review{{ID: 1, TargetID: 2, Note: 5}}, nil
}

func (reviewHTTPRepositoryStub) ListServiceReviews(context.Context, int) ([]domain.Review, error) {
	return []domain.Review{{ID: 1, ExchangeID: 8, Note: 5}}, nil
}

func TestReviewRoutes(testContext *testing.T) {
	reviewService := application.NewReviewService(reviewHTTPRepositoryStub{})
	handler := NewHandler(
		application.UserService{},
		application.CatalogService{},
		application.ExchangeService{},
		reviewService,
	).Routes()

	tests := []struct {
		name       string
		method     string
		path       string
		userID     string
		body       string
		wantStatus int
	}{
		{
			name: "créer un avis", method: http.MethodPost, path: "/api/exchanges/8/review",
			userID: "1", body: `{"note":5,"commentaire":"Excellent"}`, wantStatus: http.StatusCreated,
		},
		{
			name: "authentification obligatoire", method: http.MethodPost, path: "/api/exchanges/8/review",
			body: `{"note":5}`, wantStatus: http.StatusUnauthorized,
		},
		{
			name: "note invalide", method: http.MethodPost, path: "/api/exchanges/8/review",
			userID: "1", body: `{"note":6}`, wantStatus: http.StatusBadRequest,
		},
		{name: "avis d'un utilisateur", method: http.MethodGet, path: "/api/users/2/reviews", wantStatus: http.StatusOK},
		{name: "avis d'un service", method: http.MethodGet, path: "/api/services/5/reviews", wantStatus: http.StatusOK},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.userID != "" {
				request.Header.Set("X-User-ID", test.userID)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				testCaseContext.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
		})
	}
}
