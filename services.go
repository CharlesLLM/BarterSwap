package main

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
