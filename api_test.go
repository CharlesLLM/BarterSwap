package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateService(t *testing.T) {
	tests := []struct {
		name string
		in   Service
		ok   bool
	}{
		{"valide", Service{Titre: "Cours", Categorie: "Musique", DureeMinutes: 60, Credits: 1}, true},
		{"titre vide", Service{Categorie: "Musique", DureeMinutes: 60, Credits: 1}, false},
		{"catégorie inconnue", Service{Titre: "Cours", Categorie: "Finance", DureeMinutes: 60, Credits: 1}, false},
		{"durée négative", Service{Titre: "Cours", Categorie: "Musique", DureeMinutes: -1, Credits: 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateService(tt.in) == nil; got != tt.ok {
				t.Fatalf("validateService() succès=%v, attendu %v", got, tt.ok)
			}
		})
	}
}

func TestAPIRejectsInvalidUser(t *testing.T) {
	h := NewAPI(nil, log.New(io.Discard, "", 0))
	r := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(`{"pseudo":"  "}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d, attendu 400", w.Code)
	}
}

func TestAPINotFoundAndCORS(t *testing.T) {
	h := NewAPI(nil, log.New(io.Discard, "", 0))
	r := httptest.NewRequest(http.MethodGet, "/inconnue", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("header CORS absent")
	}
}
