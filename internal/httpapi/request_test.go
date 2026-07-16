package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

func TestDecodeJSON(testContext *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{name: "JSON valide", body: `{"pseudo":"alice"}`, want: true},
		{name: "champ inconnu", body: `{"pseudo":"alice","inconnu":true}`, want: false},
		{name: "deux objets", body: `{"pseudo":"alice"} {"pseudo":"bob"}`, want: false},
		{name: "JSON mal formé", body: `{"pseudo":`, want: false},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			var input domain.CreateUserInput

			if got := decodeJSON(response, request, &input); got != test.want {
				testCaseContext.Fatalf("decodeJSON() = %v, want %v", got, test.want)
			}
			if !test.want && response.Code != http.StatusBadRequest {
				testCaseContext.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestPositiveInteger(testContext *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "42", want: true},
		{value: "0", want: false},
		{value: "-1", want: false},
		{value: "abc", want: false},
	}

	for _, test := range tests {
		testContext.Run(test.value, func(testCaseContext *testing.T) {
			_, got := positiveInteger(test.value)
			if got != test.want {
				testCaseContext.Fatalf("positiveInteger(%q) valid = %v, want %v", test.value, got, test.want)
			}
		})
	}
}
