package main

import (
	"context"
	"net/http"
	"time"
)

func (a *API) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		start := time.Now()
		defer func() {
			if recovered := recover(); recovered != nil {
				a.logger.Printf("panic: %v", recovered)
				writeJSON(w, http.StatusInternalServerError, apiError{"erreur interne"})
			}
			a.logger.Printf("%s %s %s", r.Method, r.URL.RequestURI(), time.Since(start))
		}()
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
