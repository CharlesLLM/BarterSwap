package main

import (
	"context"
	"errors"
	"strings"
)

const (
	CategoryInformatique = "Informatique"
	CategoryJardinage    = "Jardinage"
	CategoryBricolage    = "Bricolage"
	CategoryCuisine      = "Cuisine"
	CategoryMusique      = "Musique"
	CategoryLangues      = "Langues"
	CategorySport        = "Sport"
	CategoryTutorat      = "Tutorat"
	CategoryDemenagement = "Déménagement"
	CategoryPhotographie = "Photographie"
	CategoryAnimalier    = "Animalier"
	CategoryCouture      = "Couture"
	CategoryAutre        = "Autre"
)

type Service struct {
	ID           int    `json:"id"`
	ProviderID   int    `json:"provider_id"`
	Titre        string `json:"titre"`
	Description  string `json:"description,omitempty"`
	Categorie    string `json:"categorie"`
	DureeMinutes int    `json:"duree_minutes"`
	Credits      int    `json:"credits"`
	Ville        string `json:"ville,omitempty"`
	Actif        bool   `json:"actif"`
	CreatedAt    string `json:"created_at"`
}

var (
	ErrServiceNotFound        = errors.New("service introuvable")
	ErrServiceTitleRequired   = errors.New("le titre est obligatoire")
	ErrServiceCategoryInvalid = errors.New("la catégorie est invalide")
	ErrServiceDurationInvalid = errors.New("la durée doit être supérieure à zéro")
	ErrServiceCreditsInvalid  = errors.New("le nombre de crédits doit être supérieur à zéro")
	ErrServiceSkillRequired   = errors.New("l'utilisateur ne possède pas cette compétence")
	ErrServiceForbidden       = errors.New("vous ne pouvez pas modifier ce service")
)

type CreateServiceInput struct {
	Titre        string `json:"titre"`
	Description  string `json:"description"`
	Categorie    string `json:"categorie"`
	DureeMinutes int    `json:"duree_minutes"`
	Credits      int    `json:"credits"`
	Ville        string `json:"ville"`
}

type ServiceFilter struct {
	Categorie string
	Ville     string
	Search    string
}

func CreateService(ctx context.Context, store *Store, providerID int, input CreateServiceInput) (Service, error) {
	input = cleanServiceInput(input)

	if err := validateServiceInput(input); err != nil {
		return Service{}, err
	}

	if err := checkProviderSkill(ctx, store, providerID, input.Categorie); err != nil {
		return Service{}, err
	}

	return store.InsertService(ctx, providerID, input)
}

func ListServices(ctx context.Context, store *Store, filter ServiceFilter) ([]Service, error) {
	filter.Categorie = strings.TrimSpace(filter.Categorie)
	filter.Ville = strings.TrimSpace(filter.Ville)
	filter.Search = strings.TrimSpace(filter.Search)

	if filter.Categorie != "" && !validServiceCategory(filter.Categorie) {
		return nil, ErrServiceCategoryInvalid
	}

	return store.SelectServices(ctx, filter)
}

func GetService(ctx context.Context, store *Store, id int) (Service, error) {
	return store.SelectService(ctx, id)
}

func UpdateService(ctx context.Context, store *Store, userID, id int, input CreateServiceInput) (Service, error) {
	input = cleanServiceInput(input)

	if err := validateServiceInput(input); err != nil {
		return Service{}, err
	}

	service, err := store.SelectService(ctx, id)

	if err != nil {
		return Service{}, err
	}

	if service.ProviderID != userID {
		return Service{}, ErrServiceForbidden
	}

	if err := checkProviderSkill(ctx, store, userID, input.Categorie); err != nil {
		return Service{}, err
	}

	return store.UpdateService(ctx, id, input)
}

func DeleteService(ctx context.Context, store *Store, userID, id int) error {
	service, err := store.SelectService(ctx, id)

	if err != nil {
		return err
	}

	if service.ProviderID != userID {
		return ErrServiceForbidden
	}

	return store.DeactivateService(ctx, id)
}

func cleanServiceInput(input CreateServiceInput) CreateServiceInput {
	input.Titre = strings.TrimSpace(input.Titre)
	input.Description = strings.TrimSpace(input.Description)
	input.Categorie = strings.TrimSpace(input.Categorie)
	input.Ville = strings.TrimSpace(input.Ville)

	return input
}

func validateServiceInput(input CreateServiceInput) error {
	if input.Titre == "" {
		return ErrServiceTitleRequired
	}

	if !validServiceCategory(input.Categorie) {
		return ErrServiceCategoryInvalid
	}

	if input.DureeMinutes <= 0 {
		return ErrServiceDurationInvalid
	}

	if input.Credits <= 0 {
		return ErrServiceCreditsInvalid
	}

	return nil
}

func validServiceCategory(category string) bool {
	switch category {
	case CategoryInformatique,
		CategoryJardinage,
		CategoryBricolage,
		CategoryCuisine,
		CategoryMusique,
		CategoryLangues,
		CategorySport,
		CategoryTutorat,
		CategoryDemenagement,
		CategoryPhotographie,
		CategoryAnimalier,
		CategoryCouture,
		CategoryAutre:
		return true
	default:
		return false
	}
}

func checkProviderSkill(ctx context.Context, store *Store, providerID int, category string) error {
	skills, err := store.SelectSkills(ctx, providerID)

	if err != nil {
		return err
	}

	for _, skill := range skills {
		if strings.EqualFold(skill.Nom, category) {
			return nil
		}
	}

	return ErrServiceSkillRequired
}
