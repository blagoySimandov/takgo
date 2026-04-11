package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo   IUserRepository
	tokens ITokenService
}

func NewAuthService(repo IUserRepository, tokens ITokenService) *AuthService {
	return &AuthService{repo: repo, tokens: tokens}
}

func (s *AuthService) RegisterUser(ctx context.Context, username, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // note: cost is 2^cost
	if err != nil {
		return "", err
	}
	user := &User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return "", err
	}
	return s.tokens.Generate(user.ID.String())
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil { // cost is tracked inside the hash
		return "", ErrInvalidCredentials
	}
	return s.tokens.Generate(user.ID.String())
}
