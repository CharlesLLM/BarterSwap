package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func usersHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createUser(w, r, store)
		case http.MethodGet:
			listUsers(w, r, store)
		default:
			w.Header().Set("Allow", "GET, POST")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
		}
	}
}

func createUser(w http.ResponseWriter, r *http.Request, store *Store) {
	var input CreateUserInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	user, err := CreateUser(r.Context(), store, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrPseudoRequired):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrPseudoAlreadyExists):
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			log.Printf("création de l'utilisateur : %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func listUsers(w http.ResponseWriter, r *http.Request, store *Store) {
	users, err := ListUsers(r.Context(), store)
	if err != nil {
		log.Printf("liste des utilisateurs : %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func userHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.Header().Set("Allow", http.MethodDelete)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
			return
		}

		value := strings.TrimPrefix(r.URL.Path, "/api/users/")
		id, err := strconv.Atoi(value)
		if err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "identifiant invalide"})
			return
		}

		if err := DeleteUser(r.Context(), store, id); err != nil {
			if errors.Is(err, ErrUserNotFound) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
				return
			}

			log.Printf("suppression de l'utilisateur : %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("écriture de la réponse JSON : %v", err)
	}
}
