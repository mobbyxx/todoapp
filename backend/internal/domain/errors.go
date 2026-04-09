package domain

import "errors"

var (
	ErrNotFound                 = errors.New("resource not found")
	ErrUnauthorized             = errors.New("unauthorized")
	ErrForbidden                = errors.New("forbidden")
	ErrValidation               = errors.New("validation error")
	ErrInternal                 = errors.New("internal server error")
	ErrConflictDetected         = errors.New("conflict detected during sync")
	ErrBadgeNotFound            = errors.New("badge not found")
	ErrInsufficientFreezeTokens = errors.New("insufficient freeze tokens")
)
