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
		value := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/users/"), "/")
		parts := strings.Split(value, "/")
		id, err := strconv.Atoi(parts[0])

		if err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "identifiant invalide"})
			return
		}

		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				getUser(w, r, store, id)
			case http.MethodPut:
				updateUser(w, r, store, id)
			case http.MethodDelete:
				deleteUser(w, r, store, id)
			default:
				w.Header().Set("Allow", "GET, PUT, DELETE")
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
			}

			return
		}

		if len(parts) == 2 && parts[1] == "skills" {
			switch r.Method {
			case http.MethodGet:
				getUserSkills(w, r, store, id)
			case http.MethodPut:
				replaceUserSkills(w, r, store, id)
			default:
				w.Header().Set("Allow", "GET, PUT")
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
			}

			return
		}

		writeJSON(w, http.StatusNotFound, map[string]string{"error": "route introuvable"})
	}
}

func getUser(w http.ResponseWriter, r *http.Request, store *Store, id int) {
	user, err := GetUser(r.Context(), store, id)

	if err != nil {
		writeUserError(w, err, "lecture de l'utilisateur")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func updateUser(w http.ResponseWriter, r *http.Request, store *Store, id int) {
	var input CreateUserInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	user, err := UpdateUser(r.Context(), store, id, input)

	if err != nil {
		writeUserError(w, err, "modification de l'utilisateur")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func deleteUser(w http.ResponseWriter, r *http.Request, store *Store, id int) {
	if err := DeleteUser(r.Context(), store, id); err != nil {
		writeUserError(w, err, "suppression de l'utilisateur")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getUserSkills(w http.ResponseWriter, r *http.Request, store *Store, id int) {
	skills, err := GetUserSkills(r.Context(), store, id)

	if err != nil {
		writeUserError(w, err, "lecture des compétences")
		return
	}

	writeJSON(w, http.StatusOK, skills)
}

func replaceUserSkills(w http.ResponseWriter, r *http.Request, store *Store, id int) {
	var skills []Skill
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&skills); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	skills, err := ReplaceUserSkills(r.Context(), store, id, skills)

	if err != nil {
		writeUserError(w, err, "modification des compétences")
		return
	}

	writeJSON(w, http.StatusOK, skills)
}

func writeUserError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, ErrPseudoRequired),
		errors.Is(err, ErrSkillNameRequired),
		errors.Is(err, ErrSkillLevelInvalid),
		errors.Is(err, ErrSkillDuplicate):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrPseudoAlreadyExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrUserNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	default:
		log.Printf("%s : %v", action, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("écriture de la réponse JSON : %v", err)
	}
}
