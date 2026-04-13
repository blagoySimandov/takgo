package client

import (
	"net/http"
	"strings"

	wsadapter "github.com/blagoySimandov/takgo/internal/adapters/ws"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

func connectWS(baseURL, token string) tea.Cmd {
	return func() tea.Msg {
		wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/api/v1/game/connect"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{
			"Authorization": {"Bearer " + token},
		})
		if err != nil {
			return wsErrMsg{err}
		}
		return connectedMsg{conn}
	}
}

func listenForMove(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		var msg wsadapter.GameStateMsg
		if err := conn.ReadJSON(&msg); err != nil {
			return wsErrMsg{err}
		}
		return gameMsgReceived{msg}
	}
}

func waitingView() string {
	return "\n  Waiting for an opponent...\n\n  ctrl+c to quit\n"
}
