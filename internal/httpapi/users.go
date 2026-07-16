package httpapi

import (
	"net/http"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type userProfileResponse struct {
	ID            int            `json:"id"`
	Pseudo        string         `json:"pseudo"`
	Bio           string         `json:"bio,omitempty"`
	Ville         string         `json:"ville,omitempty"`
	Skills        []domain.Skill `json:"skills,omitempty"`
	CreditBalance *int           `json:"credit_balance,omitempty"`
	CreatedAt     *string        `json:"created_at,omitempty"`
}

func publicUserProfile(user domain.User, includePrivate bool) userProfileResponse {
	response := userProfileResponse{
		ID:     user.ID,
		Pseudo: user.Pseudo,
		Bio:    user.Bio,
		Ville:  user.Ville,
		Skills: user.Skills,
	}
	if includePrivate {
		response.CreditBalance = &user.CreditBalance
		response.CreatedAt = &user.CreatedAt
	}
	return response
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

	requestUserID, valid := positiveInteger(request.Header.Get("X-User-ID"))
	writeJSON(responseWriter, http.StatusOK, publicUserProfile(user, valid && requestUserID == id))
}

func (handler Handler) updateUser(responseWriter http.ResponseWriter, request *http.Request, id int) {
	if !requireUserMatch(responseWriter, request, id) {
		return
	}

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
	if !requireUserMatch(responseWriter, request, id) {
		return
	}

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

func (handler Handler) getUserStats(responseWriter http.ResponseWriter, request *http.Request, id int) {
	if !requireUserMatch(responseWriter, request, id) {
		return
	}

	stats, err := handler.users.Stats(request.Context(), id)
	if err != nil {
		writeApplicationError(responseWriter, err, "lecture des statistiques utilisateur")
		return
	}
	writeJSON(responseWriter, http.StatusOK, stats)
}
