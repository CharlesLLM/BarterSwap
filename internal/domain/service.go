package domain

import "errors"

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

var (
	ErrServiceNotFound        = errors.New("service introuvable")
	ErrServiceTitleRequired   = errors.New("le titre est obligatoire")
	ErrServiceCategoryInvalid = errors.New("la catégorie est invalide")
	ErrServiceDurationInvalid = errors.New("la durée doit être supérieure à zéro")
	ErrServiceCreditsInvalid  = errors.New("le nombre de crédits doit être supérieur à zéro")
	ErrServiceSkillRequired   = errors.New("l'utilisateur ne possède pas cette compétence")
	ErrServiceForbidden       = errors.New("vous ne pouvez pas modifier ce service")
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
