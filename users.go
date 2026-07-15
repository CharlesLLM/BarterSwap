package main

import (
	"context"
	"errors"
	"strings"
)

const welcomeCredits = 10

var (
	ErrPseudoRequired      = errors.New("le pseudo est obligatoire")
	ErrPseudoAlreadyExists = errors.New("ce pseudo existe déjà")
	ErrUserNotFound        = errors.New("utilisateur introuvable")
)

type User struct {
	ID            int    `json:"id"`
	Pseudo        string `json:"pseudo"`
	Bio           string `json:"bio,omitempty"`
	Ville         string `json:"ville,omitempty"`
	CreditBalance int    `json:"credit_balance"`
	CreatedAt     string `json:"created_at"`
}

type CreateUserInput struct {
	Pseudo string `json:"pseudo"`
	Bio    string `json:"bio"`
	Ville  string `json:"ville"`
}

func CreateUser(ctx context.Context, store *Store, input CreateUserInput) (User, error) {
	input.Pseudo = strings.TrimSpace(input.Pseudo)
	input.Bio = strings.TrimSpace(input.Bio)
	input.Ville = strings.TrimSpace(input.Ville)

	if input.Pseudo == "" {
		return User{}, ErrPseudoRequired
	}

	return store.InsertUser(ctx, input)
}

func ListUsers(ctx context.Context, store *Store) ([]User, error) {
	return store.SelectUsers(ctx)
}

func DeleteUser(ctx context.Context, store *Store, id int) error {
	return store.DeleteUser(ctx, id)
}
