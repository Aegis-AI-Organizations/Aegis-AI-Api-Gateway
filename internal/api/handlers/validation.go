package handlers

import (
	"errors"
	"unicode"
)

// ValidatePasswordComplexity ensures the password meets the following requirements:
// - Minimum 8 characters
// - At least one uppercase letter
// - At least one digit
// - At least one special character
func ValidatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return errors.New("le mot de passe doit faire au moins 8 caractères")
	}

	var hasUpper, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("le mot de passe doit contenir au moins une majuscule")
	}
	if !hasDigit {
		return errors.New("le mot de passe doit contenir au moins un chiffre")
	}
	if !hasSpecial {
		return errors.New("le mot de passe doit contenir au moins un caractère spécial")
	}

	return nil
}
