package game

import (
	"context"

	"github.com/google/uuid"
)

var winLines = [8][3]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // cols
	{0, 4, 8}, {2, 4, 6},             // diagonals
}

type GameService struct {
	repo     GameRepository
	notifier Notifier
}

func NewGameService(repo GameRepository, notifier Notifier) *GameService {
	return &GameService{repo: repo, notifier: notifier}
}

func (s *GameService) MakeMove(ctx context.Context, gameID, playerID uuid.UUID, position int) error {
	g, err := s.repo.FindByID(ctx, gameID)
	if err != nil {
		return err
	}
	if err := g.ApplyMove(playerID, position); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, g); err != nil {
		return err
	}
	return s.notifier.NotifyMove(ctx, g, Move{PlayerID: playerID, Position: position})
}

func (g *Game) ApplyMove(playerID uuid.UUID, position int) error {
	if err := g.validateMove(playerID, position); err != nil {
		return err
	}
	symbol, _ := g.symbolFor(playerID)
	g.Board[position] = symbol
	g.updateOutcome(symbol)
	if g.State != Finished {
		g.nextTurn()
	}
	return nil
}

func (g *Game) validateMove(playerID uuid.UUID, position int) error {
	if g.State != Playing {
		return ErrGameNotPlaying
	}
	if g.CurrentTurn != playerID {
		return ErrNotYourTurn
	}
	if position < 0 || position > 8 {
		return ErrInvalidPos
	}
	if g.Board[position] != Empty {
		return ErrCellOccupied
	}
	return nil
}

func (g *Game) updateOutcome(symbol Cell) {
	if hasWon(g.Board, symbol) {
		g.State = Finished
		winnerID := g.CurrentTurn
		g.WinnerID = &winnerID
	} else if isDraw(g.Board) {
		g.State = Finished
	}
}

func (g *Game) nextTurn() {
	for _, p := range g.Players {
		if p.ID != g.CurrentTurn {
			g.CurrentTurn = p.ID
			return
		}
	}
}

func (g *Game) symbolFor(playerID uuid.UUID) (Cell, bool) {
	for _, p := range g.Players {
		if p.ID == playerID {
			return p.Symbol, true
		}
	}
	return Empty, false
}

func hasWon(board Board, symbol Cell) bool {
	for _, line := range winLines {
		if board[line[0]] == symbol && board[line[1]] == symbol && board[line[2]] == symbol {
			return true
		}
	}
	return false
}

func isDraw(board Board) bool {
	for _, cell := range board {
		if cell == Empty {
			return false
		}
	}
	return true
}
