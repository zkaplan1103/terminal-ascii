package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zkaplan/terminal-site/internal/pages/bio"
)

func main() {
	sizes := []struct {
		name          string
		width, height int
	}{
		{"130x42 (large tier)", 130, 42},
		{"108x27 (mori reference, medium tier)", 108, 27},
		{"90x30 (stacked tier)", 90, 30},
		{"80x24 (default mac terminal, text-only)", 80, 24},
	}
	for _, s := range sizes {
		m := tea.Model(bio.New())
		m, _ = m.Update(tea.WindowSizeMsg{Width: s.width, Height: s.height})
		fmt.Fprintf(os.Stderr, "\n=== %s ===\n", s.name)
		fmt.Print(m.View())
		fmt.Println()
	}
}
