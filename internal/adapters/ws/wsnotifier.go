package ws

import (
	"context"
	"sync"

	"github.com/blagoySimandov/takgo/internal/domain/game"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WsNotifier struct {
	mu    sync.RWMutex
	conns map[uuid.UUID]*websocket.Conn // not a sync map for simplicity and better type safety
}

func NewWsNotifier() *WsNotifier {
	return &WsNotifier{conns: make(map[uuid.UUID]*websocket.Conn)}
}

func (n *WsNotifier) Register(playerID uuid.UUID, conn *websocket.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.conns[playerID] = conn
}

func (n *WsNotifier) Deregister(playerID uuid.UUID) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.conns, playerID)
}

type MoveMsg struct {
	Position int `json:"position" minimum:"0" maximum:"8" description:"Board position (0-8, row-major)"`
}

type GameStateMsg struct {
	Board       game.Board     `json:"board" description:"Board cells: 0=empty 1=X 2=O"`
	State       game.GameState `json:"state" description:"Game state: 0=waiting 1=playing 2=finished"`
	CurrentTurn uuid.UUID      `json:"currentTurn" description:"UUID of the player whose turn it is"`
	WinnerID    *uuid.UUID     `json:"winnerId,omitempty" description:"UUID of the winner, null on draw or unfinished"`
	LastMove    MoveMsg        `json:"lastMove"`
}

func (n *WsNotifier) NotifyMove(ctx context.Context, g *game.Game, move game.Move) error {
	msg := GameStateMsg{
		Board:       g.Board,
		State:       g.State,
		CurrentTurn: g.CurrentTurn,
		WinnerID:    g.WinnerID,
		LastMove:    MoveMsg{Position: move.Position},
	}
	for _, p := range g.Players {
		n.mu.RLock()
		conn, ok := n.conns[p.ID]
		n.mu.RUnlock()
		if ok {
			if err := conn.WriteJSON(msg); err != nil {
				return err
			}
			if g.State == game.Finished {
				n.closeWithGameOver(conn)
			}
		}
	}
	return nil
}

func (n *WsNotifier) closeWithGameOver(conn *websocket.Conn) {
	if err := conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "game over")); err != nil {
		return
	}
}
