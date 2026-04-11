package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserExists         = errors.New("username taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}
