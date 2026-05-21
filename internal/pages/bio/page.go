package bio

import (
	_ "embed"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//go:embed portrait_large.txt
var portraitLarge string

//go:embed portrait_medium.txt
var portraitMedium string

//go:embed portrait_small.txt
var portraitSmall string

// Layout tiers. See <edge-cases> in CONTEXT.md.
//   * width >= largeMinWidth  && height >= largeMinHeight  : 60×36 portrait + text side-by-side
//   * width >= mediumMinWidth && height >= mediumMinHeight : 50×27 portrait + text side-by-side (108×27 mori-reference target)
//   * height >= stackedMinHeight                            : 30×18 portrait above text
//   * else                                                  : text-only
const (
	largeMinWidth   = 120
	largeMinHeight  = 34
	mediumMinWidth  = 100
	mediumMinHeight = 24
	stackedMinHeight = 22
)

// Placeholder copy. Edit these freely; layout reflows.
var (
	name    = "Zack Kaplan"
	tagline = "building things in terminals"
	bioBody = []string{
		"is a developer & tinkerer on the internet,",
		"making small experiments and shipping them",
		"because it's fun.",
		"",
		"this site lives entirely over ssh — no browser,",
		"no html, just a tui served by a go binary.",
	}
	handles = []string{
		"email   you@example.com",
		"github  github.com/yourname",
		"web     yoursite.com",
	}
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
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var content string
	switch {
	case m.width >= largeMinWidth && m.height >= largeMinHeight:
		content = m.viewSideBySide(portraitLarge)
	case m.width >= mediumMinWidth && m.height >= mediumMinHeight:
		content = m.viewSideBySide(portraitMedium)
	case m.height >= stackedMinHeight:
		content = m.viewStacked()
	default:
		content = m.renderText()
	}

	footer := footerStyle.Render("enter to open  ·  q to quit")
	page := lipgloss.JoinVertical(lipgloss.Left, content, "", footer)

	// Anchor to top-left with a small left margin. Looks closer to the mori
	// reference than a centered floating block — the eye expects content to
	// start at the top of the viewport.
	const leftMargin = 4
	indent := strings.Repeat(" ", leftMargin)
	indented := indent + strings.ReplaceAll(page, "\n", "\n"+indent)
	return indented
}

func (m Model) viewSideBySide(art string) string {
	portrait := portraitStyle.Render(strings.TrimRight(art, "\n"))
	text := m.renderText()

	gap := strings.Repeat(" ", 4)
	return lipgloss.JoinHorizontal(lipgloss.Top, portrait, gap, text)
}

func (m Model) viewStacked() string {
	portrait := portraitStyle.Render(strings.TrimRight(portraitSmall, "\n"))
	text := m.renderText()
	return lipgloss.JoinVertical(lipgloss.Left, portrait, "", text)
}

func (m Model) renderText() string {
	parts := []string{
		nameStyle.Render(name),
		taglineStyle.Render(tagline),
		"",
	}
	for _, line := range bioBody {
		parts = append(parts, bodyStyle.Render(line))
	}
	parts = append(parts, "")
	for _, h := range handles {
		parts = append(parts, handleStyle.Render(h))
	}
	return strings.Join(parts, "\n")
}

var (
	portraitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	nameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	taglineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	handleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)
