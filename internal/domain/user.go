package domain

const WelcomeCredits = 10

const (
	SkillLevelBeginner     = "débutant"
	SkillLevelIntermediate = "intermédiaire"
	SkillLevelExpert       = "expert"
)

var (
	ErrPseudoRequired      = Error{Kind: ErrorValidation, Message: "le pseudo est obligatoire"}
	ErrPseudoAlreadyExists = Error{Kind: ErrorConflict, Message: "ce pseudo existe déjà"}
	ErrUserNotFound        = Error{Kind: ErrorNotFound, Message: "utilisateur introuvable"}
	ErrSkillNameRequired   = Error{Kind: ErrorValidation, Message: "le nom de la compétence est obligatoire"}
	ErrSkillLevelInvalid   = Error{Kind: ErrorValidation, Message: "le niveau doit être débutant, intermédiaire ou expert"}
	ErrSkillDuplicate      = Error{Kind: ErrorValidation, Message: "une compétence ne peut pas être présente deux fois"}
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
