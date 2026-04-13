package main

import (
	"log"
	"time"

	handlers "github.com/blagoySimandov/takgo/internal/adapters/http"
	jwtadapter "github.com/blagoySimandov/takgo/internal/adapters/jwt"
	"github.com/blagoySimandov/takgo/internal/adapters/sqlite"
	wsadapter "github.com/blagoySimandov/takgo/internal/adapters/ws"
	"github.com/blagoySimandov/takgo/internal/domain/auth"
	"github.com/blagoySimandov/takgo/internal/domain/game"
	"github.com/blagoySimandov/takgo/internal/domain/matchmaking"
	"github.com/labstack/echo/v4"
)

func main() {
	database, err := sqlite.Open("takgo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	tokenSvc := jwtadapter.NewService(jwtadapter.GetSecret(), 24*time.Hour)
	authSvc := auth.NewAuthService(sqlite.NewUserRepo(database), tokenSvc)
	notifier := wsadapter.NewWsNotifier()
	hub := game.NewHub(matchmaking.NewQueue(), game.NewInMemoryGameRepo(), notifier)

	e := echo.New()
	handlers.RegisterRoutes(e, authSvc, tokenSvc, hub, notifier)

	log.Fatal(e.Start(":8080"))
}
