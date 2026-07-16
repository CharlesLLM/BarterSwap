package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CharlesLLM/BarterSwap/internal/application"
)

func TestOpenAPISpecIsEmbedded(testContext *testing.T) {
	if len(openAPISpec) == 0 {
		testContext.Fatal("le schéma OpenAPI embarqué est vide")
	}
}

func TestSwaggerRoutes(testContext *testing.T) {
	handler := NewHandler(
		application.UserService{},
		application.CatalogService{},
		application.ExchangeService{},
		application.ReviewService{},
	).Routes()

	tests := []struct {
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{method: http.MethodGet, path: "/swagger", wantStatus: http.StatusMovedPermanently},
		{method: http.MethodGet, path: "/swagger/", wantStatus: http.StatusOK, wantBody: "SwaggerUIBundle"},
		{method: http.MethodGet, path: "/openapi.yaml", wantStatus: http.StatusOK, wantBody: "/api/exchanges/{id}/review:"},
		{method: http.MethodPost, path: "/swagger", wantStatus: http.StatusMethodNotAllowed},
		{method: http.MethodPost, path: "/swagger/", wantStatus: http.StatusMethodNotAllowed},
		{method: http.MethodPost, path: "/openapi.yaml", wantStatus: http.StatusMethodNotAllowed},
	}

	for _, test := range tests {
		testContext.Run(test.path, func(testCaseContext *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				testCaseContext.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if test.wantBody != "" && !strings.Contains(response.Body.String(), test.wantBody) {
				testCaseContext.Fatalf("la réponse ne contient pas %q", test.wantBody)
			}
		})
	}
}
