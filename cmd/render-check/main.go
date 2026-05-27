package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zkaplan/terminal-site/internal/router"
)

// Renders each entry-page at one representative size so you can eyeball the
// full Phase 2 routing surface without launching the SSH server.
func main() {
	cases := []struct {
		name       string
		username   string
		w, h       int
	}{
		{"bio entry (ssh bio@) at 140x42 — large tier", "bio", 140, 42},
		{"bio entry at 120x32 — medium tier", "bio", 120, 32},
		{"bio entry at 90x40 — stacked tier", "bio", 90, 40},
		{"bio entry (ssh someone@) — unknown user lands on bio", "someone", 130, 42},
		{"adopt entry (ssh adopt@)", "adopt", 100, 30},
		{"projects entry (ssh projects@)", "projects", 100, 30},
		{"contact entry (ssh contact@)", "contact", 100, 30},
		{"bio at default 80x24 (text-only fallback)", "bio", 80, 24},
	}
	for _, c := range cases {
		m := tea.Model(router.New(c.username, c.w, c.h))
		fmt.Fprintf(os.Stderr, "\n=== %s ===\n", c.name)
		fmt.Print(m.View())
		fmt.Println()
	}
}
