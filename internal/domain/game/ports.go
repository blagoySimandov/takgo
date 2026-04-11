package game

import (
	"context"

	"github.com/google/uuid"
)

type GameRepository interface {
	Create(ctx context.Context, game *Game) error
	FindByID(ctx context.Context, id uuid.UUID) (*Game, error)
	Save(ctx context.Context, game *Game) error
}

type Notifier interface {
	NotifyMove(ctx context.Context, game *Game, move Move) error
}
