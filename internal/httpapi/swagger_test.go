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
	).Routes()

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{path: "/swagger", wantStatus: http.StatusMovedPermanently},
		{path: "/swagger/", wantStatus: http.StatusOK, wantBody: "SwaggerUIBundle"},
		{path: "/openapi.yaml", wantStatus: http.StatusOK, wantBody: "openapi: 3.0.3"},
	}

	for _, test := range tests {
		testContext.Run(test.path, func(testCaseContext *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
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
