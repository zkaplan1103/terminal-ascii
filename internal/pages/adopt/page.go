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
	viewPickSpecies adoptView = iota // pick a species (dogs, cats, horses…)
	viewPickAnimal                   // pick a named animal within that species
	viewCare                         // care view with animation
)

// zooTickMsg fires every second to advance session stats.
type zooTickMsg struct{ now time.Time }

func zooTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return zooTickMsg{now: t} })
}

// overlayMsg is sent by the care view to trigger a timed overlay burst.
type overlayMsg struct {
	kind    overlayKind
	framesLeft int
}

type overlayKind int

const (
	overlayNone     overlayKind = iota
	overlayHearts               // ♥ on pet
	overlaySparkles             // ✦ on walk
	overlayFood                 // 🍎 on feed
	overlayZzz                  // z z Z on sleep
)

// overlay holds the current active overlay and remaining display frames.
type overlay struct {
	kind       overlayKind
	framesLeft int
	tick       int // increments each animation tick, drives frame cycling
}

func (o overlay) active() bool { return o.kind != overlayNone && o.framesLeft > 0 }

func (o overlay) advance() overlay {
	if o.framesLeft > 0 {
		o.framesLeft--
		o.tick++
	}
	if o.framesLeft == 0 {
		o.kind = overlayNone
	}
	return o
}

// render returns the overlay string for the current frame.
func (o overlay) render() string {
	if !o.active() {
		return ""
	}
	switch o.kind {
	case overlayHearts:
		frames := []string{
			"  ♥       ",
			"  ♥  ♥    ",
			"♥ ♥  ♥    ",
			"♥ ♥  ♥  ♥ ",
		}
		return heartStyle.Render(frames[o.tick%len(frames)])
	case overlaySparkles:
		frames := []string{
			"  ✦       ",
			"  ✦  ✦    ",
			"✦ ✦  ✦    ",
			"✦ ✦  ✦  ✦ ",
		}
		return sparkleStyle.Render(frames[o.tick%len(frames)])
	case overlayFood:
		frames := []string{
			"  🍎      ",
			"  🍎 🥕   ",
			"🌽🍎 🥕   ",
		}
		return foodStyle.Render(frames[o.tick%len(frames)])
	case overlayZzz:
		frames := []string{"z", "z z", "z z Z", "z z Z"}
		return zzzStyle.Render(frames[o.tick%len(frames)])
	}
	return ""
}

// Model is the adopt page's Bubble Tea model.
type Model struct {
	width  int
	height int

	view         adoptView
	speciesIdx   int      // cursor in species picker
	speciesList  []string // ordered list of species keys
	animalIdx    int      // cursor in animal picker
	animals      []*zoo.Animal

	session *zoo.Session
	animal  ui.AnimalView
	ov      overlay
}

func New() Model {
	return Model{
		speciesList: zoo.SpeciesOrder,
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
			m.ov = m.ov.advance()
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

	// ── Species picker ────────────────────────────────────────────────────────
	case viewPickSpecies:
		switch key {
		case "up", "k":
			if m.speciesIdx > 0 {
				m.speciesIdx--
			}
		case "down", "j":
			if m.speciesIdx < len(m.speciesList)-1 {
				m.speciesIdx++
			}
		case "enter", " ":
			species := m.speciesList[m.speciesIdx]
			bySpecies := zoo.BySpecies()
			m.animals = bySpecies[species]
			m.animalIdx = 0
			m.view = viewPickAnimal
		case "esc":
			return m, func() tea.Msg { return page.NavigateMsg{To: "bio"} }
		}

	// ── Animal picker ─────────────────────────────────────────────────────────
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
			m.view = viewPickSpecies
		}

	// ── Care view ─────────────────────────────────────────────────────────────
	case viewCare:
		switch key {
		case "f":
			if m.session != nil {
				m.session.Feed(time.Now())
				m.ov = overlay{kind: overlayFood, framesLeft: 12}
			}
		case "w":
			if m.session != nil {
				m.session.Walk(time.Now())
				m.ov = overlay{kind: overlaySparkles, framesLeft: 12}
			}
		case "s":
			if m.session != nil {
				if m.session.State == zoo.StateSleeping {
					m.session.WakeUp()
					m.ov = overlay{}
				} else {
					m.session.Sleep()
					m.ov = overlay{kind: overlayZzz, framesLeft: 999} // persists while sleeping
				}
			}
		case "p":
			if m.session != nil {
				m.session.Pet()
				m.ov = overlay{kind: overlayHearts, framesLeft: 12}
			}
		case "esc":
			m.session = nil
			m.ov = overlay{}
			m.view = viewPickSpecies
			m.speciesIdx = 0
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
	case viewPickSpecies:
		content = m.viewSpecies()
	case viewPickAnimal:
		content = m.viewAnimal()
	case viewCare:
		content = m.viewCare()
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// ── Species picker ────────────────────────────────────────────────────────────

func (m Model) viewSpecies() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("foster an animal"))
	b.WriteString("\n\n")

	bySpecies := zoo.BySpecies()
	for i, species := range m.speciesList {
		animals := bySpecies[species]
		if len(animals) == 0 {
			continue
		}
		label := zoo.SpeciesLabel(species)
		count := fmt.Sprintf("(%d)", len(animals))
		if i == m.speciesIdx {
			b.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %-12s %s", label, count)))
		} else {
			b.WriteString(itemStyle.Render(fmt.Sprintf("  %-12s %s", label, count)))
		}
		b.WriteByte('\n')
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("↑↓ to move  ·  enter to pick  ·  esc to home  ·  q to quit"))
	return b.String()
}

// ── Animal picker ─────────────────────────────────────────────────────────────

func (m Model) viewAnimal() string {
	if len(m.animals) == 0 {
		return ""
	}
	species := zoo.SpeciesLabel(m.animals[0].Species)

	var b strings.Builder
	b.WriteString(headerStyle.Render("foster an animal"))
	b.WriteString("\n")
	b.WriteString(subStyle.Render(strings.ToLower(species)))
	b.WriteString("\n\n")

	for i, a := range m.animals {
		if i == m.animalIdx {
			b.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %-12s  %s", a.Name, a.Breed)))
		} else {
			b.WriteString(itemStyle.Render(fmt.Sprintf("  %-12s  %s", a.Name, a.Breed)))
		}
		b.WriteByte('\n')
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("↑↓ to move  ·  enter to foster  ·  esc to species  ·  q to quit"))
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
	b.WriteString(subStyle.Render(fmt.Sprintf("%s  ·  %s", a.Breed, a.Species)))
	b.WriteString("\n\n")

	// Overlay (hearts, sparkles, food) — shown above the animal
	if m.ov.active() && m.ov.kind != overlayZzz {
		b.WriteString(m.ov.render())
		b.WriteString("\n")
	}

	// Animated animal frame
	frame := m.animal.View()
	if frame != "" {
		b.WriteString(animalFrameStyle.Render(frame))
	}

	// Zzz overlay — shown to the right of the animal on the same lines
	if m.ov.active() && m.ov.kind == overlayZzz {
		b.WriteString("  " + m.ov.render())
	}
	b.WriteString("\n\n")

	// State label
	b.WriteString(stateStyle.Render(fmt.Sprintf("[ %s ]", s.State.String())))
	b.WriteString("\n\n")

	// Stat bars — hunger: high is bad. energy/happiness: low is bad.
	b.WriteString(statLine("hunger   ", s.Hunger, true))
	b.WriteString("\n")
	b.WriteString(statLine("energy   ", s.Energy, false))
	b.WriteString("\n")
	b.WriteString(statLine("happiness", s.Happiness, false))
	b.WriteString("\n\n")

	// Actions
	sleepLabel := "sleep"
	if s.State == zoo.StateSleeping {
		sleepLabel = "wake "
	}
	b.WriteString(actionStyle.Render("f") + dimStyle.Render(" feed    "))
	b.WriteString(actionStyle.Render("w") + dimStyle.Render(" walk    "))
	b.WriteString(actionStyle.Render("s") + dimStyle.Render(fmt.Sprintf(" %s   ", sleepLabel)))
	b.WriteString(actionStyle.Render("p") + dimStyle.Render(" pet"))
	b.WriteString("\n\n")

	b.WriteString(hintStyle.Render("esc to release  ·  q to quit"))
	return b.String()
}

// statLine renders a labeled progress bar for a 0..100 value.
// highIsBad flips the color logic: for hunger, a high value is bad (red).
// For energy/happiness, a low value is bad (red).
func statLine(label string, val int, highIsBad bool) string {
	const barWidth = 20
	filled := val * barWidth / 100
	if filled > barWidth {
		filled = barWidth
	}

	// Pick bar colour based on how "healthy" the value is.
	// goodness = how good the value is (0..100 regardless of direction).
	goodness := val
	if highIsBad {
		goodness = 100 - val
	}
	var barColor lipgloss.Color
	switch {
	case goodness >= 60:
		barColor = lipgloss.Color("46")  // bright green — good
	case goodness >= 30:
		barColor = lipgloss.Color("214") // amber — warning
	default:
		barColor = lipgloss.Color("196") // red — critical
	}

	filledBar := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", filled))
	emptyBar := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render(strings.Repeat("░", barWidth-filled))
	pct := fmt.Sprintf("%3d%%", val)
	return labelStatStyle.Render(label) + " " + filledBar + emptyBar + " " + pctStyle.Render(pct)
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

	pctStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	actionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	heartStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	sparkleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

	foodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	zzzStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63")).
			Italic(true)
)
