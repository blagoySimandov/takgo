package auth

import "context"

type IUserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByUsername(ctx context.Context, username string) (*User, error)
}

type ITokenService interface {
	Generate(userID string) (string, error)
	Validate(token string) (userID string, err error)
}
