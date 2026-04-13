package jwt

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewService(secret string, ttl time.Duration) *TokenService {
	return &TokenService{secret: []byte(secret), ttl: ttl}
}

func (s *TokenService) Generate(userID string) (string, error) {
	alg := jwt.SigningMethodHS256 // useing a symteric key since it's simpler. usually this would be RSA so we can verify it without the server
	claims := jwt.MapClaims{
		"sub": userID,
		"alg": alg.Name,
		"exp": time.Now().Add(s.ttl).Unix(),
	}
	return jwt.NewWithClaims(alg, claims).SignedString(s.secret)
}

func (s *TokenService) Validate(token string) (string, error) {
	t, err := jwt.Parse(token, s.keyFunc)
	if err != nil || !t.Valid {
		return "", errors.New("invalid token")
	}
	return t.Claims.GetSubject()
}

func (s *TokenService) keyFunc(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("unexpected signing method")
	}
	return s.secret, nil
}

func GetSecret() string {
	secret := os.Getenv("JWT_SECRET") // TODO: Move to cfg a centralized config package
	if secret == "" {
		secret = "secret" // TODO: Change this to something more secure
	}
	return secret
}
