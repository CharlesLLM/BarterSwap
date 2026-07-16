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
	ErrSkillNameRequired   = errors.New("le nom de la compétence est obligatoire")
	ErrSkillLevelInvalid   = errors.New("le niveau doit être débutant, intermédiaire ou expert")
	ErrSkillDuplicate      = errors.New("une compétence ne peut pas être présente deux fois")
)

type User struct {
	ID            int     `json:"id"`
	Pseudo        string  `json:"pseudo"`
	Bio           string  `json:"bio,omitempty"`
	Ville         string  `json:"ville,omitempty"`
	Skills        []Skill `json:"skills,omitempty"`
	CreditBalance int     `json:"credit_balance"`
	CreatedAt     string  `json:"created_at"`
}

type Skill struct {
	Nom    string `json:"nom"`
	Niveau string `json:"niveau"`
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

func GetUser(ctx context.Context, store *Store, id int) (User, error) {
	return store.SelectUser(ctx, id)
}

func UpdateUser(ctx context.Context, store *Store, id int, input CreateUserInput) (User, error) {
	input.Pseudo = strings.TrimSpace(input.Pseudo)
	input.Bio = strings.TrimSpace(input.Bio)
	input.Ville = strings.TrimSpace(input.Ville)

	if input.Pseudo == "" {
		return User{}, ErrPseudoRequired
	}

	return store.UpdateUser(ctx, id, input)
}

func DeleteUser(ctx context.Context, store *Store, id int) error {
	return store.DeleteUser(ctx, id)
}

func GetUserSkills(ctx context.Context, store *Store, id int) ([]Skill, error) {
	return store.SelectSkills(ctx, id)
}

func ReplaceUserSkills(ctx context.Context, store *Store, id int, skills []Skill) ([]Skill, error) {
	seen := make(map[string]bool)

	for index := range skills {
		skills[index].Nom = strings.TrimSpace(skills[index].Nom)
		skills[index].Niveau = strings.ToLower(strings.TrimSpace(skills[index].Niveau))

		if skills[index].Nom == "" {
			return nil, ErrSkillNameRequired
		}

		if !validSkillLevel(skills[index].Niveau) {
			return nil, ErrSkillLevelInvalid
		}

		key := strings.ToLower(skills[index].Nom)
		if seen[key] {
			return nil, ErrSkillDuplicate
		}

		seen[key] = true
	}

	if err := store.ReplaceSkills(ctx, id, skills); err != nil {
		return nil, err
	}

	return skills, nil
}

func validSkillLevel(level string) bool {
	return level == "débutant" || level == "intermédiaire" || level == "expert"
}
