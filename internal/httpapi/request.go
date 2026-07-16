package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func pathSegments(path, prefix string) []string {
	value := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if value == "" {
		return nil
	}
	return strings.Split(value, "/")
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
