package repositories

import "errors"

const (
	UniqueViolation = "23505" // PostgreSQL error
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)
