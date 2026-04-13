package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/blagoySimandov/takgo/internal/domain/auth"
	"github.com/danielgtaylor/huma/v2"
)

type credentialsInput struct {
	Body struct {
		Username string `json:"username" minLength:"3" maxLength:"50"`
		Password string `json:"password" minLength:"8"`
	}
}

type tokenOutput struct {
	Body struct {
		Token string `json:"token"`
	}
}

func registerAuth(api huma.API, svc *auth.AuthService) {
	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      http.MethodPost,
		Path:        "/register",
		Summary:     "Register a new user",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *credentialsInput) (*tokenOutput, error) {
		Enrich(ctx, "username", input.Body.Username)
		out, err := tokenResponse(svc.RegisterUser(ctx, input.Body.Username, input.Body.Password))
		if err != nil {
			Enrich(ctx, "auth_error", err.Error())
		}
		return out, err
	})

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/login",
		Summary:     "Login",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *credentialsInput) (*tokenOutput, error) {
		Enrich(ctx, "username", input.Body.Username)
		out, err := tokenResponse(svc.Login(ctx, input.Body.Username, input.Body.Password))
		if err != nil {
			Enrich(ctx, "auth_error", err.Error())
		}
		return out, err
	})
}

func tokenResponse(token string, err error) (*tokenOutput, error) {
	if errors.Is(err, auth.ErrUserExists) {
		return nil, huma.Error409Conflict("username taken")
	}
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return nil, huma.Error401Unauthorized("invalid credentials")
	}
	if err != nil {
		return nil, err
	}
	out := &tokenOutput{}
	out.Body.Token = token
	return out, nil
}
