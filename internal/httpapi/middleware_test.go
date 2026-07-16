package httpapi

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestWithCORSPreflight(testContext *testing.T) {
	handler := withCORS(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(http.StatusTeapot)
	}))

	request := httptest.NewRequest(http.MethodOptions, "/api/services", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Access-Control-Request-Method", http.MethodPut)
	request.Header.Set("Access-Control-Request-Headers", "Content-Type,X-User-ID")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		testContext.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		testContext.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "*")
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "OPTIONS") {
		testContext.Fatalf("Access-Control-Allow-Methods = %q, want contain OPTIONS", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "X-User-ID") {
		testContext.Fatalf("Access-Control-Allow-Headers = %q, want contain X-User-ID", got)
	}
}

func TestWithCORSAllowsRegularRequests(testContext *testing.T) {
	handler := withCORS(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		testContext.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestWithRecoveryHandlesPanic(testContext *testing.T) {
	handler := withRecovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		testContext.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(response.Body.String(), "erreur interne") {
		testContext.Fatalf("body = %q, want contain %q", response.Body.String(), "erreur interne")
	}
}

func TestWithLoggingWritesLine(testContext *testing.T) {
	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		testContext.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
	}()

	handler := withLogging(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(http.StatusCreated)
	}))

	request := httptest.NewRequest(http.MethodPost, "/api/services", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if err := writer.Close(); err != nil {
		testContext.Fatalf("writer.Close() error = %v", err)
	}

	outputBytes, err := io.ReadAll(reader)
	if err != nil {
		testContext.Fatalf("io.ReadAll() error = %v", err)
	}

	output := string(outputBytes)
	if !strings.Contains(output, "POST /api/services -> 201") {
		testContext.Fatalf("log output = %q, want contain %q", output, "POST /api/services -> 201")
	}
}

func TestChainOrder(testContext *testing.T) {
	ordered := chain(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fmt.Fprint(responseWriter, "handler")
		}),
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				fmt.Fprint(responseWriter, "A")
				next.ServeHTTP(responseWriter, request)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				fmt.Fprint(responseWriter, "B")
				next.ServeHTTP(responseWriter, request)
			})
		},
	)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	ordered.ServeHTTP(response, request)

	if got := response.Body.String(); got != "ABhandler" {
		testContext.Fatalf("body = %q, want %q", got, "ABhandler")
	}
}
