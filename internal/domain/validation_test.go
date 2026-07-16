package domain

import "testing"

func TestIsValidServiceCategory(testContext *testing.T) {
	tests := []struct {
		name     string
		category string
		want     bool
	}{
		{name: "catégorie connue", category: CategoryInformatique, want: true},
		{name: "catégorie inconnue", category: "Inconnue", want: false},
		{name: "catégorie vide", category: "", want: false},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			if got := IsValidServiceCategory(test.category); got != test.want {
				testCaseContext.Fatalf("IsValidServiceCategory(%q) = %v, want %v", test.category, got, test.want)
			}
		})
	}
}

func TestIsValidSkillLevel(testContext *testing.T) {
	tests := []struct {
		name  string
		level string
		want  bool
	}{
		{name: "niveau débutant", level: SkillLevelBeginner, want: true},
		{name: "niveau intermédiaire", level: SkillLevelIntermediate, want: true},
		{name: "niveau expert", level: SkillLevelExpert, want: true},
		{name: "niveau inconnu", level: "confirmé", want: false},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			if got := IsValidSkillLevel(test.level); got != test.want {
				testCaseContext.Fatalf("IsValidSkillLevel(%q) = %v, want %v", test.level, got, test.want)
			}
		})
	}
}
