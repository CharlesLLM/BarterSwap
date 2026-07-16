package application

import (
	"context"
	"errors"
	"testing"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type catalogRepositoryStub struct {
	providerID int
	skills     []domain.Skill
}

func (catalogRepositoryStub) CreateService(_ context.Context, providerID int, input domain.CreateServiceInput) (domain.Service, error) {
	return domain.Service{ID: 1, ProviderID: providerID, Titre: input.Titre, Categorie: input.Categorie}, nil
}

func (catalogRepositoryStub) ListServices(_ context.Context, filter domain.ServiceFilter) ([]domain.Service, error) {
	return []domain.Service{{ID: 1, Categorie: filter.Categorie, Ville: filter.Ville}}, nil
}

func (repository catalogRepositoryStub) FindService(context.Context, int) (domain.Service, error) {
	return domain.Service{ID: 1, ProviderID: repository.providerID}, nil
}

func (catalogRepositoryStub) UpdateService(_ context.Context, id int, input domain.CreateServiceInput) (domain.Service, error) {
	return domain.Service{ID: id, ProviderID: 2, Titre: input.Titre}, nil
}

func (catalogRepositoryStub) DeactivateService(context.Context, int) error { return nil }

func (repository catalogRepositoryStub) ListSkills(context.Context, int) ([]domain.Skill, error) {
	return repository.skills, nil
}

func validServiceInput() domain.CreateServiceInput {
	return domain.CreateServiceInput{
		Titre: " Jardinage ", Description: " Entretien ", Categorie: domain.CategoryJardinage,
		DureeMinutes: 60, Credits: 2, Ville: " Paris ",
	}
}

func TestCatalogService(testContext *testing.T) {
	repository := catalogRepositoryStub{
		providerID: 2,
		skills:     []domain.Skill{{Nom: domain.CategoryJardinage, Niveau: domain.SkillLevelExpert}},
	}
	service := NewCatalogService(repository)
	ctx := context.Background()

	created, err := service.Create(ctx, 2, validServiceInput())
	if err != nil || created.Titre != "Jardinage" {
		testContext.Fatalf("Create() = %+v, %v", created, err)
	}

	services, err := service.List(ctx, domain.ServiceFilter{Categorie: " Jardinage ", Ville: " Paris "})
	if err != nil || len(services) != 1 || services[0].Ville != "Paris" {
		testContext.Fatalf("List() = %+v, %v", services, err)
	}

	if found, err := service.Get(ctx, 1); err != nil || found.ID != 1 {
		testContext.Fatalf("Get() = %+v, %v", found, err)
	}
	if updated, err := service.Update(ctx, 2, 1, validServiceInput()); err != nil || updated.ID != 1 {
		testContext.Fatalf("Update() = %+v, %v", updated, err)
	}
	if err := service.Delete(ctx, 2, 1); err != nil {
		testContext.Fatalf("Delete() error = %v", err)
	}
}

func TestCatalogRules(testContext *testing.T) {
	ctx := context.Background()
	service := NewCatalogService(catalogRepositoryStub{providerID: 2})

	if _, err := service.Create(ctx, 2, validServiceInput()); !errors.Is(err, domain.ErrServiceSkillRequired) {
		testContext.Fatalf("Create() error = %v, want %v", err, domain.ErrServiceSkillRequired)
	}
	if _, err := service.List(ctx, domain.ServiceFilter{Categorie: "inconnue"}); !errors.Is(err, domain.ErrServiceCategoryInvalid) {
		testContext.Fatalf("List() error = %v, want %v", err, domain.ErrServiceCategoryInvalid)
	}
	if _, err := service.Update(ctx, 3, 1, validServiceInput()); !errors.Is(err, domain.ErrServiceForbidden) {
		testContext.Fatalf("Update() error = %v, want %v", err, domain.ErrServiceForbidden)
	}
	if err := service.Delete(ctx, 3, 1); !errors.Is(err, domain.ErrServiceForbidden) {
		testContext.Fatalf("Delete() error = %v, want %v", err, domain.ErrServiceForbidden)
	}
}

func TestValidateServiceInput(testContext *testing.T) {
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
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			if err := validateServiceInput(test.input); !errors.Is(err, test.want) {
				testCaseContext.Fatalf("validateServiceInput() error = %v, want %v", err, test.want)
			}
		})
	}
}
