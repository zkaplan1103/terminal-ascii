package bio

import (
	_ "embed"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zkaplan/terminal-site/internal/page"
)

//go:embed portrait_large.txt
var portraitLarge string

//go:embed portrait_medium.txt
var portraitMedium string

//go:embed portrait_small.txt
var portraitSmall string

// Layout tiers. Three side-by-side tiers (portrait left, text right) + a
// too-small fallback. Every tier produces a fixed-rectangle render that fits
// exactly within the terminal so animation re-paints never scroll.
//
//   - width >= largeMinWidth  && height >= largeMinHeight  : 60-col portrait, full bio, large title
//   - width >= mediumMinWidth && height >= mediumMinHeight : 50-col portrait, full bio, small title
//   - width >= tinyMinWidth   && height >= tinyMinHeight   : 30-col portrait, compact bio, small title
//   - else                                                  : "please resize" message
//
// Tier budgets (width = margin + portrait + gap + textCol):
//
//	Large : 4 + 60 + 4 + 58 = 126 cols, 34 rows
//	Medium: 4 + 50 + 4 + 58 = 116 → relaxed to 110
//	Tiny  : 4 + 30 + 3 + 43 = 80 cols, 24 rows
const (
	largeMinWidth   = 126
	largeMinHeight  = 34
	mediumMinWidth  = 110
	mediumMinHeight = 26
	tinyMinWidth    = 80
	tinyMinHeight   = 24

	// Hard right edge of the braille field, in columns, per tier.
	// = portrait gap + text column width (no extra margin — braille IS the border).
	brailleWidthLarge  = 58 // large + medium tiers
	brailleWidthTiny   = 43 // tiny tier
	braillePad         = 2  // spaces between text and first braille char
)

// bioParagraphs is the full bio used at large + medium tiers.
var bioParagraphs = [][]string{
	{
		"is a technical designer obsessed with the idea that",
		"the design space has unlimited potential — and the",
		"best projects live at the edges of it. weird, cool,",
		"and built for the experience first.",
	},
	{
		"this site is an ascii art showroom. I saw people",
		"making terminal bio pages and wanted to take it",
		"further: an interactive hub where every section is",
		"its own ascii project with its own reason to exist.",
		"the terminal is the perfect canvas — its constraints",
		"are the whole point.",
	},
	{
		"each project here comes with a description of why I",
		"built it and what I liked about it. if you have an",
		"idea for a section I should make, reach out.",
	},
}

// bioParagraphsTiny is the compact bio for the 80×24 tier.
var bioParagraphsTiny = [][]string{
	{
		"technical designer. weird projects,",
		"cool experiences, terminals first.",
	},
	{
		"ascii art showroom over ssh. every",
		"section is its own project — reach",
		"out if you have an idea for one.",
	},
}

type navItem struct {
	label       string
	target      string
	description string
}

var navItems = []navItem{
	{"projects", "projects", "ascii projects & experiments"},
	{"about", "adopt", "the petting zoo — foster an animal"},
	{"contact", "contact", "say hi"},
}

type Model struct {
	width  int
	height int
	cursor int
	title  titleAnim
}

func New() Model {
	return Model{title: newTitleAnim()}
}

func (m Model) Init() tea.Cmd { return m.title.tick() }

// computeBrailRows calculates the total braille row count for the current
// tier without rendering. This is deterministic from dimensions alone so we
// can keep titleAnim.brailRows accurate inside Update(), not just View().
func (m Model) computeBrailRows() int {
	var paragraphs [][]string
	var titleLineCount int
	var portraitH int

	switch {
	case m.width >= largeMinWidth && m.height >= largeMinHeight:
		paragraphs = bioParagraphs
		titleLineCount = len(m.title.lines)
		portraitH = 32 // large portrait rows
	case m.width >= mediumMinWidth && m.height >= mediumMinHeight:
		paragraphs = bioParagraphs
		titleLineCount = len(m.title.lines)
		portraitH = 22 // medium portrait rows
	default:
		paragraphs = bioParagraphsTiny
		titleLineCount = len(m.title.linesSm)
		portraitH = 14 // small portrait rows
	}

	// Flatten bio lines (same logic as renderBrailleTextColumn).
	bioCount := 0
	for pi, para := range paragraphs {
		bioCount += len(para)
		if pi < len(paragraphs)-1 {
			bioCount++ // blank between paragraphs
		}
	}

	// naturalH = titleLines + 1(div) + bioLines + 1(div) + 1(nav)
	naturalH := titleLineCount + 1 + bioCount + 1 + 1
	extraPad := 0
	if portraitH > naturalH {
		extraPad = portraitH - naturalH
	}

	// totalRows = naturalH + extraPad
	return naturalH + extraPad
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Sync brailRows immediately so the animation has the correct total
		// from the first tick after a resize.
		m.title.brailRows = m.computeBrailRows()
	case titleTickMsg:
		// Re-sync brailRows before advancing so the drop phase always knows
		// the correct total for the current tier.
		m.title.brailRows = m.computeBrailRows()
		m.title = m.title.advance()
		return m, m.title.tick()
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right", "l":
			if m.cursor < len(navItems)-1 {
				m.cursor++
			}
		case "enter":
			target := navItems[m.cursor].target
			return m, func() tea.Msg { return page.NavigateMsg{To: target} }
		}
	}
	return m, nil
}

// View renders the bio page as a fixed rectangle that always fits within the
// terminal. Below the tiny tier (80×24) we show a resize prompt instead.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	const leftMargin = 4
	indent := strings.Repeat(" ", leftMargin)

	if m.width < tinyMinWidth || m.height < tinyMinHeight {
		msg := footerStyle.Render("please resize your terminal to at least 80×24")
		topPad := m.height / 2
		if topPad < 1 {
			topPad = 1
		}
		var rows []string
		for i := 0; i < topPad; i++ {
			rows = append(rows, "")
		}
		rows = append(rows, indent+msg)
		for len(rows) < m.height {
			rows = append(rows, "")
		}
		return strings.Join(rows[:m.height], "\n")
	}

	// Pick tier.
	var (
		portraitArt  string
		paragraphs   [][]string
		brailleWidth int
		gapWidth     int
		tier         titleTier
	)
	switch {
	case m.width >= largeMinWidth && m.height >= largeMinHeight:
		portraitArt = portraitLarge
		paragraphs = bioParagraphs
		brailleWidth = brailleWidthLarge
		gapWidth = 4
		tier = titleLarge
	case m.width >= mediumMinWidth && m.height >= mediumMinHeight:
		portraitArt = portraitMedium
		paragraphs = bioParagraphs
		brailleWidth = brailleWidthLarge
		gapWidth = 4
		tier = titleLarge
	default:
		portraitArt = portraitSmall
		paragraphs = bioParagraphsTiny
		brailleWidth = brailleWidthTiny
		gapWidth = 3
		tier = titleTiny
	}

	body := m.renderSideBySide(portraitArt, paragraphs, brailleWidth, gapWidth, tier)

	body = clampToHeight(body, m.height)

	return indent + strings.ReplaceAll(body, "\n", "\n"+indent)
}

// renderSideBySide builds the full page: portrait left, braille text column right.
func (m Model) renderSideBySide(art string, paragraphs [][]string, brailleWidth, gapWidth int, tier titleTier) string {
	portrait := portraitStyle.Render(strings.TrimRight(art, "\n"))
	portraitH := strings.Count(strings.TrimRight(art, "\n"), "\n") + 1

	titleStr := m.title.render(tier)
	navStr := m.renderNavStr()
	glowRow := m.title.titleGlowRow()
	flicker := m.title.flickerColor()

	textCol, _ := renderBrailleTextColumn(
		titleStr,
		paragraphs,
		navStr,
		brailleWidth,
		braillePad,
		glowRow,
		flicker,
		portraitH,
	)

	gap := strings.Repeat(" ", gapWidth)
	row := lipgloss.JoinHorizontal(lipgloss.Top, portrait, gap, textCol)

	footer := footerStyle.Render("←→ to move  ·  enter to open  ·  q to quit")
	return lipgloss.JoinVertical(lipgloss.Left, row, "", footer)
}

// renderNavStr returns the nav menu as a plain string (no padding/wrapper).
func (m Model) renderNavStr() string {
	var parts []string
	for i, it := range navItems {
		if i == m.cursor {
			parts = append(parts, focusedLabelStyle.Render("["+it.label+"]"))
		} else {
			parts = append(parts, labelStyle.Render(" "+it.label+" "))
		}
	}
	return strings.Join(parts, "  ")
}

// clampToHeight ensures body has exactly maxRows lines.
func clampToHeight(body string, maxRows int) string {
	lines := strings.Split(body, "\n")
	if len(lines) > maxRows {
		lines = lines[:maxRows]
	}
	for len(lines) < maxRows {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

var (
	portraitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	labelStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	descStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	focusedLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	focusedDescStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)
