package security

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrPasswordTooShort      = errors.New("password must be at least 12 characters")
	ErrPasswordTooLong       = errors.New("password must not exceed 128 characters")
	ErrPasswordNoUpper       = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower       = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit       = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial     = errors.New("password must contain at least one special character")
	ErrPasswordCommon        = errors.New("password is too common or easily guessable")
	ErrPasswordContainsEmail = errors.New("password must not contain parts of email address")
	ErrPasswordSequential    = errors.New("password must not contain sequential characters")
	ErrPasswordRepeated      = errors.New("password must not contain repeated characters")
)

type PasswordPolicy struct {
	MinLength            int
	MaxLength            int
	RequireUpper         bool
	RequireLower         bool
	RequireDigit         bool
	RequireSpecial       bool
	CheckCommonPasswords bool
	CheckSequential      bool
	CheckRepeated        bool
}

func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:            12,
		MaxLength:            128,
		RequireUpper:         true,
		RequireLower:         true,
		RequireDigit:         true,
		RequireSpecial:       true,
		CheckCommonPasswords: true,
		CheckSequential:      true,
		CheckRepeated:        true,
	}
}

func (p *PasswordPolicy) Validate(password string, email string) error {
	if len(password) < p.MinLength {
		return ErrPasswordTooShort
	}

	if len(password) > p.MaxLength {
		return ErrPasswordTooLong
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if p.RequireUpper && !hasUpper {
		return ErrPasswordNoUpper
	}

	if p.RequireLower && !hasLower {
		return ErrPasswordNoLower
	}

	if p.RequireDigit && !hasDigit {
		return ErrPasswordNoDigit
	}

	if p.RequireSpecial && !hasSpecial {
		return ErrPasswordNoSpecial
	}

	if p.CheckCommonPasswords && isCommonPassword(password) {
		return ErrPasswordCommon
	}

	if p.CheckSequential && containsSequential(password) {
		return ErrPasswordSequential
	}

	if p.CheckRepeated && containsRepeatedChars(password) {
		return ErrPasswordRepeated
	}

	if email != "" && containsEmailParts(password, email) {
		return ErrPasswordContainsEmail
	}

	return nil
}

var commonPasswords = map[string]bool{
	"password": true, "123456": true, "12345678": true, "qwerty": true,
	"abc123": true, "monkey": true, "letmein": true, "dragon": true,
	"111111": true, "baseball": true, "iloveyou": true, "trustno1": true,
	"sunshine": true, "princess": true, "admin": true, "welcome": true,
	"shadow": true, "ashley": true, "football": true, "jesus": true,
	"michael": true, "ninja": true, "mustang": true, "password1": true,
	"123456789": true, "adobe123": true, "admin123": true, "letmein1": true,
	"photoshop": true, "qazwsx": true, "qwertyuiop": true, "zaq12wsx": true,
	"password123": true, "1234567890": true, "qwerty123": true, "1q2w3e4r": true,
	"baseball1": true, "football1": true, "welcome1": true, "jesus1": true,
	"michael1": true, "ninja1": true, "mustang1": true, "access": true,
	"master": true, "love": true, "pussy": true, "696969": true,
	"qwertyui": true, "zaq1zaq1": true, "password!": true, "1234": true,
}

func isCommonPassword(password string) bool {
	lowerPassword := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(password, "")
	lowerPassword = regexp.MustCompile(`[0-9]+$`).ReplaceAllString(lowerPassword, "")

	if commonPasswords[strings.ToLower(password)] {
		return true
	}

	if commonPasswords[strings.ToLower(lowerPassword)] {
		return true
	}

	return false
}

func containsSequential(password string) bool {
	for i := 0; i < len(password)-2; i++ {
		if password[i]+1 == password[i+1] && password[i+1]+1 == password[i+2] {
			return true
		}
		if password[i]-1 == password[i+1] && password[i+1]-1 == password[i+2] {
			return true
		}
	}

	sequences := []string{"abc", "bcd", "cde", "def", "efg", "fgh", "ghi", "hij",
		"ijk", "jkl", "klm", "lmn", "mno", "nop", "opq", "pqr", "qrs", "rst",
		"stu", "tuv", "uvw", "vwx", "wxy", "xyz", "123", "234", "345", "456",
		"567", "678", "789", "890", "qwerty", "asdf", "zxcv"}

	lowerPassword := strings.ToLower(password)
	for _, seq := range sequences {
		if strings.Contains(lowerPassword, seq) {
			return true
		}
	}

	return false
}

func containsRepeatedChars(password string) bool {
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			return true
		}
	}
	return false
}

func containsEmailParts(password string, email string) bool {
	lowerPassword := strings.ToLower(password)
	lowerEmail := strings.ToLower(email)

	if strings.Contains(lowerPassword, lowerEmail) {
		return true
	}

	parts := strings.Split(lowerEmail, "@")
	if len(parts) > 0 {
		localPart := parts[0]
		if len(localPart) > 3 && strings.Contains(lowerPassword, localPart) {
			return true
		}
	}

	if len(parts) > 1 {
		domain := strings.Split(parts[1], ".")[0]
		if len(domain) > 3 && strings.Contains(lowerPassword, domain) {
			return true
		}
	}

	return false
}
