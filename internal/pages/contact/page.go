package contact

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
	title := titleStyle.Render("contact")
	handles := handleStyle.Render(strings.Join([]string{
		"email   you@example.com",
		"github  github.com/yourname",
		"web     yoursite.com",
	}, "\n"))
	hint := hintStyle.Render("esc to home  ·  q to quit")
	page := lipgloss.JoinVertical(lipgloss.Left, title, "", handles, "", hint)
	indent := strings.Repeat(" ", 4)
	return indent + strings.ReplaceAll(page, "\n", "\n"+indent)
}

var (
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	handleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
)
