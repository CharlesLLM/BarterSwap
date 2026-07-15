package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

func (handler *Handler) usersHandler(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		handler.createUser(responseWriter, request)
	case http.MethodGet:
		handler.listUsers(responseWriter, request)
	default:
		responseWriter.Header().Set("Allow", "GET, POST")
		writeJSON(responseWriter, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
	}
}

func (handler *Handler) userHandler(responseWriter http.ResponseWriter, request *http.Request) {
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
			handler.getUser(responseWriter, request, id)
		case http.MethodPut:
			handler.updateUser(responseWriter, request, id)
		case http.MethodDelete:
			handler.deleteUser(responseWriter, request, id)
		default:
			responseWriter.Header().Set("Allow", "GET, PUT, DELETE")
			writeJSON(responseWriter, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
		}
		return
	}

	if len(parts) == 2 && parts[1] == "skills" {
		switch request.Method {
		case http.MethodGet:
			handler.getUserSkills(responseWriter, request, id)
		case http.MethodPut:
			handler.replaceUserSkills(responseWriter, request, id)
		default:
			responseWriter.Header().Set("Allow", "GET, PUT")
			writeJSON(responseWriter, http.StatusMethodNotAllowed, map[string]string{"error": "méthode non autorisée"})
		}
		return
	}

	writeJSON(responseWriter, http.StatusNotFound, map[string]string{"error": "route introuvable"})
}

func (handler *Handler) createUser(responseWriter http.ResponseWriter, request *http.Request) {
	var input domain.CreateUserInput
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	user, err := handler.users.Create(request.Context(), input)
	if err != nil {
		writeUserError(responseWriter, err, "création de l'utilisateur")
		return
	}
	writeJSON(responseWriter, http.StatusCreated, user)
}

func (handler *Handler) listUsers(responseWriter http.ResponseWriter, request *http.Request) {
	users, err := handler.users.List(request.Context())
	if err != nil {
		log.Printf("liste des utilisateurs : %v", err)
		writeJSON(responseWriter, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
		return
	}
	writeJSON(responseWriter, http.StatusOK, users)
}

func (handler *Handler) getUser(responseWriter http.ResponseWriter, request *http.Request, id int) {
	user, err := handler.users.Get(request.Context(), id)
	if err != nil {
		writeUserError(responseWriter, err, "lecture de l'utilisateur")
		return
	}
	writeJSON(responseWriter, http.StatusOK, user)
}

func (handler *Handler) updateUser(responseWriter http.ResponseWriter, request *http.Request, id int) {
	var input domain.CreateUserInput
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	user, err := handler.users.Update(request.Context(), id, input)
	if err != nil {
		writeUserError(responseWriter, err, "modification de l'utilisateur")
		return
	}
	writeJSON(responseWriter, http.StatusOK, user)
}

func (handler *Handler) deleteUser(responseWriter http.ResponseWriter, request *http.Request, id int) {
	if err := handler.users.Delete(request.Context(), id); err != nil {
		writeUserError(responseWriter, err, "suppression de l'utilisateur")
		return
	}
	responseWriter.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) getUserSkills(responseWriter http.ResponseWriter, request *http.Request, id int) {
	skills, err := handler.users.Skills(request.Context(), id)
	if err != nil {
		writeUserError(responseWriter, err, "lecture des compétences")
		return
	}
	writeJSON(responseWriter, http.StatusOK, skills)
}

func (handler *Handler) replaceUserSkills(responseWriter http.ResponseWriter, request *http.Request, id int) {
	var skills []domain.Skill
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&skills); err != nil {
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": "JSON invalide"})
		return
	}

	skills, err := handler.users.ReplaceSkills(request.Context(), id, skills)
	if err != nil {
		writeUserError(responseWriter, err, "modification des compétences")
		return
	}
	writeJSON(responseWriter, http.StatusOK, skills)
}

func writeUserError(responseWriter http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, domain.ErrPseudoRequired),
		errors.Is(err, domain.ErrSkillNameRequired),
		errors.Is(err, domain.ErrSkillLevelInvalid),
		errors.Is(err, domain.ErrSkillDuplicate):
		writeJSON(responseWriter, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrPseudoAlreadyExists):
		writeJSON(responseWriter, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrUserNotFound):
		writeJSON(responseWriter, http.StatusNotFound, map[string]string{"error": err.Error()})
	default:
		log.Printf("%s : %v", action, err)
		writeJSON(responseWriter, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
	}
}
