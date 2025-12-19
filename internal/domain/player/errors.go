package player

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPhone       = errors.New("invalid phone")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrInvalidGender      = errors.New("invalid gender")
	ErrInvalidCountryCode = errors.New("invalid country_code")
	ErrInvalidTimeZone    = errors.New("invalid time_zone")
	ErrInvalidLocale      = errors.New("invalid locale")

	ErrNotFound   = errors.New("player not found")
	ErrConflict   = errors.New("conflict")
	ErrForbidden  = errors.New("forbidden")
	ErrValidation = errors.New("validation error")
)
