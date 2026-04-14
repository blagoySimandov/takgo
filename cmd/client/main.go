package main

import (
	"fmt"
	"os"

	"github.com/blagoySimandov/takgo/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(client.NewAppModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
