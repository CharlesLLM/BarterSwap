package main

import (
	"fmt"
	"net/http"
)

func (a *API) users(w http.ResponseWriter, r *http.Request, path []string) {
	if len(path) == 0 {
		if r.Method == http.MethodPost {
			a.createUser(w, r)
			return
		}
		a.methodNotAllowed(w)
		return
	}
	id, err := pathID(path[0])
	if err != nil {
		a.fail(w, err)
		return
	}
	if len(path) == 1 {
		switch r.Method {
		case http.MethodGet:
			a.getUser(w, r, id)
		case http.MethodPut:
			a.updateUser(w, r, id)
		default:
			a.methodNotAllowed(w)
		}
		return
	}
	if len(path) != 2 {
		a.notFound(w)
		return
	}
	switch path[1] {
	case "skills":
		a.userSkills(w, r, id)
	case "reviews":
		if r.Method != http.MethodGet {
			a.methodNotAllowed(w)
			return
		}
		a.writeReviews(w, r, "user", id)
	case "stats":
		if r.Method != http.MethodGet {
			a.methodNotAllowed(w)
			return
		}
		stats, err := a.store.Stats(r.Context(), id)
		if err != nil {
			a.fail(w, err)
			return
		}
		writeJSON(w, http.StatusOK, stats)
	default:
		a.notFound(w)
	}
}

func (a *API) createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if !decode(w, r, &user) {
		return
	}
	user.Pseudo = clean(user.Pseudo)
	if user.Pseudo == "" {
		a.fail(w, fmt.Errorf("%w: pseudo requis", ErrInvalid))
		return
	}
	created, err := a.store.CreateUser(r.Context(), user)
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *API) getUser(w http.ResponseWriter, r *http.Request, id int64) {
	user, err := a.store.User(r.Context(), id)
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (a *API) updateUser(w http.ResponseWriter, r *http.Request, id int64) {
	actor, err := authenticatedUserID(r)
	if err != nil || actor != id {
		a.fail(w, ErrForbidden)
		return
	}
	var user User
	if !decode(w, r, &user) {
		return
	}
	user.ID, user.Pseudo = id, clean(user.Pseudo)
	if user.Pseudo == "" {
		a.fail(w, fmt.Errorf("%w: pseudo requis", ErrInvalid))
		return
	}
	updated, err := a.store.UpdateUser(r.Context(), user)
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *API) userSkills(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		skills, err := a.store.Skills(r.Context(), id)
		if err != nil {
			a.fail(w, err)
			return
		}
		writeJSON(w, http.StatusOK, skills)
	case http.MethodPut:
		actor, err := authenticatedUserID(r)
		if err != nil || actor != id {
			a.fail(w, ErrForbidden)
			return
		}
		var skills []Skill
		if !decode(w, r, &skills) {
			return
		}
		for _, skill := range skills {
			if clean(skill.Nom) == "" || !validSkillLevel(skill.Niveau) {
				a.fail(w, fmt.Errorf("%w: compétence ou niveau invalide", ErrInvalid))
				return
			}
		}
		if err = a.store.ReplaceSkills(r.Context(), id, skills); err != nil {
			a.fail(w, err)
			return
		}
		writeJSON(w, http.StatusOK, skills)
	default:
		a.methodNotAllowed(w)
	}
}

func validSkillLevel(level string) bool {
	return level == "débutant" || level == "intermédiaire" || level == "expert"
}

func (a *API) writeReviews(w http.ResponseWriter, r *http.Request, resource string, id int64) {
	reviews, err := a.store.Reviews(r.Context(), resource, id)
	if err != nil {
		a.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, reviews)
}
