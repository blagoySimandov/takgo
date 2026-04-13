package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type authMode int

const (
	modeLogin authMode = iota
	modeRegister
)

type authModel struct {
	mode     authMode
	username textinput.Model
	password textinput.Model
	focused  int
	loading  bool
	err      error
}

func newAuthModel() authModel {
	u := textinput.New()
	u.Placeholder = "username"
	u.Focus()

	p := textinput.New()
	p.Placeholder = "password"
	p.EchoMode = textinput.EchoPassword
	p.EchoCharacter = '•'

	return authModel{username: u, password: p}
}

func (m authModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m authModel) update(msg tea.Msg) (authModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m.updateInputs(msg)
}

func (m authModel) handleKey(msg tea.KeyMsg) (authModel, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.mode = 1 - m.mode
	case "up", "shift+tab":
		m.focused = 0
		m.username.Focus()
		m.password.Blur()
	case "down", "enter":
		if m.focused == 0 {
			m.focused = 1
			m.username.Blur()
			m.password.Focus()
			return m, nil
		}
		m.loading = true
		m.err = nil
		return m, submitAuth(m)
	}
	return m.updateInputs(msg)
}

func (m authModel) updateInputs(msg tea.Msg) (authModel, tea.Cmd) {
	var cmds [2]tea.Cmd
	m.username, cmds[0] = m.username.Update(msg)
	m.password, cmds[1] = m.password.Update(msg)
	return m, tea.Batch(cmds[0], cmds[1])
}

func (m authModel) view() string {
	var b strings.Builder
	b.WriteString("\n  TakGo\n\n")

	loginStyle := "  [ Login ]"
	registerStyle := "  [ Register ]"
	if m.mode == modeLogin {
		loginStyle = "  [ Login* ]"
	} else {
		registerStyle = "  [ Register* ]"
	}
	b.WriteString(loginStyle + "  " + registerStyle + "  (tab to switch)\n\n")
	b.WriteString("  " + m.username.View() + "\n")
	b.WriteString("  " + m.password.View() + "\n\n")

	if m.loading {
		b.WriteString("  ...")
	} else if m.err != nil {
		b.WriteString("  error: " + m.err.Error())
	} else {
		b.WriteString("  enter to confirm")
	}
	b.WriteString("\n")
	return b.String()
}

func submitAuth(m authModel) tea.Cmd {
	return func() tea.Msg {
		path := "/api/v1/login"
		if m.mode == modeRegister {
			path = "/api/v1/register"
		}
		token, err := authRequest("http://localhost:8080"+path, m.username.Value(), m.password.Value())
		if err != nil {
			return authErrMsg{err}
		}
		_ = saveToken(token)
		return tokenMsg{token}
	}
}

type authErrMsg struct{ err error }

func authRequest(url, username, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned %d", resp.StatusCode)
	}
	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Token, nil
}
