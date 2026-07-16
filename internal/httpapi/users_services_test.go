package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CharlesLLM/BarterSwap/internal/application"
	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type userHTTPRepositoryStub struct{}

func (userHTTPRepositoryStub) CreateUser(_ context.Context, input domain.CreateUserInput) (domain.User, error) {
	return domain.User{ID: 1, Pseudo: input.Pseudo}, nil
}
func (userHTTPRepositoryStub) ListUsers(context.Context) ([]domain.User, error) {
	return []domain.User{{ID: 1, Pseudo: "Alice"}}, nil
}
func (userHTTPRepositoryStub) FindUser(context.Context, int) (domain.User, error) {
	return domain.User{ID: 1, Pseudo: "Alice", CreditBalance: 10, CreatedAt: "2026-01-01T10:00:00Z"}, nil
}
func (userHTTPRepositoryStub) GetUserStats(context.Context, int) (domain.UserStats, error) {
	return domain.UserStats{UserID: 1, CreditBalance: 10}, nil
}
func (userHTTPRepositoryStub) UpdateUser(_ context.Context, id int, input domain.CreateUserInput) (domain.User, error) {
	return domain.User{ID: id, Pseudo: input.Pseudo}, nil
}
func (userHTTPRepositoryStub) DeleteUser(context.Context, int) error { return nil }
func (userHTTPRepositoryStub) ListSkills(context.Context, int) ([]domain.Skill, error) {
	return []domain.Skill{{Nom: domain.CategoryJardinage, Niveau: domain.SkillLevelExpert}}, nil
}
func (userHTTPRepositoryStub) ReplaceSkills(context.Context, int, []domain.Skill) error { return nil }

type serviceHTTPRepositoryStub struct{}

func (serviceHTTPRepositoryStub) CreateService(_ context.Context, providerID int, input domain.CreateServiceInput) (domain.Service, error) {
	return domain.Service{ID: 1, ProviderID: providerID, Titre: input.Titre}, nil
}
func (serviceHTTPRepositoryStub) ListServices(context.Context, domain.ServiceFilter) ([]domain.Service, error) {
	return []domain.Service{{ID: 1, Titre: "Jardinage"}}, nil
}
func (serviceHTTPRepositoryStub) FindService(context.Context, int) (domain.Service, error) {
	return domain.Service{ID: 1, ProviderID: 1, Titre: "Jardinage"}, nil
}
func (serviceHTTPRepositoryStub) UpdateService(_ context.Context, id int, input domain.CreateServiceInput) (domain.Service, error) {
	return domain.Service{ID: id, ProviderID: 1, Titre: input.Titre}, nil
}
func (serviceHTTPRepositoryStub) DeactivateService(context.Context, int) error { return nil }
func (serviceHTTPRepositoryStub) ListSkills(context.Context, int) ([]domain.Skill, error) {
	return []domain.Skill{{Nom: domain.CategoryJardinage, Niveau: domain.SkillLevelExpert}}, nil
}

func TestUserAndServiceRoutes(testContext *testing.T) {
	handler := NewHandler(
		application.NewUserService(userHTTPRepositoryStub{}),
		application.NewCatalogService(serviceHTTPRepositoryStub{}),
		application.ExchangeService{},
		application.ReviewService{},
	).Routes()
	serviceBody := `{"titre":"Jardinage","categorie":"Jardinage","duree_minutes":60,"credits":2}`

	tests := []struct {
		name, method, path, userID, body string
		wantStatus                       int
	}{
		{name: "créer utilisateur", method: http.MethodPost, path: "/api/users", body: `{"pseudo":"Alice"}`, wantStatus: http.StatusCreated},
		{name: "lister utilisateurs", method: http.MethodGet, path: "/api/users", wantStatus: http.StatusOK},
		{name: "détail utilisateur", method: http.MethodGet, path: "/api/users/1", wantStatus: http.StatusOK},
		{name: "modifier utilisateur", method: http.MethodPut, path: "/api/users/1", userID: "1", body: `{"pseudo":"Alice 2"}`, wantStatus: http.StatusOK},
		{name: "modifier utilisateur sans header", method: http.MethodPut, path: "/api/users/1", body: `{"pseudo":"Alice 2"}`, wantStatus: http.StatusUnauthorized},
		{name: "modifier utilisateur interdit", method: http.MethodPut, path: "/api/users/1", userID: "2", body: `{"pseudo":"Alice 2"}`, wantStatus: http.StatusForbidden},
		{name: "supprimer utilisateur", method: http.MethodDelete, path: "/api/users/1", wantStatus: http.StatusNoContent},
		{name: "lister compétences", method: http.MethodGet, path: "/api/users/1/skills", wantStatus: http.StatusOK},
		{name: "modifier compétences", method: http.MethodPut, path: "/api/users/1/skills", userID: "1", body: `[{"nom":"Jardinage","niveau":"expert"}]`, wantStatus: http.StatusOK},
		{name: "modifier compétences sans header", method: http.MethodPut, path: "/api/users/1/skills", body: `[{"nom":"Jardinage","niveau":"expert"}]`, wantStatus: http.StatusUnauthorized},
		{name: "modifier compétences interdit", method: http.MethodPut, path: "/api/users/1/skills", userID: "2", body: `[{"nom":"Jardinage","niveau":"expert"}]`, wantStatus: http.StatusForbidden},
		{name: "stats utilisateur", method: http.MethodGet, path: "/api/users/1/stats", userID: "1", wantStatus: http.StatusOK},
		{name: "stats utilisateur sans header", method: http.MethodGet, path: "/api/users/1/stats", wantStatus: http.StatusUnauthorized},
		{name: "stats utilisateur interdit", method: http.MethodGet, path: "/api/users/1/stats", userID: "2", wantStatus: http.StatusForbidden},
		{name: "méthode utilisateur refusée", method: http.MethodPatch, path: "/api/users", wantStatus: http.StatusMethodNotAllowed},
		{name: "identifiant utilisateur invalide", method: http.MethodGet, path: "/api/users/abc", wantStatus: http.StatusBadRequest},
		{name: "lister services", method: http.MethodGet, path: "/api/services", wantStatus: http.StatusOK},
		{name: "créer service", method: http.MethodPost, path: "/api/services", userID: "1", body: serviceBody, wantStatus: http.StatusCreated},
		{name: "détail service", method: http.MethodGet, path: "/api/services/1", wantStatus: http.StatusOK},
		{name: "modifier service", method: http.MethodPut, path: "/api/services/1", userID: "1", body: serviceBody, wantStatus: http.StatusOK},
		{name: "supprimer service", method: http.MethodDelete, path: "/api/services/1", userID: "1", wantStatus: http.StatusNoContent},
		{name: "méthode service refusée", method: http.MethodPatch, path: "/api/services", wantStatus: http.StatusMethodNotAllowed},
		{name: "route service inconnue", method: http.MethodGet, path: "/api/services/1/inconnue", wantStatus: http.StatusNotFound},
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

func TestUserProfileHidesPrivateFieldsWhenNotOwner(testContext *testing.T) {
	handler := NewHandler(
		application.NewUserService(userHTTPRepositoryStub{}),
		application.NewCatalogService(serviceHTTPRepositoryStub{}),
		application.ExchangeService{},
		application.ReviewService{},
	).Routes()

	tests := []struct {
		name                 string
		userID               string
		wantPrivateFieldsSet bool
	}{
		{name: "anonyme", wantPrivateFieldsSet: false},
		{name: "autre utilisateur", userID: "2", wantPrivateFieldsSet: false},
		{name: "propriétaire", userID: "1", wantPrivateFieldsSet: true},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
			if test.userID != "" {
				request.Header.Set("X-User-ID", test.userID)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				testCaseContext.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}

			var payload map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				testCaseContext.Fatalf("JSON invalide: %v", err)
			}

			_, hasCredit := payload["credit_balance"]
			_, hasCreatedAt := payload["created_at"]
			if hasCredit != test.wantPrivateFieldsSet || hasCreatedAt != test.wantPrivateFieldsSet {
				testCaseContext.Fatalf(
					"champs privés présents = (%v, %v), attendu = %v; body = %s",
					hasCredit,
					hasCreatedAt,
					test.wantPrivateFieldsSet,
					response.Body.String(),
				)
			}
		})
	}
}
