package main

import (
	"errors"
	"testing"
)

func TestValidateServiceInput(testContext *testing.T) {
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
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			err := validateServiceInput(test.input)

			if !errors.Is(err, test.want) {
				testCaseContext.Fatalf("validateServiceInput() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestValidServiceCategory(testContext *testing.T) {
	if !validServiceCategory(CategoryInformatique) {
		testContext.Fatal("la catégorie Informatique devrait être valide")
	}

	if validServiceCategory("Inconnue") {
		testContext.Fatal("une catégorie inconnue ne devrait pas être valide")
	}
}
