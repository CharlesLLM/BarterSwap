package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

func decodeJSON(responseWriter http.ResponseWriter, request *http.Request, destination any) bool {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		writeError(responseWriter, http.StatusBadRequest, "JSON invalide")
		return false
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(responseWriter, http.StatusBadRequest, "JSON invalide")
		return false
	}

	return true
}

func positiveInteger(value string) (int, bool) {
	number, err := strconv.Atoi(value)
	return number, err == nil && number > 0
}

func requireUserID(responseWriter http.ResponseWriter, request *http.Request) (int, bool) {
	userID, valid := positiveInteger(request.Header.Get("X-User-ID"))
	if !valid {
		writeError(responseWriter, http.StatusUnauthorized, "header X-User-ID invalide")
	}
	return userID, valid
}

func requireUserMatch(responseWriter http.ResponseWriter, request *http.Request, targetUserID int) bool {
	userID, valid := requireUserID(responseWriter, request)
	if !valid {
		return false
	}
	if userID != targetUserID {
		writeError(responseWriter, http.StatusForbidden, "action interdite")
		return false
	}
	return true
}
