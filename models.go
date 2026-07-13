package main

import "time"

type User struct {
	ID            int64     `json:"id"`
	Pseudo        string    `json:"pseudo"`
	Bio           string    `json:"bio,omitempty"`
	Ville         string    `json:"ville,omitempty"`
	Skills        []Skill   `json:"skills,omitempty"`
	CreditBalance int       `json:"credit_balance"`
	CreatedAt     time.Time `json:"created_at"`
}
type Skill struct {
	Nom    string `json:"nom"`
	Niveau string `json:"niveau"`
}
type Service struct {
	ID           int64     `json:"id"`
	ProviderID   int64     `json:"provider_id"`
	Titre        string    `json:"titre"`
	Description  string    `json:"description,omitempty"`
	Categorie    string    `json:"categorie"`
	DureeMinutes int       `json:"duree_minutes"`
	Credits      int       `json:"credits"`
	Ville        string    `json:"ville,omitempty"`
	Actif        bool      `json:"actif"`
	CreatedAt    time.Time `json:"created_at"`
}
type Exchange struct {
	ID          int64     `json:"id"`
	ServiceID   int64     `json:"service_id"`
	RequesterID int64     `json:"requester_id"`
	OwnerID     int64     `json:"owner_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type Review struct {
	ID          int64     `json:"id"`
	ExchangeID  int64     `json:"exchange_id"`
	AuthorID    int64     `json:"author_id"`
	TargetID    int64     `json:"target_id"`
	Note        int       `json:"note"`
	Commentaire string    `json:"commentaire,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
type UserStats struct {
	UserID            int64   `json:"user_id"`
	ServicesActifs    int     `json:"services_actifs"`
	EchangesCompletes int     `json:"echanges_completes"`
	CreditBalance     int     `json:"credit_balance"`
	NoteMoyenne       float64 `json:"note_moyenne"`
	NbAvis            int     `json:"nb_avis"`
	TotalGagne        int     `json:"total_gagne"`
	TotalDepense      int     `json:"total_depense"`
}
type ServiceFilter struct{ Categorie, Ville, Search string }

var categories = map[string]bool{"Informatique": true, "Jardinage": true, "Bricolage": true, "Cuisine": true, "Musique": true, "Langues": true, "Sport": true, "Tutorat": true, "Déménagement": true, "Photographie": true, "Animalier": true, "Couture": true, "Autre": true}
