package adopt

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zkaplan/terminal-site/internal/page"
	"github.com/zkaplan/terminal-site/internal/ui"
	"github.com/zkaplan/terminal-site/internal/zoo"
)

// adoptView tracks which sub-view is active.
type adoptView int

const (
	viewPickCategory adoptView = iota
	viewPickAnimal
	viewCare
)

// zooTickMsg fires every second to advance session stats.
type zooTickMsg struct{ now time.Time }

func zooTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return zooTickMsg{now: t} })
}

// Model is the adopt page's Bubble Tea model.
type Model struct {
	width  int
	height int

	view     adoptView
	catIdx   int      // cursor in category picker
	cats     []string // ["farm", "exotic"]
	animalIdx int     // cursor in animal picker
	animals  []*zoo.Animal

	session *zoo.Session
	animal  ui.AnimalView
}

func New() Model {
	return Model{
		cats: []string{"farm", "exotic"},
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case zooTickMsg:
		if m.session != nil {
			m.session.Tick(msg.now)
		}
		return m, zooTick()

	case ui.AnimalTickMsg:
		if m.view == viewCare {
			var cmd tea.Cmd
			m.animal, cmd = m.animal.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.view {

	// ── Category picker ──────────────────────────────────────────
	case viewPickCategory:
		switch key {
		case "up", "k":
			if m.catIdx > 0 {
				m.catIdx--
			}
		case "down", "j":
			if m.catIdx < len(m.cats)-1 {
				m.catIdx++
			}
		case "enter", " ":
			cat := m.cats[m.catIdx]
			order := zoo.FarmOrder
			if cat == "exotic" {
				order = zoo.ExoticOrder
			}
			bycat := zoo.ByCategory()
			catAnimals := bycat[cat]
			// Sort by the defined order.
			sorted := make([]*zoo.Animal, 0, len(catAnimals))
			for _, species := range order {
				for _, a := range catAnimals {
					if a.Species == species {
						sorted = append(sorted, a)
						break
					}
				}
			}
			m.animals = sorted
			m.animalIdx = 0
			m.view = viewPickAnimal
		case "esc":
			return m, func() tea.Msg { return page.NavigateMsg{To: "bio"} }
		}

	// ── Animal picker ─────────────────────────────────────────────
	case viewPickAnimal:
		switch key {
		case "up", "k":
			if m.animalIdx > 0 {
				m.animalIdx--
			}
		case "down", "j":
			if m.animalIdx < len(m.animals)-1 {
				m.animalIdx++
			}
		case "enter", " ":
			chosen := m.animals[m.animalIdx]
			m.session = zoo.NewSession(chosen)
			m.animal = ui.NewAnimalView(m.session)
			m.view = viewCare
			return m, tea.Batch(m.animal.Init(), zooTick())
		case "esc":
			m.view = viewPickCategory
		}

	// ── Care view ─────────────────────────────────────────────────
	case viewCare:
		switch key {
		case "f":
			if m.session != nil {
				m.session.Feed(time.Now())
			}
		case "w":
			if m.session != nil {
				m.session.Walk(time.Now())
			}
		case "s":
			if m.session != nil {
				if m.session.State == zoo.StateSleeping {
					m.session.WakeUp()
				} else {
					m.session.Sleep()
				}
			}
		case "p":
			if m.session != nil {
				m.session.Pet()
			}
		case "esc":
			// Release the animal, go back to category picker.
			m.session = nil
			m.view = viewPickCategory
			m.catIdx = 0
			m.animalIdx = 0
		}
	}

	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var content string
	switch m.view {
	case viewPickCategory:
		content = m.viewCategory()
	case viewPickAnimal:
		content = m.viewAnimal()
	case viewCare:
		content = m.viewCare()
	}

	// Center in terminal.
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// ── Category picker ───────────────────────────────────────────────────────────

func (m Model) viewCategory() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("foster an animal"))
	b.WriteString("\n\n")
	b.WriteString(subStyle.Render("choose a category"))
	b.WriteString("\n\n")

	for i, cat := range m.cats {
		icon := catIcons[cat]
		label := fmt.Sprintf("  %s  %s", icon, cat)
		if i == m.catIdx {
			b.WriteString(selectedStyle.Render("▶ " + fmt.Sprintf("%s  %s", icon, cat)))
		} else {
			b.WriteString(itemStyle.Render(label))
		}
		b.WriteByte('\n')
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("↑↓ to move  ·  enter to pick  ·  esc to home  ·  q to quit"))
	return b.String()
}

var catIcons = map[string]string{
	"farm":   "🌾",
	"exotic": "🌿",
}

// ── Animal picker ─────────────────────────────────────────────────────────────

func (m Model) viewAnimal() string {
	cat := m.cats[m.catIdx]
	var b strings.Builder
	b.WriteString(headerStyle.Render("foster an animal"))
	b.WriteString("\n\n")
	b.WriteString(subStyle.Render(fmt.Sprintf("%s animals", cat)))
	b.WriteString("\n\n")

	for i, a := range m.animals {
		name := fmt.Sprintf("%-10s  %s", a.Name, a.Species)
		if i == m.animalIdx {
			b.WriteString(selectedStyle.Render("▶ " + name))
		} else {
			b.WriteString(itemStyle.Render("  " + name))
		}
		b.WriteByte('\n')
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("↑↓ to move  ·  enter to foster  ·  esc to categories  ·  q to quit"))
	return b.String()
}

// ── Care view ─────────────────────────────────────────────────────────────────

func (m Model) viewCare() string {
	if m.session == nil {
		return ""
	}
	s := m.session
	a := s.Animal

	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Render(fmt.Sprintf("fostering %s", a.Name)))
	b.WriteString("\n")
	b.WriteString(subStyle.Render(fmt.Sprintf("%s  ·  %s", a.Species, a.Category)))
	b.WriteString("\n\n")

	// Animated animal frame
	frame := m.animal.View()
	if frame != "" {
		b.WriteString(animalFrameStyle.Render(frame))
		b.WriteString("\n\n")
	}

	// State label
	stateStr := s.State.String()
	b.WriteString(stateStyle.Render(fmt.Sprintf("[ %s ]", stateStr)))
	b.WriteString("\n\n")

	// Stat bars
	b.WriteString(statLine("hunger   ", s.Hunger))
	b.WriteString("\n")
	b.WriteString(statLine("energy   ", s.Energy))
	b.WriteString("\n")
	b.WriteString(statLine("happiness", s.Happiness))
	b.WriteString("\n\n")

	// Actions
	b.WriteString(actionStyle.Render("f") + dimStyle.Render(" feed    "))
	b.WriteString(actionStyle.Render("w") + dimStyle.Render(" walk    "))
	sleepLabel := "sleep"
	if s.State == zoo.StateSleeping {
		sleepLabel = "wake"
	}
	b.WriteString(actionStyle.Render("s") + dimStyle.Render(fmt.Sprintf(" %-6s  ", sleepLabel)))
	b.WriteString(actionStyle.Render("p") + dimStyle.Render(" pet"))
	b.WriteString("\n\n")

	b.WriteString(hintStyle.Render("esc to release  ·  q to quit"))
	return b.String()
}

// statLine renders a labeled progress bar for a 0..100 value.
func statLine(label string, val int) string {
	const barWidth = 20
	filled := val * barWidth / 100
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	pct := fmt.Sprintf("%3d%%", val)
	return labelStatStyle.Render(label) + " " + barStyle.Render(bar) + " " + pctStyle.Render(pct)
}

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	subStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	stateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	animalFrameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	labelStatStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	barStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("34"))

	pctStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	actionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)
