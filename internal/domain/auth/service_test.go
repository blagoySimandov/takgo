package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	jwtadapter "github.com/blagoySimandov/takgo/internal/adapters/jwt"
	"github.com/blagoySimandov/takgo/internal/adapters/sqlite"
	"github.com/blagoySimandov/takgo/internal/domain/auth"
	"github.com/blagoySimandov/takgo/internal/migrations"
	"github.com/uptrace/bun/migrate"
)

func newTestService(t *testing.T) *auth.AuthService {
	t.Helper()

	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	})

	ctx := context.Background()
	migrator := migrate.NewMigrator(db, migrations.Migrations)
	if err := migrator.Init(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := migrator.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	repo := sqlite.NewUserRepo(db)
	tokenSvc := jwtadapter.NewService("test-secret", time.Hour)
	return auth.NewAuthService(repo, tokenSvc)
}

func TestUserCanRegisterWithValidCredentials(t *testing.T) {
	svc := newTestService(t)

	token, err := svc.RegisterUser(context.Background(), "alice", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected token, got empty string")
	}
}

func TestUserCannotRegisterWithDuplicateUsername(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, _ = svc.RegisterUser(ctx, "alice", "password123")
	_, err := svc.RegisterUser(ctx, "alice", "otherpassword")

	if !errors.Is(err, auth.ErrUserExists) {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}
}

func TestUserCanLoginAfterRegistering(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, _ = svc.RegisterUser(ctx, "alice", "password123")
	token, err := svc.Login(ctx, "alice", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected token, got empty string")
	}
}

func TestUserCannotLoginWithWrongPassword(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, _ = svc.RegisterUser(ctx, "alice", "password123")
	_, err := svc.Login(ctx, "alice", "wrongpassword")

	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserCannotLoginIfNotRegistered(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.Login(context.Background(), "ghost", "password123")

	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
