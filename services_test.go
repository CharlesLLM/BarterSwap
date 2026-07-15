package main

import (
	"errors"
	"testing"
)

func TestValidateServiceInput(t *testing.T) {
	validInput := CreateServiceInput{
		Titre:        "Initiation au jardinage",
		Categorie:    CategoryJardinage,
		DureeMinutes: 60,
		Credits:      1,
	}

	tests := []struct {
		name  string
		input CreateServiceInput
		want  error
	}{
		{name: "service valide", input: validInput},
		{name: "titre vide", input: CreateServiceInput{Categorie: CategoryJardinage, DureeMinutes: 60, Credits: 1}, want: ErrServiceTitleRequired},
		{name: "catégorie invalide", input: CreateServiceInput{Titre: "Test", Categorie: "Inconnue", DureeMinutes: 60, Credits: 1}, want: ErrServiceCategoryInvalid},
		{name: "durée invalide", input: CreateServiceInput{Titre: "Test", Categorie: CategoryJardinage, Credits: 1}, want: ErrServiceDurationInvalid},
		{name: "crédits invalides", input: CreateServiceInput{Titre: "Test", Categorie: CategoryJardinage, DureeMinutes: 60}, want: ErrServiceCreditsInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateServiceInput(test.input)

			if !errors.Is(err, test.want) {
				t.Fatalf("validateServiceInput() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestValidServiceCategory(t *testing.T) {
	if !validServiceCategory(CategoryInformatique) {
		t.Fatal("la catégorie Informatique devrait être valide")
	}

	if validServiceCategory("Inconnue") {
		t.Fatal("une catégorie inconnue ne devrait pas être valide")
	}
}
