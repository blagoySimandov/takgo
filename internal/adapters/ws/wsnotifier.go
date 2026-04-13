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

type gameStateMsg struct {
	Board       game.Board     `json:"board"`
	State       game.GameState `json:"state"`
	CurrentTurn uuid.UUID      `json:"currentTurn"`
	WinnerID    *uuid.UUID     `json:"winnerId,omitempty"`
	LastMove    game.Move      `json:"lastMove"`
}

func (n *WsNotifier) NotifyMove(ctx context.Context, g *game.Game, move game.Move) error {
	msg := gameStateMsg{
		Board:       g.Board,
		State:       g.State,
		CurrentTurn: g.CurrentTurn,
		WinnerID:    g.WinnerID,
		LastMove:    move,
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
