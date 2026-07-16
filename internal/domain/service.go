package domain

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
	ErrServiceNotFound        = Error{Kind: ErrorNotFound, Message: "service introuvable"}
	ErrServiceTitleRequired   = Error{Kind: ErrorValidation, Message: "le titre est obligatoire"}
	ErrServiceCategoryInvalid = Error{Kind: ErrorValidation, Message: "la catégorie est invalide"}
	ErrServiceDurationInvalid = Error{Kind: ErrorValidation, Message: "la durée doit être supérieure à zéro"}
	ErrServiceCreditsInvalid  = Error{Kind: ErrorValidation, Message: "le nombre de crédits doit être supérieur à zéro"}
	ErrServiceSkillRequired   = Error{Kind: ErrorValidation, Message: "l'utilisateur ne possède pas cette compétence"}
	ErrServiceForbidden       = Error{Kind: ErrorForbidden, Message: "vous ne pouvez pas modifier ce service"}
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

func IsValidServiceCategory(category string) bool {
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
