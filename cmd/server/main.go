package main

import (
	"log"
	"time"

	handlers "github.com/blagoySimandov/takgo/internal/adapters/http"
	jwtadapter "github.com/blagoySimandov/takgo/internal/adapters/jwt"
	"github.com/blagoySimandov/takgo/internal/adapters/sqlite"
	"github.com/blagoySimandov/takgo/internal/domain/auth"
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

	userRepo := sqlite.NewUserRepo(database)

	tokenSvc := jwtadapter.NewService("secret-change-me", 24*time.Hour)
	authSvc := auth.NewAuthService(userRepo, tokenSvc)

	e := echo.New()
	handlers.RegisterRoutes(e, authSvc, tokenSvc)

	log.Fatal(e.Start(":8080"))
}
