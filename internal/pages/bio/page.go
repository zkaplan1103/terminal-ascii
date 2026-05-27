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
//   * width >= largeMinWidth  && height >= largeMinHeight  : 60-col portrait, full bio, large title
//   * width >= mediumMinWidth && height >= mediumMinHeight : 50-col portrait, full bio, small title
//   * width >= tinyMinWidth   && height >= tinyMinHeight   : 30-col portrait, compact bio, small title
//   * else                                                  : "please resize" message
//
// Tier budgets (width = margin + portrait + gap + frame + safety):
//   Large : 4 + 60 + 4 + 58 + 0 = 126 cols, 34 rows
//   Medium: 4 + 50 + 4 + 58 + 0 = 116 → relaxed to 110 (right-edge clip ok at narrow widths)
//   Tiny  : 4 + 30 + 3 + 42 + 1 = 80 cols, 24 rows (default Mac terminal)
//
// Tier heights:
//   Large : portrait 32r drives → 32 + 1 footer = 33 (+1 safety = 34)
//   Medium: portrait 22r drives → 22 + 1 footer = 23 (+1 safety = 24)
//   Tiny  : portrait 14r drives → 14 + 1 footer = 15 (+1 safety = 16, but we want
//           the bio text fully visible at 24r terminal, so we use the 24 minimum)
const (
	largeMinWidth   = 126
	largeMinHeight  = 34
	mediumMinWidth  = 110
	mediumMinHeight = 26
	tinyMinWidth    = 80
	tinyMinHeight   = 24

	// Frame inner widths per tier. Outer frame width = inner + 4 (chevron + padding).
	bioBoxInnerWidth     = 54 // large + medium
	bioBoxInnerWidthTiny = 38 // tiny
)

// Bio prose lives in a chevron-framed block. Two paragraphs: who I am, what
// the site is. Width is fixed to bioBoxInnerWidth so the chevron frame's
// corners and rails align cleanly; lines should be wrapped to that width
// when authored (no runtime reflow).
//
// The "name" is the animated larry3d figlet (see title.go) and isn't part of
// the prose. The tagline + contact handles were intentionally dropped — the
// contact page handles the latter; the former was redundant once the figlet
// title carried the visual identity.
// bioParagraphs is the full bio used at large + medium tiers (54-col inner).
var bioParagraphs = [][]string{
	{
		"is a developer making small things on the internet.",
		"mostly terminals, browsers second. likes weird",
		"interfaces, slow software, and projects that don't",
		"take themselves too seriously.",
	},
	{
		"this site is a petting zoo over ssh. pick an animal",
		"from the farm or exotic list, foster it for your",
		"session, and care for it with feed / walk / sleep /",
		"pet actions. when you disconnect, the animal goes",
		"back. no accounts, no saves — like the best kind of",
		"foster: brief.",
	},
}

// bioParagraphsTiny is the compact bio for the 80×24 tier (38-col inner).
// Same voice, fewer lines, narrower wrap.
var bioParagraphsTiny = [][]string{
	{
		"developer making small things on the",
		"internet. mostly terminals.",
	},
	{
		"this site is a petting zoo over ssh.",
		"pick an animal, foster it for your",
		"session, then let go — it goes back.",
	},
}

// Nav menu items, rendered below the bio. The bio "page" is implicit (you're
// already here), so it isn't a menu option — that would be redundant.
type navItem struct {
	label       string
	target      string
	description string
}

var navItems = []navItem{
	{"adopt", "adopt", "foster an ascii animal (zoo opens soon)"},
	{"projects", "projects", "things I'm building"},
	{"contact", "contact", "say hi"},
}

type Model struct {
	width  int
	height int
	cursor int // index into navItems
	title  titleAnim
}

func New() Model {
	return Model{title: newTitleAnim()}
}

func (m Model) Init() tea.Cmd { return m.title.tick() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case titleTickMsg:
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
// terminal. Below the tiny tier (80×24), we show a resize-prompt instead of
// attempting a layout. Within a tier, the output is exactly tierHeight rows
// tall — animation re-paints can't scroll the terminal.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	const leftMargin = 4
	indent := strings.Repeat(" ", leftMargin)

	// Below the tiny floor: show resize prompt centered.
	if m.width < tinyMinWidth || m.height < tinyMinHeight {
		msg := footerStyle.Render("please resize your terminal to at least 80×24")
		// Pad above so the message sits roughly center-ish.
		topPad := m.height / 2
		if topPad < 1 {
			topPad = 1
		}
		var rows []string
		for i := 0; i < topPad; i++ {
			rows = append(rows, "")
		}
		rows = append(rows, indent+msg)
		// Pad to exact terminal height so no scroll.
		for len(rows) < m.height {
			rows = append(rows, "")
		}
		return strings.Join(rows[:m.height], "\n")
	}

	// Pick tier (largest that fits).
	var (
		portraitArt string
		paragraphs  [][]string
		innerWidth  int
		gapWidth    int
		tier        titleTier
	)
	switch {
	case m.width >= largeMinWidth && m.height >= largeMinHeight:
		portraitArt = portraitLarge
		paragraphs = bioParagraphs
		innerWidth = bioBoxInnerWidth
		gapWidth = 4
		tier = titleLarge
	case m.width >= mediumMinWidth && m.height >= mediumMinHeight:
		portraitArt = portraitMedium
		paragraphs = bioParagraphs
		innerWidth = bioBoxInnerWidth
		gapWidth = 4
		tier = titleLarge // graffiti fits in medium (6r text col ≤ 22r portrait)
	default:
		// Tiny tier (80×24 floor).
		portraitArt = portraitSmall
		paragraphs = bioParagraphsTiny
		innerWidth = bioBoxInnerWidthTiny
		gapWidth = 3
		tier = titleTiny
	}

	body := m.renderSideBySide(portraitArt, paragraphs, innerWidth, gapWidth, tier)

	// Clamp to terminal height — never emit more rows than the terminal has.
	// This is the single guarantee that prevents scroll-induced phantom frames
	// on every animation tick.
	body = clampToHeight(body, m.height)

	// Apply left margin.
	return indent + strings.ReplaceAll(body, "\n", "\n"+indent)
}

// renderSideBySide is the only layout function — portrait left, text right.
// Builds a fixed-row-count block: portrait_rows + 1 blank + footer = total.
// The text column is padded vertically to match the portrait, with the menu
// pinned to the portrait's bottom row.
func (m Model) renderSideBySide(art string, paragraphs [][]string, innerWidth, gapWidth int, tier titleTier) string {
	portrait := portraitStyle.Render(strings.TrimRight(art, "\n"))
	portraitH := strings.Count(strings.TrimRight(art, "\n"), "\n") + 1

	text := m.renderTextColumn(paragraphs, innerWidth, tier, portraitH)

	gap := strings.Repeat(" ", gapWidth)
	row := lipgloss.JoinHorizontal(lipgloss.Top, portrait, gap, text)

	footer := footerStyle.Render("←→ to move  ·  enter to open  ·  q to quit")
	return lipgloss.JoinVertical(lipgloss.Left, row, "", footer)
}

// renderTextColumn composes the right side: title + blank + framed bio +
// (optional pad) + menu, sized to exactly padToHeight rows.
func (m Model) renderTextColumn(paragraphs [][]string, innerWidth int, tier titleTier, padToHeight int) string {
	titleStr := m.title.render(tier)
	frame := renderBioBox(paragraphs, innerWidth)
	menu := m.renderMenu()

	titleLines := strings.Count(titleStr, "\n") + 1
	frameLines := strings.Count(frame, "\n") + 1

	// Natural rows: title + 1 blank + frame + menu.
	naturalRows := titleLines + 1 + frameLines + 1

	rows := []string{titleStr, "", frame}
	if padToHeight > naturalRows {
		extra := padToHeight - naturalRows
		for i := 0; i < extra; i++ {
			rows = append(rows, "")
		}
	}
	rows = append(rows, menu)
	return strings.Join(rows, "\n")
}

// clampToHeight ensures the rendered body has exactly maxRows lines. If too
// short, pads with blank lines; if too long, truncates. Either way, the
// returned string contains exactly maxRows-1 newlines.
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

// renderMenu lays out nav items horizontally with bracket markers around the
// focused item. Sits inside the text column so it's flush below the bio text,
// not below the side-by-side block as a whole.
func (m Model) renderMenu() string {
	var parts []string
	for i, it := range navItems {
		if i == m.cursor {
			parts = append(parts, focusedLabelStyle.Render("["+it.label+"]"))
		} else {
			// Pad unfocused items so spacing matches the focused [label] form
			// (two extra chars from the brackets), keeping items in stable
			// columns as the cursor moves.
			parts = append(parts, labelStyle.Render(" "+it.label+" "))
		}
	}
	return strings.Join(parts, "  ")
}

var (
	portraitStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Nav uses the same green family as the title — focused = bright/bold,
	// unfocused = dim green so they still read as siblings of the logo.
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34")) // dim green
	descStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	focusedMarkerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true) // bright green
	focusedLabelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	focusedDescStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)
