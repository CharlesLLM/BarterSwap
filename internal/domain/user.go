package domain

import "errors"

const WelcomeCredits = 10

const (
	SkillLevelBeginner     = "débutant"
	SkillLevelIntermediate = "intermédiaire"
	SkillLevelExpert       = "expert"
)

var (
	ErrPseudoRequired      = errors.New("le pseudo est obligatoire")
	ErrPseudoAlreadyExists = errors.New("ce pseudo existe déjà")
	ErrUserNotFound        = errors.New("utilisateur introuvable")
	ErrSkillNameRequired   = errors.New("le nom de la compétence est obligatoire")
	ErrSkillLevelInvalid   = errors.New("le niveau doit être débutant, intermédiaire ou expert")
	ErrSkillDuplicate      = errors.New("une compétence ne peut pas être présente deux fois")
)

type User struct {
	ID            int     `json:"id"`
	Pseudo        string  `json:"pseudo"`
	Bio           string  `json:"bio,omitempty"`
	Ville         string  `json:"ville,omitempty"`
	Skills        []Skill `json:"skills,omitempty"`
	CreditBalance int     `json:"credit_balance"`
	CreatedAt     string  `json:"created_at"`
}

type Skill struct {
	Nom    string `json:"nom"`
	Niveau string `json:"niveau"`
}

type CreateUserInput struct {
	Pseudo string `json:"pseudo"`
	Bio    string `json:"bio"`
	Ville  string `json:"ville"`
}

func IsValidSkillLevel(level string) bool {
	switch level {
	case SkillLevelBeginner, SkillLevelIntermediate, SkillLevelExpert:
		return true
	default:
		return false
	}
}
