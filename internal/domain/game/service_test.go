package game_test

import (
	"errors"
	"testing"
	"time"

	"github.com/blagoySimandov/takgo/internal/domain/game"
	"github.com/google/uuid"
)

func newTestGame() (*game.Game, uuid.UUID, uuid.UUID) {
	p1, p2 := uuid.New(), uuid.New()
	g := &game.Game{
		ID:    uuid.New(),
		State: game.Playing,
		Players: [2]game.Player{
			{ID: p1, Symbol: game.X},
			{ID: p2, Symbol: game.O},
		},
		CurrentTurn: p1,
		CreatedAt:   time.Now(),
	}
	return g, p1, p2
}

func TestPlayerCanMakeAValidMove(t *testing.T) {
	g, p1, _ := newTestGame()

	if err := g.ApplyMove(p1, 4); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Board[4] != game.X {
		t.Fatal("expected X at position 4")
	}
}

func TestTurnSwitchesAfterMove(t *testing.T) {
	g, p1, p2 := newTestGame()

	_ = g.ApplyMove(p1, 0)

	if g.CurrentTurn != p2 {
		t.Fatal("expected turn to switch to p2")
	}
}

func TestPlayerCannotMoveOutOfTurn(t *testing.T) {
	g, _, p2 := newTestGame()

	err := g.ApplyMove(p2, 0)

	if !errors.Is(err, game.ErrNotYourTurn) {
		t.Fatalf("expected ErrNotYourTurn, got %v", err)
	}
}

func TestPlayerCannotMoveToOccupiedCell(t *testing.T) {
	g, p1, p2 := newTestGame()
	_ = g.ApplyMove(p1, 0)

	err := g.ApplyMove(p2, 0)

	if !errors.Is(err, game.ErrCellOccupied) {
		t.Fatalf("expected ErrCellOccupied, got %v", err)
	}
}

func TestPlayerCannotMoveWhenGameIsNotPlaying(t *testing.T) {
	g, p1, _ := newTestGame()
	g.State = game.Finished

	err := g.ApplyMove(p1, 0)

	if !errors.Is(err, game.ErrGameNotPlaying) {
		t.Fatalf("expected ErrGameNotPlaying, got %v", err)
	}
}

func TestPlayerCannotMoveToInvalidPosition(t *testing.T) {
	g, p1, _ := newTestGame()

	err := g.ApplyMove(p1, 9)

	if !errors.Is(err, game.ErrInvalidPos) {
		t.Fatalf("expected ErrInvalidPos, got %v", err)
	}
}

func TestGameDetectsWinByRow(t *testing.T) {
	g, p1, p2 := newTestGame()
	// X _ _
	// O O _
	// X _ _  → X wins top row
	moves := []struct {
		player uuid.UUID
		pos    int
	}{
		{p1, 0},
		{p2, 3},
		{p1, 1},
		{p2, 4},
		{p1, 2},
	}
	for _, m := range moves {
		_ = g.ApplyMove(m.player, m.pos)
	}

	if g.State != game.Finished {
		t.Fatal("expected game to be finished")
	}
	if g.WinnerID == nil || *g.WinnerID != p1 {
		t.Fatal("expected p1 to be the winner")
	}
}

func TestGameDetectsWinByColumn(t *testing.T) {
	g, p1, p2 := newTestGame()
	// X O _
	// X O _
	// X _ _  → X wins left column
	moves := []struct {
		player uuid.UUID
		pos    int
	}{
		{p1, 0},
		{p2, 1},
		{p1, 3},
		{p2, 4},
		{p1, 6},
	}
	for _, m := range moves {
		_ = g.ApplyMove(m.player, m.pos)
	}

	if g.State != game.Finished {
		t.Fatal("expected game to be finished")
	}
	if g.WinnerID == nil || *g.WinnerID != p1 {
		t.Fatal("expected p1 to be the winner")
	}
}

func TestGameDetectsWinByDiagonal(t *testing.T) {
	g, p1, p2 := newTestGame()
	// X O _
	// _ X O
	// _ _ X  → X wins main diagonal
	moves := []struct {
		player uuid.UUID
		pos    int
	}{
		{p1, 0},
		{p2, 1},
		{p1, 4},
		{p2, 5},
		{p1, 8},
	}
	for _, m := range moves {
		_ = g.ApplyMove(m.player, m.pos)
	}

	if g.State != game.Finished {
		t.Fatal("expected game to be finished")
	}
	if g.WinnerID == nil || *g.WinnerID != p1 {
		t.Fatal("expected p1 to be the winner")
	}
}

func TestGameDetectsDrawWhenBoardIsFull(t *testing.T) {
	g, p1, p2 := newTestGame()
	// X O X
	// X O O
	// O X X  → draw
	moves := []struct {
		player uuid.UUID
		pos    int
	}{
		{p1, 0},
		{p2, 1},
		{p1, 2},
		{p2, 4},
		{p1, 3},
		{p2, 5},
		{p1, 7},
		{p2, 6},
		{p1, 8},
	}
	for _, m := range moves {
		_ = g.ApplyMove(m.player, m.pos)
	}

	if g.State != game.Finished {
		t.Fatal("expected game to be finished")
	}
	if g.WinnerID != nil {
		t.Fatal("expected no winner on draw")
	}
}

func TestGameDoesNotContinueAfterWin(t *testing.T) {
	g, p1, p2 := newTestGame()
	// p1 wins top row
	moves := []struct {
		player uuid.UUID
		pos    int
	}{
		{p1, 0},
		{p2, 3},
		{p1, 1},
		{p2, 4},
		{p1, 2},
	}
	for _, m := range moves {
		_ = g.ApplyMove(m.player, m.pos)
	}

	err := g.ApplyMove(p2, 5)

	if !errors.Is(err, game.ErrGameNotPlaying) {
		t.Fatalf("expected ErrGameNotPlaying after win, got %v", err)
	}
}
