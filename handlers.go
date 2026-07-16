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
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			createUser(responseWriter, request, store)
		case http.MethodGet:
			listUsers(responseWriter, request, store)
		default:
			responseWriter.Header().Set("Allow", "GET, POST")
			writeJSON(responseWriter, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
		}
	}
}

func createUser(responseWriter http.ResponseWriter, request *http.Request, store *Store) {
	var input CreateUserInput
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	user, err := CreateUser(request.Context(), store, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrPseudoRequired):
			writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, ErrPseudoAlreadyExists):
			writeJSON(responseWriter, http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			log.Printf("création de l'utilisateur : %v", err)
			writeJSON(responseWriter, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
		}
		return
	}

	writeJSON(responseWriter, http.StatusCreated, user)
}

func listUsers(responseWriter http.ResponseWriter, request *http.Request, store *Store) {
	users, err := ListUsers(request.Context(), store)
	if err != nil {
		log.Printf("liste des utilisateurs : %v", err)
		writeJSON(responseWriter, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
		return
	}

	writeJSON(responseWriter, http.StatusOK, users)
}

func userHandler(store *Store) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		value := strings.Trim(strings.TrimPrefix(request.URL.Path, "/api/users/"), "/")
		parts := strings.Split(value, "/")
		id, err := strconv.Atoi(parts[0])

		if err != nil || id <= 0 {
			writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "identifiant invalide"})
			return
		}

		if len(parts) == 1 {
			switch request.Method {
			case http.MethodGet:
				getUser(responseWriter, request, store, id)
			case http.MethodPut:
				updateUser(responseWriter, request, store, id)
			case http.MethodDelete:
				deleteUser(responseWriter, request, store, id)
			default:
				responseWriter.Header().Set("Allow", "GET, PUT, DELETE")
				writeJSON(responseWriter, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
			}

			return
		}

		if len(parts) == 2 && parts[1] == "skills" {
			switch request.Method {
			case http.MethodGet:
				getUserSkills(responseWriter, request, store, id)
			case http.MethodPut:
				replaceUserSkills(responseWriter, request, store, id)
			default:
				responseWriter.Header().Set("Allow", "GET, PUT")
				writeJSON(responseWriter, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
			}

			return
		}

		writeJSON(responseWriter, http.StatusNotFound, map[string]string{"error": "route introuvable"})
	}
}

func getUser(responseWriter http.ResponseWriter, request *http.Request, store *Store, id int) {
	user, err := GetUser(request.Context(), store, id)

	if err != nil {
		writeUserError(responseWriter, err, "lecture de l'utilisateur")
		return
	}

	writeJSON(responseWriter, http.StatusOK, user)
}

func updateUser(responseWriter http.ResponseWriter, request *http.Request, store *Store, id int) {
	var input CreateUserInput
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	user, err := UpdateUser(request.Context(), store, id, input)

	if err != nil {
		writeUserError(responseWriter, err, "modification de l'utilisateur")
		return
	}

	writeJSON(responseWriter, http.StatusOK, user)
}

func deleteUser(responseWriter http.ResponseWriter, request *http.Request, store *Store, id int) {
	if err := DeleteUser(request.Context(), store, id); err != nil {
		writeUserError(responseWriter, err, "suppression de l'utilisateur")
		return
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}

func getUserSkills(responseWriter http.ResponseWriter, request *http.Request, store *Store, id int) {
	skills, err := GetUserSkills(request.Context(), store, id)

	if err != nil {
		writeUserError(responseWriter, err, "lecture des compétences")
		return
	}

	writeJSON(responseWriter, http.StatusOK, skills)
}

func replaceUserSkills(responseWriter http.ResponseWriter, request *http.Request, store *Store, id int) {
	var skills []Skill
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&skills); err != nil {
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	skills, err := ReplaceUserSkills(request.Context(), store, id, skills)

	if err != nil {
		writeUserError(responseWriter, err, "modification des compétences")
		return
	}

	writeJSON(responseWriter, http.StatusOK, skills)
}

func writeUserError(responseWriter http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, ErrPseudoRequired),
		errors.Is(err, ErrSkillNameRequired),
		errors.Is(err, ErrSkillLevelInvalid),
		errors.Is(err, ErrSkillDuplicate):
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrPseudoAlreadyExists):
		writeJSON(responseWriter, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrUserNotFound):
		writeJSON(responseWriter, http.StatusNotFound, map[string]string{"error": err.Error()})
	default:
		log.Printf("%s : %v", action, err)
		writeJSON(responseWriter, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
	}
}

func writeJSON(responseWriter http.ResponseWriter, status int, value interface{}) {
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(status)

	if err := json.NewEncoder(responseWriter).Encode(value); err != nil {
		log.Printf("écriture de la requestéponse JSON : %v", err)
	}
}
