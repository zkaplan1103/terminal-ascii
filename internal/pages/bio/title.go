package bio

import (
	_ "embed"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//go:embed title.txt
var titleArt string

//go:embed title_small.txt
var titleArtSmall string

// titleTier selects which font variant to render.
type titleTier int

const (
	titleLarge  titleTier = iota // graffiti — large + medium tiers
	titleTiny                    // smslant  — tiny tier (80×24)
)

// titleAnim drives the animated "zack" title: a green wavefront sweeps left to
// right across the columns of the figlet art, holds, flickers, then resets to
// white. The whole thing is column-driven, so vertical strokes of letters all
// change color in unison — feels like a solid wall passing through.
//
// The sweep cycle length is always driven by the LARGER art's column count so
// the timing is consistent regardless of which variant is currently displayed.
type titleAnim struct {
	phase    animPhase
	progress int      // sweep: 0..maxCol  |  flicker: 0..len(flickerColors)
	maxCol   int      // width of the large art (drives sweep timing)
	lines    []string // graffiti — large + medium tiers
	linesSm  []string // smslant  — tiny tier
}

type animPhase int

const (
	phaseSweep animPhase = iota
	phaseHoldGreen
	phaseFlicker
	phaseHoldWhite
)

// Tunables. All in milliseconds.
const (
	sweepStepMs   = 60   // ms per column during sweep
	holdGreenMs   = 600  // ms holding all-green
	flickerStepMs = 80   // ms per flicker frame
	holdWhiteMs   = 4500 // ms resting at all-white before next cycle
)

var flickerColors = []lipgloss.Color{
	lipgloss.Color("46"),  // bright green
	lipgloss.Color("252"), // white
	lipgloss.Color("46"),
	lipgloss.Color("245"), // dim
	lipgloss.Color("46"),
}

// titleTickMsg is emitted on each animation tick. The model handles the phase
// transitions and schedules the next tick with the appropriate delay.
type titleTickMsg struct{}

func newTitleAnim() titleAnim {
	lines := strings.Split(strings.TrimRight(titleArt, "\n"), "\n")
	linesSm := strings.Split(strings.TrimRight(titleArtSmall, "\n"), "\n")
	maxCol := 0
	for _, l := range lines {
		if len(l) > maxCol {
			maxCol = len(l)
		}
	}
	return titleAnim{
		phase:   phaseSweep,
		lines:   lines,
		linesSm: linesSm,
		maxCol:  maxCol,
	}
}

// tick returns a tea.Cmd that fires the next titleTickMsg after the delay for
// the current phase. Returned from Init and from Update on every titleTickMsg.
func (a titleAnim) tick() tea.Cmd {
	var delay time.Duration
	switch a.phase {
	case phaseSweep:
		delay = sweepStepMs * time.Millisecond
	case phaseHoldGreen:
		delay = holdGreenMs * time.Millisecond
	case phaseFlicker:
		delay = flickerStepMs * time.Millisecond
	case phaseHoldWhite:
		delay = holdWhiteMs * time.Millisecond
	}
	return tea.Tick(delay, func(time.Time) tea.Msg { return titleTickMsg{} })
}

// advance progresses the animation by one tick. Returns the new state.
func (a titleAnim) advance() titleAnim {
	switch a.phase {
	case phaseSweep:
		a.progress++
		if a.progress > a.maxCol {
			a.phase = phaseHoldGreen
			a.progress = 0
		}
	case phaseHoldGreen:
		a.phase = phaseFlicker
		a.progress = 0
	case phaseFlicker:
		a.progress++
		if a.progress >= len(flickerColors) {
			a.phase = phaseHoldWhite
			a.progress = 0
		}
	case phaseHoldWhite:
		a.phase = phaseSweep
		a.progress = 0
	}
	return a
}

// render colorizes the title art according to the current phase + progress.
func (a titleAnim) render(tier titleTier) string {
	lines := a.lines
	if tier == titleTiny {
		lines = a.linesSm
	}
	switch a.phase {
	case phaseSweep:
		return a.renderSweep(lines, a.progress)
	case phaseHoldGreen:
		return titleStyleGreen.Render(strings.Join(lines, "\n"))
	case phaseFlicker:
		c := flickerColors[a.progress]
		st := lipgloss.NewStyle().Foreground(c).Bold(true)
		return st.Render(strings.Join(lines, "\n"))
	case phaseHoldWhite:
		return titleStyleWhite.Render(strings.Join(lines, "\n"))
	}
	return strings.Join(lines, "\n")
}

// renderSweep colors columns 0..wavefront-1 green, the rest white.
// Wavefront is in the LARGER art's column space — for the small variant it
// gets proportionally scaled so the sweep visually maps to the same fraction
// of the title regardless of which variant is being drawn.
func (a titleAnim) renderSweep(lines []string, wavefront int) string {
	// Scale wavefront to the displayed art's column count.
	displayCol := wavefront
	if a.maxCol > 0 {
		smMax := 0
		for _, l := range lines {
			if len(l) > smMax {
				smMax = len(l)
			}
		}
		// Map progress (0..maxCol) → (0..smMax) for the displayed lines.
		displayCol = int(float64(wavefront) * float64(smMax) / float64(a.maxCol))
	}
	var out strings.Builder
	for li, line := range lines {
		var green, white string
		if displayCol <= 0 {
			white = line
		} else if displayCol >= len(line) {
			green = line
		} else {
			green = line[:displayCol]
			white = line[displayCol:]
		}
		if green != "" {
			out.WriteString(titleStyleGreen.Render(green))
		}
		if white != "" {
			out.WriteString(titleStyleWhite.Render(white))
		}
		if li < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

var (
	titleStyleWhite = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)
	titleStyleGreen = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
)
