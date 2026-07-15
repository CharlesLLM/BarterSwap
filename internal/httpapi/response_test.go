package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

func TestStatusForError(testContext *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "validation", err: domain.ErrPseudoRequired, want: http.StatusBadRequest},
		{name: "conflit", err: domain.ErrPseudoAlreadyExists, want: http.StatusConflict},
		{name: "interdit", err: domain.ErrServiceForbidden, want: http.StatusForbidden},
		{name: "introuvable", err: domain.ErrUserNotFound, want: http.StatusNotFound},
		{name: "erreur wrappée", err: fmt.Errorf("lecture : %w", domain.ErrServiceNotFound), want: http.StatusNotFound},
		{name: "erreur interne", err: errors.New("base indisponible"), want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			if got := statusForError(test.err); got != test.want {
				testCaseContext.Fatalf("statusForError() = %d, want %d", got, test.want)
			}
		})
	}
}
