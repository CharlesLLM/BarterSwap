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

type fakeHTTPExchangeRepository struct {
	exchange domain.Exchange
}

func (repository fakeHTTPExchangeRepository) FindService(context.Context, int) (domain.Service, error) {
	return domain.Service{ID: 5, ProviderID: 2, Credits: 3}, nil
}

func (repository fakeHTTPExchangeRepository) FindUser(context.Context, int) (domain.User, error) {
	return domain.User{ID: 1}, nil
}

func (repository fakeHTTPExchangeRepository) CreditBalance(context.Context, int) (int, error) {
	return 10, nil
}

func (repository fakeHTTPExchangeRepository) CreateExchange(context.Context, int, domain.Service) (domain.Exchange, error) {
	return repository.exchange, nil
}

func (repository fakeHTTPExchangeRepository) ListExchanges(context.Context, int, domain.ExchangeFilter) ([]domain.Exchange, error) {
	return []domain.Exchange{repository.exchange}, nil
}

func (repository fakeHTTPExchangeRepository) FindExchange(context.Context, int) (domain.Exchange, error) {
	return repository.exchange, nil
}

func (repository fakeHTTPExchangeRepository) UpdateExchangeStatus(
	_ context.Context,
	_ int,
	_ string,
	newStatus string,
	_ []domain.CreditChange,
) (domain.Exchange, error) {
	exchange := repository.exchange
	exchange.Status = newStatus
	return exchange, nil
}

func TestExchangeRoutes(testContext *testing.T) {
	repository := fakeHTTPExchangeRepository{exchange: domain.Exchange{
		ID: 8, ServiceID: 5, RequesterID: 1, OwnerID: 2, Status: domain.ExchangeStatusPending, Credits: 3,
	}}
	exchangeService := application.NewExchangeService(repository)
	handler := NewHandler(application.UserService{}, application.CatalogService{}, exchangeService).Routes()

	tests := []struct {
		name       string
		method     string
		path       string
		userID     string
		body       string
		wantStatus int
	}{
		{
			name:       "création",
			method:     http.MethodPost,
			path:       "/api/exchanges",
			userID:     "1",
			body:       `{"service_id":5}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "authentification obligatoire",
			method:     http.MethodGet,
			path:       "/api/exchanges",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "filtre invalide",
			method:     http.MethodGet,
			path:       "/api/exchanges?status=inconnu",
			userID:     "1",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "acceptation",
			method:     http.MethodPut,
			path:       "/api/exchanges/8/accept",
			userID:     "2",
			wantStatus: http.StatusOK,
		},
		{
			name:       "action inconnue",
			method:     http.MethodPut,
			path:       "/api/exchanges/8/inconnue",
			userID:     "2",
			wantStatus: http.StatusNotFound,
		},
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
