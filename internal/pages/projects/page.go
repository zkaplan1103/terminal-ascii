package projects

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zkaplan/terminal-site/internal/page"
)

type project struct {
	name   string
	tag    string
	desc   string
	why    string
	target string // NavigateMsg target on enter
}

var projects = []project{
	{
		name:   "ascii petting zoo",
		tag:    "interactive",
		desc:   "foster a farm animal for your ssh session. feed it, walk it,\nput it to sleep. when you disconnect it goes back — no saves,\nno accounts. the terminal as a tiny living thing.",
		why:    "I wanted to see how far you could push character-cell animation.\nbraille dots at 48×24 turned out to be the answer.",
		target: "adopt",
	},
}

type Model struct {
	width  int
	height int
	cursor int
}

func New() Model { return Model{} }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(projects)-1 {
				m.cursor++
			}
		case "enter", " ":
			target := projects[m.cursor].target
			return m, func() tea.Msg { return page.NavigateMsg{To: target} }
		case "esc":
			return m, func() tea.Msg { return page.NavigateMsg{To: "bio"} }
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString(headerStyle.Render("projects"))
	b.WriteString("\n")
	b.WriteString(subStyle.Render("ascii experiments & interactive things"))
	b.WriteString("\n\n")

	for i, p := range projects {
		selected := i == m.cursor

		// Name + tag line
		nameStr := fmt.Sprintf("%-28s", p.name)
		tagStr := p.tag
		if selected {
			b.WriteString(selectedNameStyle.Render("▶ "+nameStr) + tagStyle.Render(tagStr))
		} else {
			b.WriteString(itemNameStyle.Render("  "+nameStr) + tagStyle.Render(tagStr))
		}
		b.WriteString("\n")

		// Description — always visible
		for _, line := range strings.Split(p.desc, "\n") {
			b.WriteString(descStyle.Render("  " + line))
			b.WriteString("\n")
		}

		// Why I built it — only shown when selected
		if selected {
			b.WriteString("\n")
			b.WriteString(whyLabelStyle.Render("  why: "))
			lines := strings.Split(p.why, "\n")
			b.WriteString(whyStyle.Render(lines[0]))
			b.WriteString("\n")
			for _, line := range lines[1:] {
				b.WriteString(whyStyle.Render("        " + line))
				b.WriteString("\n")
			}
		}

		b.WriteString("\n")
	}

	b.WriteString(hintStyle.Render("↑↓ to move  ·  enter to visit  ·  esc to home  ·  q to quit"))

	content := b.String()
	indent := strings.Repeat(" ", 4)
	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Center,
		indent+strings.ReplaceAll(strings.TrimRight(content, "\n"), "\n", "\n"+indent))
}

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	subStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	selectedNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true)

	itemNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63")).
			Italic(true)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	whyLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	whyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)
