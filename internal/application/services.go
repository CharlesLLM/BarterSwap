package application

import (
	"context"
	"strings"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type ServiceRepository interface {
	CreateService(context.Context, int, domain.CreateServiceInput) (domain.Service, error)
	ListServices(context.Context, domain.ServiceFilter) ([]domain.Service, error)
	FindService(context.Context, int) (domain.Service, error)
	UpdateService(context.Context, int, domain.CreateServiceInput) (domain.Service, error)
	DeactivateService(context.Context, int) error
	ListSkills(context.Context, int) ([]domain.Skill, error)
}

type CatalogService struct {
	repository ServiceRepository
}

func NewCatalogService(repository ServiceRepository) *CatalogService {
	return &CatalogService{repository: repository}
}

func (service *CatalogService) Create(ctx context.Context, providerID int, input domain.CreateServiceInput) (domain.Service, error) {
	input = cleanServiceInput(input)
	if err := validateServiceInput(input); err != nil {
		return domain.Service{}, err
	}

	if err := service.checkProviderSkill(ctx, providerID, input.Categorie); err != nil {
		return domain.Service{}, err
	}

	return service.repository.CreateService(ctx, providerID, input)
}

func (service *CatalogService) List(ctx context.Context, filter domain.ServiceFilter) ([]domain.Service, error) {
	filter.Categorie = strings.TrimSpace(filter.Categorie)
	filter.Ville = strings.TrimSpace(filter.Ville)
	filter.Search = strings.TrimSpace(filter.Search)

	if filter.Categorie != "" && !domain.IsValidServiceCategory(filter.Categorie) {
		return nil, domain.ErrServiceCategoryInvalid
	}

	return service.repository.ListServices(ctx, filter)
}

func (service *CatalogService) Get(ctx context.Context, id int) (domain.Service, error) {
	return service.repository.FindService(ctx, id)
}

func (service *CatalogService) Update(ctx context.Context, userID, id int, input domain.CreateServiceInput) (domain.Service, error) {
	input = cleanServiceInput(input)
	if err := validateServiceInput(input); err != nil {
		return domain.Service{}, err
	}

	existingService, err := service.repository.FindService(ctx, id)
	if err != nil {
		return domain.Service{}, err
	}

	if existingService.ProviderID != userID {
		return domain.Service{}, domain.ErrServiceForbidden
	}

	if err := service.checkProviderSkill(ctx, userID, input.Categorie); err != nil {
		return domain.Service{}, err
	}

	return service.repository.UpdateService(ctx, id, input)
}

func (service *CatalogService) Delete(ctx context.Context, userID, id int) error {
	existingService, err := service.repository.FindService(ctx, id)
	if err != nil {
		return err
	}

	if existingService.ProviderID != userID {
		return domain.ErrServiceForbidden
	}

	return service.repository.DeactivateService(ctx, id)
}

func cleanServiceInput(input domain.CreateServiceInput) domain.CreateServiceInput {
	input.Titre = strings.TrimSpace(input.Titre)
	input.Description = strings.TrimSpace(input.Description)
	input.Categorie = strings.TrimSpace(input.Categorie)
	input.Ville = strings.TrimSpace(input.Ville)
	return input
}

func validateServiceInput(input domain.CreateServiceInput) error {
	if input.Titre == "" {
		return domain.ErrServiceTitleRequired
	}
	if !domain.IsValidServiceCategory(input.Categorie) {
		return domain.ErrServiceCategoryInvalid
	}
	if input.DureeMinutes <= 0 {
		return domain.ErrServiceDurationInvalid
	}
	if input.Credits <= 0 {
		return domain.ErrServiceCreditsInvalid
	}
	return nil
}

func (service *CatalogService) checkProviderSkill(ctx context.Context, providerID int, category string) error {
	skills, err := service.repository.ListSkills(ctx, providerID)
	if err != nil {
		return err
	}

	for _, skill := range skills {
		if strings.EqualFold(skill.Nom, category) {
			return nil
		}
	}

	return domain.ErrServiceSkillRequired
}
