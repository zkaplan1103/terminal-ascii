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
	titleLarge titleTier = iota // graffiti — large + medium tiers
	titleTiny                   // smslant  — tiny tier (80×24)
)

// animPhase is the state machine for the combined title + braille animation.
//
// Sequence:
//
//	phaseSweep      title sweeps left→right (green wavefront by column)
//	phaseHoldGreen  title holds all-green; braille drop is about to start
//	phaseBrailDrop  braille glow sweeps top→bottom row by row; title stays green
//	phaseFlicker    BOTH title + braille flash together
//	phaseHoldWhite  both return to dim/white; long pause before next cycle
type animPhase int

const (
	phaseSweep animPhase = iota
	phaseHoldGreen
	phaseBrailDrop
	phaseFlicker
	phaseHoldWhite
)

// Tunables — all in milliseconds.
const (
	sweepStepMs   = 60   // ms per column during title sweep
	holdGreenMs   = 200  // ms holding all-green before braille drop starts
	brailStepMs   = 55   // ms per row during braille drop
	flickerStepMs = 80   // ms per flicker frame
	holdWhiteMs   = 4500 // ms at rest before next cycle
)

var flickerColors = []lipgloss.Color{
	lipgloss.Color("46"),  // bright green
	lipgloss.Color("252"), // white
	lipgloss.Color("46"),
	lipgloss.Color("245"), // dim
	lipgloss.Color("46"),
}

// titleTickMsg fires on every animation tick.
type titleTickMsg struct{}

// titleAnim holds all animation state. brailRows is set by the page on first
// render (it depends on the tier's row count) and stays constant after that.
type titleAnim struct {
	phase     animPhase
	progress  int      // multi-purpose: column index / flicker index / braille row index
	maxCol    int      // width of the large art (drives sweep timing)
	brailRows int      // total rows in the braille field (set by page)
	lines     []string // graffiti — large + medium tiers
	linesSm   []string // smslant  — tiny tier
}

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

// tick returns the next tea.Cmd with the appropriate delay for the current phase.
func (a titleAnim) tick() tea.Cmd {
	var delay time.Duration
	switch a.phase {
	case phaseSweep:
		delay = sweepStepMs * time.Millisecond
	case phaseHoldGreen:
		delay = holdGreenMs * time.Millisecond
	case phaseBrailDrop:
		delay = brailStepMs * time.Millisecond
	case phaseFlicker:
		delay = flickerStepMs * time.Millisecond
	case phaseHoldWhite:
		delay = holdWhiteMs * time.Millisecond
	}
	return tea.Tick(delay, func(time.Time) tea.Msg { return titleTickMsg{} })
}

// advance moves the animation forward by one tick.
func (a titleAnim) advance() titleAnim {
	switch a.phase {
	case phaseSweep:
		a.progress++
		if a.progress > a.maxCol {
			a.phase = phaseHoldGreen
			a.progress = 0
		}
	case phaseHoldGreen:
		a.phase = phaseBrailDrop
		a.progress = 0
	case phaseBrailDrop:
		a.progress++
		// brailRows is set by renderBrailleColumn on every View() call before
		// advance() is invoked, so it should always be current. Guard zero only
		// for the very first tick before any render has happened.
		total := a.brailRows
		if total == 0 {
			total = 40 // generous fallback — better to overshoot than stop early
		}
		if a.progress >= total {
			a.phase = phaseFlicker
			a.progress = 0
		}
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

// titleGlowRow returns the current braille wavefront row index.
//   - During phaseBrailDrop: 0..brailRows-1
//   - During phaseHoldGreen: -1 (drop not yet started)
//   - During phaseFlicker/phaseHoldWhite: brailRows (all rows lit / all reset)
//   - During phaseSweep: -1
func (a titleAnim) titleGlowRow() int {
	switch a.phase {
	case phaseBrailDrop:
		return a.progress
	case phaseFlicker:
		return a.brailRows // all rows lit
	default:
		return -1 // nothing lit
	}
}

// flickerColor returns the current flicker override colour, or "" if not flickering.
func (a titleAnim) flickerColor() lipgloss.Color {
	if a.phase == phaseFlicker {
		return flickerColors[a.progress]
	}
	return ""
}

// render returns the colourised title string for this tick.
func (a titleAnim) render(tier titleTier) string {
	lines := a.lines
	if tier == titleTiny {
		lines = a.linesSm
	}
	switch a.phase {
	case phaseSweep:
		return a.renderSweep(lines, a.progress)
	case phaseHoldGreen, phaseBrailDrop:
		// Title stays solid green while braille drops.
		return titleStyleGreen.Render(strings.Join(lines, "\n"))
	case phaseFlicker:
		c := flickerColors[a.progress]
		return lipgloss.NewStyle().Foreground(c).Bold(true).Render(strings.Join(lines, "\n"))
	case phaseHoldWhite:
		return titleStyleWhite.Render(strings.Join(lines, "\n"))
	}
	return strings.Join(lines, "\n")
}

// renderSweep colours columns 0..wavefront-1 green, the rest white.
// Wavefront is in the large art's column space; scaled for small variant.
func (a titleAnim) renderSweep(lines []string, wavefront int) string {
	displayCol := wavefront
	if a.maxCol > 0 {
		smMax := 0
		for _, l := range lines {
			if len(l) > smMax {
				smMax = len(l)
			}
		}
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
