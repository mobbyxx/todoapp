package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User-related errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidEmail      = errors.New("invalid email address")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidDisplayName = errors.New("invalid display name")
	ErrPasswordTooWeak   = errors.New("password must be at least 8 characters")
	ErrEmailTaken        = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// User represents a user entity in the system
type User struct {
	ID                uuid.UUID      `json:"id"`
	Email             string         `json:"email"`
	PasswordHash      string         `json:"-"` // Never expose in JSON
	DisplayName       string         `json:"display_name"`
	AvatarURL         *string        `json:"avatar_url,omitempty"`
	IsActive          bool           `json:"is_active"`
	EmailVerifiedAt   *time.Time     `json:"email_verified_at,omitempty"`
	LastSeenAt        *time.Time     `json:"last_seen_at,omitempty"`
	CurrentLevelID    *uuid.UUID     `json:"current_level_id,omitempty"`
	TotalPoints       int            `json:"total_points"`
	StreakCount       int            `json:"streak_count"`
	StreakFreezeTokens int           `json:"streak_freeze_tokens"`
	LastStreakDate    *time.Time     `json:"last_streak_date,omitempty"`
	Preferences       map[string]interface{} `json:"preferences,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// UserProfile represents the public profile of a user (excludes sensitive fields)
type UserProfile struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	IsActive     bool   `json:"is_active"`
	TotalPoints  int    `json:"total_points"`
	StreakCount  int    `json:"streak_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// ToProfile converts a User to a UserProfile
func (u *User) ToProfile() UserProfile {
	profile := UserProfile{
		ID:          u.ID.String(),
		Email:       u.Email,
		DisplayName: u.DisplayName,
		IsActive:    u.IsActive,
		TotalPoints: u.TotalPoints,
		StreakCount: u.StreakCount,
		CreatedAt:   u.CreatedAt,
	}
	
	if u.AvatarURL != nil {
		profile.AvatarURL = *u.AvatarURL
	}
	
	return profile
}

// IsEmailVerified returns true if the user's email is verified
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// UpdateLastSeen updates the user's last seen timestamp
func (u *User) UpdateLastSeen() {
	now := time.Now()
	u.LastSeenAt = &now
}

// UserRepository defines the interface for user persistence
type UserRepository interface {
	Create(user *User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
	UpdateLastSeen(id uuid.UUID) error
}

// UserService defines the interface for user business logic
type UserService interface {
	Register(email, password, displayName string) (*User, error)
	Login(email, password string) (*User, error)
	GetUser(id uuid.UUID) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateProfile(id uuid.UUID, displayName string) (*User, error)
	UpdateLastSeen(id uuid.UUID) error
	SoftDelete(id uuid.UUID) error
}

// RegistrationRequest represents a user registration request
type RegistrationRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=50"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" validate:"omitempty,min=2,max=50"`
	AvatarURL   string `json:"avatar_url,omitempty" validate:"omitempty,url"`
}
