package application

import (
	"errors"
	"testing"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

func TestValidateServiceInput(t *testing.T) {
	validInput := domain.CreateServiceInput{
		Titre:        "Initiation au jardinage",
		Categorie:    domain.CategoryJardinage,
		DureeMinutes: 60,
		Credits:      1,
	}

	tests := []struct {
		name  string
		input domain.CreateServiceInput
		want  error
	}{
		{name: "service valide", input: validInput},
		{name: "titre vide", input: domain.CreateServiceInput{Categorie: domain.CategoryJardinage, DureeMinutes: 60, Credits: 1}, want: domain.ErrServiceTitleRequired},
		{name: "catégorie invalide", input: domain.CreateServiceInput{Titre: "Test", Categorie: "Inconnue", DureeMinutes: 60, Credits: 1}, want: domain.ErrServiceCategoryInvalid},
		{name: "durée invalide", input: domain.CreateServiceInput{Titre: "Test", Categorie: domain.CategoryJardinage, Credits: 1}, want: domain.ErrServiceDurationInvalid},
		{name: "crédits invalides", input: domain.CreateServiceInput{Titre: "Test", Categorie: domain.CategoryJardinage, DureeMinutes: 60}, want: domain.ErrServiceCreditsInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateServiceInput(test.input); !errors.Is(err, test.want) {
				t.Fatalf("validateServiceInput() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestValidServiceCategory(t *testing.T) {
	if !validServiceCategory(domain.CategoryInformatique) {
		t.Fatal("la catégorie Informatique devrait être valide")
	}
	if validServiceCategory("Inconnue") {
		t.Fatal("une catégorie inconnue ne devrait pas être valide")
	}
}
