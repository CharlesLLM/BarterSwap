package main

import "errors"

var (
	ErrNotFound            = errors.New("ressource introuvable")
	ErrInvalid             = errors.New("données invalides")
	ErrForbidden           = errors.New("action interdite")
	ErrConflict            = errors.New("conflit")
	ErrInsufficientCredits = errors.New("crédits insuffisants")
)
