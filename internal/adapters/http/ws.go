package handlers

import (
	"net/http"

	wsadapter "github.com/blagoySimandov/takgo/internal/adapters/ws"
	"github.com/blagoySimandov/takgo/internal/domain/auth"
	"github.com/blagoySimandov/takgo/internal/domain/game"
	"github.com/blagoySimandov/takgo/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type moveMsg struct {
	Position int `json:"position"`
}

// NOTE: bypassing huma since huma doesn't support websockets yet will generate the docs with AsyncAPI
func registerWs(e *echo.Echo, hub *game.Hub, notifier *wsadapter.WsNotifier, tokens auth.ITokenService) {
	e.GET("/api/v1/game/connect", makeWsHandler(hub, notifier, tokens))
}

func makeWsHandler(hub *game.Hub, notifier *wsadapter.WsNotifier, tokens auth.ITokenService) echo.HandlerFunc {
	return func(c echo.Context) error {
		playerID, err := authenticateWs(c, tokens) // we need to do this seince we are bypassing huma in this case.
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		utils.DeferErrJoin(conn.Close, &err)

		notifier.Register(playerID, conn)
		defer notifier.Deregister(playerID)

		gameCh := hub.Connect(playerID)
		defer hub.Disconnect(playerID)

		g, ok := <-gameCh
		if !ok || g == nil {
			return nil
		}
		err = conn.WriteJSON(g)
		if err != nil {
			return err
		}
		return runMoveLoop(c, conn, hub, g, playerID)
	}
}

func authenticateWs(c echo.Context, tokens auth.ITokenService) (uuid.UUID, error) {
	userIDStr, err := validateBearer(c.Request().Header.Get("Authorization"), tokens)
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(userIDStr)
}

func runMoveLoop(c echo.Context, conn *websocket.Conn, hub *game.Hub, g *game.Game, playerID uuid.UUID) error {
	for {
		var msg moveMsg
		if err := conn.ReadJSON(&msg); err != nil {
			return nil
		}
		if err := hub.MakeMove(c.Request().Context(), g.ID, playerID, msg.Position); err != nil {
			return err
		}
	}
}
