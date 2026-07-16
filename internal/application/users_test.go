package application

import (
	"context"
	"errors"
	"testing"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type userRepositoryStub struct{}

func (userRepositoryStub) CreateUser(_ context.Context, input domain.CreateUserInput) (domain.User, error) {
	return domain.User{ID: 1, Pseudo: input.Pseudo, Bio: input.Bio, Ville: input.Ville}, nil
}

func (userRepositoryStub) ListUsers(context.Context) ([]domain.User, error) {
	return []domain.User{{ID: 1, Pseudo: "Alice"}}, nil
}

func (userRepositoryStub) FindUser(context.Context, int) (domain.User, error) {
	return domain.User{ID: 1, Pseudo: "Alice"}, nil
}

func (userRepositoryStub) UpdateUser(_ context.Context, id int, input domain.CreateUserInput) (domain.User, error) {
	return domain.User{ID: id, Pseudo: input.Pseudo}, nil
}

func (userRepositoryStub) DeleteUser(context.Context, int) error { return nil }

func (userRepositoryStub) ListSkills(context.Context, int) ([]domain.Skill, error) {
	return []domain.Skill{{Nom: domain.CategoryJardinage, Niveau: domain.SkillLevelBeginner}}, nil
}

func (userRepositoryStub) ReplaceSkills(context.Context, int, []domain.Skill) error { return nil }

func TestUserService(testContext *testing.T) {
	service := NewUserService(userRepositoryStub{})
	ctx := context.Background()

	created, err := service.Create(ctx, domain.CreateUserInput{Pseudo: " Alice ", Bio: " Bio ", Ville: " Paris "})
	if err != nil || created.Pseudo != "Alice" || created.Bio != "Bio" || created.Ville != "Paris" {
		testContext.Fatalf("Create() = %+v, %v", created, err)
	}

	users, err := service.List(ctx)
	if err != nil || len(users) != 1 {
		testContext.Fatalf("List() = %+v, %v", users, err)
	}

	user, err := service.Get(ctx, 1)
	if err != nil || user.ID != 1 {
		testContext.Fatalf("Get() = %+v, %v", user, err)
	}

	updated, err := service.Update(ctx, 1, domain.CreateUserInput{Pseudo: " Alice 2 "})
	if err != nil || updated.Pseudo != "Alice 2" {
		testContext.Fatalf("Update() = %+v, %v", updated, err)
	}

	if err := service.Delete(ctx, 1); err != nil {
		testContext.Fatalf("Delete() error = %v", err)
	}

	skills, err := service.ListSkills(ctx, 1)
	if err != nil || len(skills) != 1 {
		testContext.Fatalf("ListSkills() = %+v, %v", skills, err)
	}
}

func TestUserValidation(testContext *testing.T) {
	service := NewUserService(userRepositoryStub{})
	ctx := context.Background()

	if _, err := service.Create(ctx, domain.CreateUserInput{Pseudo: " "}); !errors.Is(err, domain.ErrPseudoRequired) {
		testContext.Fatalf("Create() error = %v, want %v", err, domain.ErrPseudoRequired)
	}
	if _, err := service.Update(ctx, 1, domain.CreateUserInput{}); !errors.Is(err, domain.ErrPseudoRequired) {
		testContext.Fatalf("Update() error = %v, want %v", err, domain.ErrPseudoRequired)
	}

	tests := []struct {
		name   string
		skills []domain.Skill
		want   error
	}{
		{name: "valide", skills: []domain.Skill{{Nom: " Jardinage ", Niveau: " DÉBUTANT "}}},
		{name: "nom vide", skills: []domain.Skill{{Niveau: domain.SkillLevelBeginner}}, want: domain.ErrSkillNameRequired},
		{name: "niveau invalide", skills: []domain.Skill{{Nom: "Jardinage", Niveau: "inconnu"}}, want: domain.ErrSkillLevelInvalid},
		{name: "doublon", skills: []domain.Skill{{Nom: "Jardinage", Niveau: domain.SkillLevelBeginner}, {Nom: "jardinage", Niveau: domain.SkillLevelExpert}}, want: domain.ErrSkillDuplicate},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			skills, err := service.ReplaceSkills(ctx, 1, test.skills)
			if !errors.Is(err, test.want) {
				testCaseContext.Fatalf("ReplaceSkills() error = %v, want %v", err, test.want)
			}
			if err == nil && (skills[0].Nom != "Jardinage" || skills[0].Niveau != domain.SkillLevelBeginner) {
				testCaseContext.Fatalf("ReplaceSkills() = %+v", skills)
			}
		})
	}
}
