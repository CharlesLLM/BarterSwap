package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type API struct {
	store  *Store
	logger *log.Logger
}

func NewAPI(store *Store, logger *log.Logger) http.Handler {
	a := &API{store: store, logger: logger}
	return a.middleware(http.HandlerFunc(a.route))
}

type apiError struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		writeJSON(w, 400, apiError{"JSON invalide: " + err.Error()})
		return false
	}
	return true
}
func (a *API) fail(w http.ResponseWriter, err error) {
	status := 500
	switch {
	case errors.Is(err, ErrNotFound):
		status = 404
	case errors.Is(err, ErrForbidden):
		status = 403
	case errors.Is(err, ErrConflict):
		status = 409
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrInsufficientCredits):
		status = 400
	}
	if status == 500 {
		a.logger.Printf("erreur interne: %v", err)
		writeJSON(w, status, apiError{"erreur interne"})
		return
	}
	writeJSON(w, status, apiError{err.Error()})
}
func idOf(v string) (int64, error) {
	id, e := strconv.ParseInt(v, 10, 64)
	if e != nil || id < 1 {
		return 0, ErrNotFound
	}
	return id, nil
}
func userID(r *http.Request) (int64, error) {
	v := r.Header.Get("X-User-ID")
	id, e := strconv.ParseInt(v, 10, 64)
	if e != nil || id < 1 {
		return 0, fmt.Errorf("%w: header X-User-ID manquant ou invalide", ErrForbidden)
	}
	return id, nil
}
func parts(path string) []string { return strings.Split(strings.Trim(path, "/"), "/") }
func (a *API) route(w http.ResponseWriter, r *http.Request) {
	p := parts(r.URL.Path)
	if len(p) < 2 || p[0] != "api" {
		writeJSON(w, 404, apiError{"route introuvable"})
		return
	}
	if p[1] == "users" {
		a.users(w, r, p[2:])
		return
	}
	if p[1] == "services" {
		a.services(w, r, p[2:])
		return
	}
	if p[1] == "exchanges" {
		a.exchanges(w, r, p[2:])
		return
	}
	writeJSON(w, 404, apiError{"route introuvable"})
}
func (a *API) users(w http.ResponseWriter, r *http.Request, p []string) {
	if len(p) == 0 && r.Method == http.MethodPost {
		var u User
		if !decode(w, r, &u) {
			return
		}
		u.Pseudo = clean(u.Pseudo)
		if u.Pseudo == "" {
			a.fail(w, fmt.Errorf("%w: pseudo requis", ErrInvalid))
			return
		}
		v, e := a.store.CreateUser(r.Context(), u)
		if e != nil {
			a.fail(w, e)
			return
		}
		writeJSON(w, 201, v)
		return
	}
	if len(p) < 1 {
		a.method(w)
		return
	}
	id, e := idOf(p[0])
	if e != nil {
		a.fail(w, e)
		return
	}
	if len(p) == 1 {
		switch r.Method {
		case http.MethodGet:
			v, e := a.store.User(r.Context(), id)
			if e != nil {
				a.fail(w, e)
				return
			}
			writeJSON(w, 200, v)
		case http.MethodPut:
			uid, e := userID(r)
			if e != nil || uid != id {
				a.fail(w, ErrForbidden)
				return
			}
			var u User
			if !decode(w, r, &u) {
				return
			}
			u.ID = id
			u.Pseudo = clean(u.Pseudo)
			if u.Pseudo == "" {
				a.fail(w, ErrInvalid)
				return
			}
			v, e := a.store.UpdateUser(r.Context(), u)
			if e != nil {
				a.fail(w, e)
				return
			}
			writeJSON(w, 200, v)
		default:
			a.method(w)
		}
		return
	}
	if len(p) == 2 && p[1] == "skills" {
		if r.Method == http.MethodGet {
			v, e := a.store.Skills(r.Context(), id)
			if e != nil {
				a.fail(w, e)
				return
			}
			writeJSON(w, 200, v)
			return
		}
		if r.Method == http.MethodPut {
			uid, e := userID(r)
			if e != nil || uid != id {
				a.fail(w, ErrForbidden)
				return
			}
			var v []Skill
			if !decode(w, r, &v) {
				return
			}
			for _, s := range v {
				if clean(s.Nom) == "" || (s.Niveau != "débutant" && s.Niveau != "intermédiaire" && s.Niveau != "expert") {
					a.fail(w, fmt.Errorf("%w: compétence ou niveau invalide", ErrInvalid))
					return
				}
			}
			if e = a.store.ReplaceSkills(r.Context(), id, v); e != nil {
				a.fail(w, e)
				return
			}
			writeJSON(w, 200, v)
			return
		}
	}
	if len(p) == 2 && p[1] == "reviews" && r.Method == http.MethodGet {
		v, e := a.store.Reviews(r.Context(), "user", id)
		if e != nil {
			a.fail(w, e)
			return
		}
		writeJSON(w, 200, v)
		return
	}
	if len(p) == 2 && p[1] == "stats" && r.Method == http.MethodGet {
		v, e := a.store.Stats(r.Context(), id)
		if e != nil {
			a.fail(w, e)
			return
		}
		writeJSON(w, 200, v)
		return
	}
	a.method(w)
}
func validateService(v Service) error {
	if clean(v.Titre) == "" || !categories[v.Categorie] || v.DureeMinutes < 1 || v.Credits < 1 {
		return fmt.Errorf("%w: titre, catégorie, durée et crédits requis", ErrInvalid)
	}
	return nil
}
func (a *API) services(w http.ResponseWriter, r *http.Request, p []string) {
	if len(p) == 0 {
		if r.Method == http.MethodGet {
			f := ServiceFilter{r.URL.Query().Get("categorie"), r.URL.Query().Get("ville"), r.URL.Query().Get("search")}
			v, e := a.store.Services(r.Context(), f)
			if e != nil {
				a.fail(w, e)
				return
			}
			writeJSON(w, 200, v)
			return
		}
		if r.Method == http.MethodPost {
			uid, e := userID(r)
			if e != nil {
				a.fail(w, e)
				return
			}
			var v Service
			if !decode(w, r, &v) {
				return
			}
			v.ProviderID = uid
			if e = validateService(v); e != nil {
				a.fail(w, e)
				return
			}
			ok, e := a.store.HasSkill(r.Context(), uid, v.Categorie)
			if e != nil {
				a.fail(w, e)
				return
			}
			if !ok {
				a.fail(w, fmt.Errorf("%w: la catégorie doit correspondre à une compétence", ErrInvalid))
				return
			}
			v, e = a.store.CreateService(r.Context(), v)
			if e != nil {
				a.fail(w, e)
				return
			}
			writeJSON(w, 201, v)
			return
		}
		a.method(w)
		return
	}
	id, e := idOf(p[0])
	if e != nil {
		a.fail(w, e)
		return
	}
	if len(p) == 2 && p[1] == "reviews" && r.Method == http.MethodGet {
		v, e := a.store.Reviews(r.Context(), "service", id)
		if e != nil {
			a.fail(w, e)
			return
		}
		writeJSON(w, 200, v)
		return
	}
	if len(p) != 1 {
		a.method(w)
		return
	}
	v, e := a.store.Service(r.Context(), id)
	if e != nil {
		a.fail(w, e)
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, v)
	case http.MethodPut:
		uid, e := userID(r)
		if e != nil || uid != v.ProviderID {
			a.fail(w, ErrForbidden)
			return
		}
		var nv Service
		if !decode(w, r, &nv) {
			return
		}
		nv.ID = id
		nv.ProviderID = uid
		if e = validateService(nv); e != nil {
			a.fail(w, e)
			return
		}
		nv, e = a.store.UpdateService(r.Context(), nv)
		if e != nil {
			a.fail(w, e)
			return
		}
		writeJSON(w, 200, nv)
	case http.MethodDelete:
		uid, e := userID(r)
		if e != nil || uid != v.ProviderID {
			a.fail(w, ErrForbidden)
			return
		}
		if e = a.store.DeleteService(r.Context(), id); e != nil {
			a.fail(w, e)
			return
		}
		writeJSON(w, 204, nil)
	default:
		a.method(w)
	}
}
func (a *API) exchanges(w http.ResponseWriter, r *http.Request, p []string) {
	uid, e := userID(r)
	if e != nil {
		a.fail(w, e)
		return
	}
	if len(p) == 0 {
		if r.Method == http.MethodPost {
			var in struct {
				ServiceID int64 `json:"service_id"`
			}
			if !decode(w, r, &in) {
				return
			}
			v, e := a.store.CreateExchange(r.Context(), in.ServiceID, uid)
			if e != nil {
				a.fail(w, e)
				return
			}
			writeJSON(w, 201, v)
			return
		}
		if r.Method == http.MethodGet {
			status := r.URL.Query().Get("status")
			if status != "" && !validStatus(status) {
				a.fail(w, ErrInvalid)
				return
			}
			v, e := a.store.Exchanges(r.Context(), uid, status)
			if e != nil {
				a.fail(w, e)
				return
			}
			writeJSON(w, 200, v)
			return
		}
		a.method(w)
		return
	}
	id, e := idOf(p[0])
	if e != nil {
		a.fail(w, e)
		return
	}
	if len(p) == 1 && r.Method == http.MethodGet {
		v, e := a.store.Exchange(r.Context(), id)
		if e != nil {
			a.fail(w, e)
			return
		}
		if uid != v.OwnerID && uid != v.RequesterID {
			a.fail(w, ErrForbidden)
			return
		}
		writeJSON(w, 200, v)
		return
	}
	if len(p) == 2 && p[1] == "review" && r.Method == http.MethodPost {
		var v Review
		if !decode(w, r, &v) {
			return
		}
		v.ExchangeID = id
		v.AuthorID = uid
		if v.Note < 1 || v.Note > 5 {
			a.fail(w, fmt.Errorf("%w: note entre 1 et 5", ErrInvalid))
			return
		}
		v, e = a.store.CreateReview(r.Context(), v)
		if e != nil {
			a.fail(w, e)
			return
		}
		writeJSON(w, 201, v)
		return
	}
	if len(p) == 2 && r.Method == http.MethodPut {
		v, e := a.store.Transition(r.Context(), id, uid, p[1])
		if e != nil {
			a.fail(w, e)
			return
		}
		writeJSON(w, 200, v)
		return
	}
	a.method(w)
}
func (a *API) method(w http.ResponseWriter) { writeJSON(w, 405, apiError{"méthode non autorisée"}) }
func (a *API) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		start := time.Now()
		defer func() {
			if v := recover(); v != nil {
				a.logger.Printf("panic: %v", v)
				writeJSON(w, 500, apiError{"erreur interne"})
			}
			a.logger.Printf("%s %s %s", r.Method, r.URL.RequestURI(), time.Since(start))
		}()
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
