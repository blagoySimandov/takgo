package handlers

import (
	wsadapter "github.com/blagoySimandov/takgo/internal/adapters/ws"
	"github.com/blagoySimandov/takgo/internal/domain/auth"
	"github.com/blagoySimandov/takgo/internal/domain/game"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, authSvc *auth.AuthService, tokens auth.ITokenService, hub *game.Hub, notifier *wsadapter.WsNotifier) {
	e.Use(wideEventMiddleware())
	api := humaecho.NewWithGroup(e, e.Group("/api/v1"), apiConfig())
	api.UseMiddleware(makeJWTMiddleware(api, tokens), makeValidationMiddleware())
	registerAuth(api, authSvc)
	registerWs(e, hub, notifier, tokens)
}

func apiConfig() huma.Config {
	config := huma.DefaultConfig("TakGo API", "1.0.0") // TODO: Move to cfg or constants
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
	}
	return config
}
