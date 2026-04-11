package handlers

import (
	"github.com/blagoySimandov/takgo/internal/domain/auth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, authSvc *auth.AuthService, tokens auth.ITokenService) {
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), apiConfig())
	api.UseMiddleware(makeJWTMiddleware(api, tokens))
	registerAuth(api, authSvc)
}

func apiConfig() huma.Config {
	config := huma.DefaultConfig("TakGo API", "1.0.0") // TODO: Move to cfg or constants
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
	}
	return config
}
