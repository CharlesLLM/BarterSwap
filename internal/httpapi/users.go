package httpapi

import (
	"net/http"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

func (handler Handler) usersHandler(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		handler.createUser(responseWriter, request)
	case http.MethodGet:
		handler.listUsers(responseWriter, request)
	default:
		methodNotAllowed(responseWriter, http.MethodGet, http.MethodPost)
	}
}

func (handler Handler) userHandler(responseWriter http.ResponseWriter, request *http.Request) {
	parts := pathSegments(request.URL.Path, "/api/users/")
	if len(parts) == 0 {
		writeError(responseWriter, http.StatusBadRequest, "identifiant invalide")
		return
	}

	id, valid := positiveInteger(parts[0])
	if !valid {
		writeError(responseWriter, http.StatusBadRequest, "identifiant invalide")
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
			methodNotAllowed(responseWriter, http.MethodGet, http.MethodPut, http.MethodDelete)
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
			methodNotAllowed(responseWriter, http.MethodGet, http.MethodPut)
		}
		return
	}

	writeError(responseWriter, http.StatusNotFound, "route introuvable")
}

func (handler Handler) createUser(responseWriter http.ResponseWriter, request *http.Request) {
	var input domain.CreateUserInput
	if !decodeJSON(responseWriter, request, &input) {
		return
	}

	user, err := handler.users.Create(request.Context(), input)
	if err != nil {
		writeApplicationError(responseWriter, err, "création de l'utilisateur")
		return
	}
	writeJSON(responseWriter, http.StatusCreated, user)
}

func (handler Handler) listUsers(responseWriter http.ResponseWriter, request *http.Request) {
	users, err := handler.users.List(request.Context())
	if err != nil {
		writeApplicationError(responseWriter, err, "liste des utilisateurs")
		return
	}
	writeJSON(responseWriter, http.StatusOK, users)
}

func (handler Handler) getUser(responseWriter http.ResponseWriter, request *http.Request, id int) {
	user, err := handler.users.Get(request.Context(), id)
	if err != nil {
		writeApplicationError(responseWriter, err, "lecture de l'utilisateur")
		return
	}
	writeJSON(responseWriter, http.StatusOK, user)
}

func (handler Handler) updateUser(responseWriter http.ResponseWriter, request *http.Request, id int) {
	var input domain.CreateUserInput
	if !decodeJSON(responseWriter, request, &input) {
		return
	}

	user, err := handler.users.Update(request.Context(), id, input)
	if err != nil {
		writeApplicationError(responseWriter, err, "modification de l'utilisateur")
		return
	}
	writeJSON(responseWriter, http.StatusOK, user)
}

func (handler Handler) deleteUser(responseWriter http.ResponseWriter, request *http.Request, id int) {
	if err := handler.users.Delete(request.Context(), id); err != nil {
		writeApplicationError(responseWriter, err, "suppression de l'utilisateur")
		return
	}
	responseWriter.WriteHeader(http.StatusNoContent)
}

func (handler Handler) getUserSkills(responseWriter http.ResponseWriter, request *http.Request, id int) {
	skills, err := handler.users.ListSkills(request.Context(), id)
	if err != nil {
		writeApplicationError(responseWriter, err, "lecture des compétences")
		return
	}
	writeJSON(responseWriter, http.StatusOK, skills)
}

func (handler Handler) replaceUserSkills(responseWriter http.ResponseWriter, request *http.Request, id int) {
	var skills []domain.Skill
	if !decodeJSON(responseWriter, request, &skills) {
		return
	}

	skills, err := handler.users.ReplaceSkills(request.Context(), id, skills)
	if err != nil {
		writeApplicationError(responseWriter, err, "modification des compétences")
		return
	}
	writeJSON(responseWriter, http.StatusOK, skills)
}
