package adopt

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zkaplan/terminal-site/internal/page"
)

// Phase 2 stub. Phase 3 will replace this with the full petting-zoo:
// category picker → animal picker → care view. See <zoo> in CONTEXT.md.

type Model struct {
	width  int
	height int
}

func New() Model { return Model{} }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return m, func() tea.Msg { return page.NavigateMsg{To: "bio"} }
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	title := titleStyle.Render("adopt")
	body := bodyStyle.Render(strings.Join([]string{
		"the petting zoo opens soon.",
		"",
		"you'll be able to pick from farm + exotic animals,",
		"foster one for the duration of your session, and",
		"care for it with feed / walk / sleep / pet actions.",
		"",
		"no accounts, no saves — when you disconnect, the",
		"animal goes back. like a real foster.",
	}, "\n"))
	hint := hintStyle.Render("esc to home  ·  q to quit")
	page := lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", hint)
	indent := strings.Repeat(" ", 4)
	return indent + strings.ReplaceAll(page, "\n", "\n"+indent)
}

var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	bodyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
)
