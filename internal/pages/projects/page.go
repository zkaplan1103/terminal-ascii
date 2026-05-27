package projects

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zkaplan/terminal-site/internal/page"
)

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
	title := titleStyle.Render("projects")
	body := bodyStyle.Render("nothing here yet — check back later.\n\nesc to home  ·  q to quit")
	page := lipgloss.JoinVertical(lipgloss.Left, title, "", body)
	indent := strings.Repeat(" ", 4)
	return indent + strings.ReplaceAll(page, "\n", "\n"+indent)
}

var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	bodyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)
