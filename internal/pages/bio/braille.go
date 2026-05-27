package bio

import (
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// brailleField renders the right-side braille texture for one row of text.
//
// The texture is sparse near the text (wisp → light → mid → dense) as you
// move rightward toward the hard edge. A fixed per-row seed makes the texture
// stable across animation frames — the same dots appear in the same positions
// every render; only their colour changes when glowing.
//
// Glow state is communicated via glowRow: the index (0-based, counting all
// text+divider rows top to bottom) of the current wavefront. A row is "lit"
// when its own index <= glowRow. -1 means nothing is lit.

// brailleChars grouped by density (sparse → dense).
var (
	brailleWisp  = []rune("⠁⠂⠄⠈⠐⠠⡀⢀")
	brailleLight = []rune("⠌⠊⠑⠒⠔⠘⠡⠢⠤⠥⠦⠧⠨⠩⠪")
	brailleMid   = []rune("⠿⠾⠽⠻⠷⠯⠟⢻⡻⢽⡽⢷⡷⢯⡯⠶⠵⠳⠼⠸⠰⢸⡸⢰⡰")
	brailleDense = []rune("⣿⢿⡿⣾⣽⣻⣷⣯⣟⣺⣹⣸⢾⡾⣶⣴⣲⣱⣰⢶⡶⣤⣣⣢⣡⣠⢴⡴⢲⡲⢱⡱⢤⡤⣾⣽⣻")
)

func brailleChar(rng *rand.Rand, frac float64) rune {
	switch {
	case frac < 0.20:
		return brailleWisp[rng.Intn(len(brailleWisp))]
	case frac < 0.40:
		return brailleLight[rng.Intn(len(brailleLight))]
	case frac < 0.65:
		return brailleMid[rng.Intn(len(brailleMid))]
	default:
		return brailleDense[rng.Intn(len(brailleDense))]
	}
}

// brailleContentRow renders one content row's braille fill.
//
//   - textLen: byte length of the text already placed on this row
//   - totalWidth: hard right edge (cols)
//   - pad: minimum spaces between text and first braille char
//   - rowSeed: deterministic seed for this row (row index * prime)
//   - lit: whether this row is in the glow wavefront
//   - flickerColor: if non-empty, override the lit colour (flicker phase)
func brailleContentRow(textLen, totalWidth, pad, rowSeed int, lit bool, flickerColor lipgloss.Color) string {
	rng := rand.New(rand.NewSource(int64(rowSeed)))

	// Jagged left edge: 0–2 extra spaces, seeded so it's stable.
	jag := rng.Intn(3)
	brailleStart := textLen + pad + jag
	brailleWidth := totalWidth - brailleStart

	if brailleWidth <= 0 {
		return ""
	}

	// Build the raw rune slice for this row.
	runes := make([]rune, brailleWidth)
	for i := range runes {
		frac := float64(i) / float64(brailleWidth)
		runes[i] = brailleChar(rng, frac)
	}
	raw := string(runes)

	// Colour: lit rows glow green (or flicker colour), unlit rows are dim.
	if lit {
		c := lipgloss.Color("46") // bright green
		if flickerColor != "" {
			c = flickerColor
		}
		return strings.Repeat(" ", pad+jag) + lipgloss.NewStyle().Foreground(c).Render(raw)
	}
	return strings.Repeat(" ", pad+jag) + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(raw)
}

// brailleDividerRow renders a full-width divider row (between sections).
// Dividers are always wisp-only — barely there, just scattered dots.
// They glow the same as content rows when lit.
func brailleDividerRow(totalWidth, rowSeed int, lit bool, flickerColor lipgloss.Color) string {
	rng := rand.New(rand.NewSource(int64(rowSeed)))

	runes := make([]rune, totalWidth)
	for i := range runes {
		frac := float64(i) / float64(totalWidth)
		// Dividers: wisp left half, light right half — never reach mid/dense.
		if frac < 0.55 {
			runes[i] = brailleWisp[rng.Intn(len(brailleWisp))]
		} else {
			runes[i] = brailleLight[rng.Intn(len(brailleLight))]
		}
	}
	raw := string(runes)

	if lit {
		c := lipgloss.Color("46")
		if flickerColor != "" {
			c = flickerColor
		}
		return lipgloss.NewStyle().Foreground(c).Render(raw)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render(raw) // dimmer than content
}

// braillePadRow renders one padding row in the gap between bio and nav.
// Each row gets a distinct density profile based on its position in the zone
// so the padding block looks organic rather than a uniform band.
//
// Pattern across the padding zone (top to bottom):
//   - First few rows: wisp-only across full width (tailing off from bio)
//   - Middle rows: sparse left, a small mid/light patch somewhere in the middle
//   - Last row before divider: wisp-only again (bleeding into the divider)
func braillePadRow(totalWidth, seed, padIdx, totalPad int, lit bool, flickerColor lipgloss.Color) string {
	rng := rand.New(rand.NewSource(int64(seed)))

	// Position within the padding zone: 0.0 = just after bio, 1.0 = just before nav.
	pos := 0.0
	if totalPad > 1 {
		pos = float64(padIdx) / float64(totalPad-1)
	}

	// Each row has a random "peak density column" — where the densest char
	// appears. Seeded so it's stable. This breaks the even-gradient look.
	peakFrac := 0.5 + rng.Float64()*0.4 // peak always in right half
	peakWidth := 0.08 + rng.Float64()*0.12

	runes := make([]rune, totalWidth)
	for i := range runes {
		frac := float64(i) / float64(totalWidth)

		// Distance from the peak determines local density.
		dist := abs64(frac - peakFrac)
		localDensity := 1.0 - dist/peakWidth // 1.0 at peak, <0 outside peak
		if localDensity < 0 {
			localDensity = 0
		}

		// Scale peak density by position in padding zone:
		// rows near bio or nav get lower peak, middle rows get higher.
		// Use a tent function: peaks at pos=0.5.
		tent := 1.0 - abs64(pos-0.5)*2.0 // 1.0 at middle, 0 at edges
		localDensity *= tent * 0.6        // cap so it never gets dense

		switch {
		case localDensity > 0.4:
			runes[i] = brailleMid[rng.Intn(len(brailleMid))]
		case localDensity > 0.15:
			runes[i] = brailleLight[rng.Intn(len(brailleLight))]
		case localDensity > 0.02:
			runes[i] = brailleWisp[rng.Intn(len(brailleWisp))]
		default:
			runes[i] = ' '
		}
	}
	raw := string(runes)

	if lit {
		c := lipgloss.Color("46")
		if flickerColor != "" {
			c = flickerColor
		}
		return lipgloss.NewStyle().Foreground(c).Render(raw)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render(raw)
}

func abs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// rowSeed produces a stable per-row seed that won't collide across rows.
func rowSeed(rowIdx int) int {
	return (rowIdx + 1) * 104729 // large prime keeps seeds well-spread
}
