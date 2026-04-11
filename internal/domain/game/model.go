package game

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotYourTurn    = errors.New("not your turn")
	ErrCellOccupied   = errors.New("cell already occupied")
	ErrGameNotPlaying = errors.New("game is not in playing state")
	ErrInvalidPos     = errors.New("position must be 0-8")
)

type Cell int8

const (
	Empty Cell = iota
	X
	O
)

type GameState int8

const (
	Waiting  GameState = iota // waiting for second player to join
	Playing                   // both players present, game in progress
	Finished                  // game over (win or draw)
)

type Player struct {
	ID     uuid.UUID
	Symbol Cell
}

type Move struct {
	PlayerID uuid.UUID
	Position int // 0-8, row-major order
}

type Board [9]Cell

type Game struct {
	ID          uuid.UUID
	Players     [2]Player
	Board       Board
	State       GameState
	CurrentTurn uuid.UUID
	WinnerID    *uuid.UUID // nil on draw or unfinished
	CreatedAt   time.Time
}
