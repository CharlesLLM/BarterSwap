package httpapi

import (
	"net/http"
	"time"
)

type middleware func(http.Handler) http.Handler

func chain(handler http.Handler, middlewares ...middleware) http.Handler {
	wrapped := handler
	for index := len(middlewares) - 1; index >= 0; index-- {
		wrapped = middlewares[index](wrapped)
	}
	return wrapped
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		writer := &statusWriter{ResponseWriter: responseWriter, statusCode: http.StatusOK}

		next.ServeHTTP(writer, request)

		logger.Info(
			"requête HTTP",
			"method", request.Method,
			"path", request.URL.Path,
			"status", writer.statusCode,
			"duration", time.Since(startedAt),
		)
	})
}

func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic pendant la requête HTTP", "method", request.Method, "path", request.URL.Path, "panic", recovered)
				writeError(responseWriter, http.StatusInternalServerError, "erreur interne")
			}
		}()

		next.ServeHTTP(responseWriter, request)
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Set("Access-Control-Allow-Origin", "*")
		responseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		responseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")

		if request.Method == http.MethodOptions {
			responseWriter.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(responseWriter, request)
	})
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (writer *statusWriter) WriteHeader(statusCode int) {
	writer.statusCode = statusCode
	writer.ResponseWriter.WriteHeader(statusCode)
}
