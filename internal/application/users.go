package application

import (
	"context"
	"strings"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type UserRepository interface {
	InsertUser(context.Context, domain.CreateUserInput) (domain.User, error)
	SelectUsers(context.Context) ([]domain.User, error)
	SelectUser(context.Context, int) (domain.User, error)
	UpdateUser(context.Context, int, domain.CreateUserInput) (domain.User, error)
	DeleteUser(context.Context, int) error
	SelectSkills(context.Context, int) ([]domain.Skill, error)
	ReplaceSkills(context.Context, int, []domain.Skill) error
}

type UserService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository}
}

func (service *UserService) Create(ctx context.Context, input domain.CreateUserInput) (domain.User, error) {
	input = cleanUserInput(input)
	if input.Pseudo == "" {
		return domain.User{}, domain.ErrPseudoRequired
	}

	return service.repository.InsertUser(ctx, input)
}

func (service *UserService) List(ctx context.Context) ([]domain.User, error) {
	return service.repository.SelectUsers(ctx)
}

func (service *UserService) Get(ctx context.Context, id int) (domain.User, error) {
	return service.repository.SelectUser(ctx, id)
}

func (service *UserService) Update(ctx context.Context, id int, input domain.CreateUserInput) (domain.User, error) {
	input = cleanUserInput(input)
	if input.Pseudo == "" {
		return domain.User{}, domain.ErrPseudoRequired
	}

	return service.repository.UpdateUser(ctx, id, input)
}

func (service *UserService) Delete(ctx context.Context, id int) error {
	return service.repository.DeleteUser(ctx, id)
}

func (service *UserService) Skills(ctx context.Context, id int) ([]domain.Skill, error) {
	return service.repository.SelectSkills(ctx, id)
}

func (service *UserService) ReplaceSkills(ctx context.Context, id int, skills []domain.Skill) ([]domain.Skill, error) {
	seen := make(map[string]bool)

	for index := range skills {
		skills[index].Nom = strings.TrimSpace(skills[index].Nom)
		skills[index].Niveau = strings.ToLower(strings.TrimSpace(skills[index].Niveau))

		if skills[index].Nom == "" {
			return nil, domain.ErrSkillNameRequired
		}

		if !validSkillLevel(skills[index].Niveau) {
			return nil, domain.ErrSkillLevelInvalid
		}

		key := strings.ToLower(skills[index].Nom)
		if seen[key] {
			return nil, domain.ErrSkillDuplicate
		}
		seen[key] = true
	}

	if err := service.repository.ReplaceSkills(ctx, id, skills); err != nil {
		return nil, err
	}

	return skills, nil
}

func cleanUserInput(input domain.CreateUserInput) domain.CreateUserInput {
	input.Pseudo = strings.TrimSpace(input.Pseudo)
	input.Bio = strings.TrimSpace(input.Bio)
	input.Ville = strings.TrimSpace(input.Ville)
	return input
}

func validSkillLevel(level string) bool {
	return level == "débutant" || level == "intermédiaire" || level == "expert"
}
