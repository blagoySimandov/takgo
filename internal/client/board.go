package client

import (
	"strings"

	wsadapter "github.com/blagoySimandov/takgo/internal/adapters/ws"
	"github.com/blagoySimandov/takgo/internal/domain/game"
)

func renderBoard(state wsadapter.GameStateMsg, myID string) string {
	var b strings.Builder
	b.WriteString("\n")
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			pos := row*3 + col
			b.WriteString(cellStr(state.Board[pos]))
			if col < 2 {
				b.WriteString(" │ ")
			}
		}
		b.WriteString("\n")
		if row < 2 {
			b.WriteString("──┼───┼──\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(statusLine(state, myID))
	if state.State == game.Finished {
		b.WriteString("\n\n  press any key to exit")
	}
	b.WriteString("\n")
	return b.String()
}

func cellStr(c game.Cell) string {
	switch c {
	case game.X:
		return "X"
	case game.O:
		return "O"
	default:
		return "·"
	}
}

func statusLine(state wsadapter.GameStateMsg, myID string) string {
	switch state.State {
	case game.Finished:
		if state.WinnerID == nil {
			return "Draw!"
		}
		if state.WinnerID.String() == myID {
			return "You win!"
		}
		return "You lose."
	default:
		if state.CurrentTurn.String() == myID {
			return "Your turn — press 1-9:"
		}
		return "Opponent's turn..."
	}
}
