package client

import (
	wsadapter "github.com/blagoySimandov/takgo/internal/adapters/ws"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

type appScreen int

const (
	screenAuth appScreen = iota
	screenWaiting
	screenPlaying
	screenFinished
)

type appModel struct {
	screen appScreen
	auth   authModel
	conn   *websocket.Conn
	game   *wsadapter.GameStateMsg
	myID   string
	err    error
}

type tokenMsg struct{ token string }
type connectedMsg struct{ conn *websocket.Conn }
type gameMsgReceived struct{ state wsadapter.GameStateMsg }
type wsErrMsg struct{ err error }

func NewAppModel() appModel {
	return appModel{
		screen: screenAuth,
		auth:   newAuthModel(),
	}
}

func (m appModel) Init() tea.Cmd {
	return tea.Batch(m.auth.Init(), tryLoadToken())
}

func tryLoadToken() tea.Cmd {
	return func() tea.Msg {
		token, err := loadToken()
		if err != nil {
			return nil
		}
		return tokenMsg{token}
	}
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "ctrl+c" {
		return m, tea.Quit
	}
	switch m.screen {
	case screenAuth:
		return m.updateAuth(msg)
	case screenWaiting, screenPlaying:
		return m.updateGame(msg)
	case screenFinished:
		return m, tea.Quit
	}
	return m, nil
}

func (m appModel) updateAuth(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tokenMsg:
		myID, err := subjectFromJWT(msg.token)
		if err != nil {
			m.auth.err = err
			return m, nil
		}
		m.myID = myID
		m.screen = screenWaiting
		return m, connectWS("http://localhost:8080", msg.token)
	case authErrMsg:
		m.auth.loading = false
		m.auth.err = msg.err
		return m, nil
	default:
		var cmd tea.Cmd
		m.auth, cmd = m.auth.update(msg)
		return m, cmd
	}
}

func (m appModel) updateGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectedMsg:
		m.conn = msg.conn
		return m, listenForMove(m.conn)
	case gameMsgReceived:
		m.game = &msg.state
		m.screen = screenPlaying
		if msg.state.State == 2 { // Finished
			m.screen = screenFinished
			return m, nil
		}
		return m, listenForMove(m.conn)
	case wsErrMsg:
		m.err = msg.err
		return m, tea.Quit
	case tea.KeyMsg:
		return m.handleMoveKey(msg)
	}
	return m, nil
}

func (m appModel) handleMoveKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.game == nil || m.game.CurrentTurn.String() != m.myID {
		return m, nil
	}
	key := msg.String()
	if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
		pos := int(key[0] - '1')
		if err := m.conn.WriteJSON(wsadapter.MoveMsg{Position: pos}); err != nil {
			m.err = err
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m appModel) View() string {
	switch m.screen {
	case screenAuth:
		return m.auth.view()
	case screenWaiting:
		return waitingView()
	case screenPlaying, screenFinished:
		return renderBoard(*m.game, m.myID)
	}
	return ""
}
