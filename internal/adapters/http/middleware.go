package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/blagoySimandov/takgo/internal/domain/auth"
	"github.com/danielgtaylor/huma/v2"
)

type contextKey string

const UserIDKey contextKey = "userID"

func makeJWTMiddleware(api huma.API, tokens auth.ITokenService) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !requiresAuth(ctx) {
			next(ctx)
			return
		}
		userID, err := validateBearer(ctx.Header("Authorization"), tokens)
		if err != nil {
			if err := huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized"); err != nil {
				log.Printf("error writing unauthorized response: %v", err)
			}
			return
		}
		Enrich(ctx.Context(), "user_id", userID)
		next(huma.WithValue(ctx, UserIDKey, userID))
	}
}

func makeValidationMiddleware() func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		next(ctx)
		if status := ctx.Status(); status == 422 {
			Enrich(ctx.Context(), "validation_status", 422)
		}
	}
}

func requiresAuth(ctx huma.Context) bool {
	for _, scheme := range ctx.Operation().Security {
		if _, ok := scheme["bearer"]; ok {
			return true
		}
	}
	return false
}

func validateBearer(header string, tokens auth.ITokenService) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", auth.ErrInvalidCredentials
	}
	return tokens.Validate(parts[1])
}
