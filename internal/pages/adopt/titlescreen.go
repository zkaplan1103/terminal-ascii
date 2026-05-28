package adopt

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zkaplan/terminal-site/internal/zoo"
)

// ── Tick messages ─────────────────────────────────────────────────────────────

type sceneMoveMsg struct{}
type playTickMsg struct{}

func sceneMove() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg { return sceneMoveMsg{} })
}
func playTick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return playTickMsg{} })
}

// ── Play button animation ─────────────────────────────────────────────────────

type playPhase int

const (
	playSweep   playPhase = iota // sweep green left→right col by col
	playHold                     // hold all-green briefly
	playFlicker                  // glitch flash
	playWhite                    // rest dim/white, long pause
)

const (
	playSweepMs   = 55  // ms per column
	playHoldMs    = 200
	playFlickerMs = 80
	playWhiteMs   = 3000
)

var playFlickerColors = []lipgloss.Color{
	lipgloss.Color("46"),
	lipgloss.Color("252"),
	lipgloss.Color("46"),
	lipgloss.Color("245"),
	lipgloss.Color("46"),
}

type playAnim struct {
	phase    playPhase
	progress int // column index during sweep, flicker index otherwise
	maxCol   int // widest line in the play art
}

func newPlayAnim(playStr string) playAnim {
	lines := strings.Split(strings.TrimRight(playStr, "\n"), "\n")
	maxCol := 0
	for _, l := range lines {
		if len([]rune(l)) > maxCol {
			maxCol = len([]rune(l))
		}
	}
	return playAnim{phase: playSweep, maxCol: maxCol}
}

func (p playAnim) tickCmd() tea.Cmd {
	switch p.phase {
	case playSweep:
		return playTick(playSweepMs * time.Millisecond)
	case playHold:
		return playTick(playHoldMs * time.Millisecond)
	case playFlicker:
		return playTick(playFlickerMs * time.Millisecond)
	default:
		return playTick(playWhiteMs * time.Millisecond)
	}
}

func (p playAnim) advance() playAnim {
	switch p.phase {
	case playSweep:
		p.progress++
		if p.progress > p.maxCol {
			p.phase = playHold
			p.progress = 0
		}
	case playHold:
		p.phase = playFlicker
		p.progress = 0
	case playFlicker:
		p.progress++
		if p.progress >= len(playFlickerColors) {
			p.phase = playWhite
			p.progress = 0
		}
	case playWhite:
		p.phase = playSweep
		p.progress = 0
	}
	return p
}

// render returns the colourised PLAY string for this tick.
func (p playAnim) render(lines []string) string {
	switch p.phase {
	case playSweep:
		var out strings.Builder
		for li, line := range lines {
			runes := []rune(line)
			wave := p.progress
			green := ""
			white := ""
			if wave <= 0 {
				white = line
			} else if wave >= len(runes) {
				green = line
			} else {
				green = string(runes[:wave])
				white = string(runes[wave:])
			}
			if green != "" {
				out.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render(green))
			}
			if white != "" {
				out.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render(white))
			}
			if li < len(lines)-1 {
				out.WriteByte('\n')
			}
		}
		return out.String()
	case playHold:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render(strings.Join(lines, "\n"))
	case playFlicker:
		c := playFlickerColors[p.progress]
		return lipgloss.NewStyle().Foreground(c).Bold(true).Render(strings.Join(lines, "\n"))
	default: // playWhite
		return lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render(strings.Join(lines, "\n"))
	}
}

// ── Tier config ───────────────────────────────────────────────────────────────

type tierConfig struct {
	bg        string // braille scene art
	play      string // PLAY figlet text
	sceneRows int    // rows the bg art occupies
	sceneW    int    // columns of the scene art
	// foreground walk lane
	walkRowStart int
	walkRowCount int
	// horizontal gap in foreground bushes
	gapLeft  int
	gapRight int
}

// Animal frames are scaled to 16 cols × 8 rows (1/3 width, 1/2 height of 48×16 originals).
const animalW = 16
const animalH = 8

var tierLarge = tierConfig{
	bg:           sceneLargeBg,
	play:         playLarge,
	sceneRows:    27,
	sceneW:       126,
	walkRowStart: 19,
	walkRowCount: animalH,
	gapLeft:      37,
	gapRight:     94,
}

var tierMedium = tierConfig{
	bg:           sceneMediumBg,
	play:         playLarge,
	sceneRows:    20,
	sceneW:       110,
	walkRowStart: 13,
	walkRowCount: animalH,
	gapLeft:      28,
	gapRight:     82,
}

var tierSmall = tierConfig{
	bg:           sceneSmallBg,
	play:         playSmall,
	sceneRows:    17,
	sceneW:       80,
	walkRowStart: 10,
	walkRowCount: animalH,
	gapLeft:      18,
	gapRight:     62,
}

func tierForSize(w, h int) tierConfig {
	switch {
	case w >= 126 && h >= 34:
		return tierLarge
	case w >= 110 && h >= 26:
		return tierMedium
	default:
		return tierSmall
	}
}

// ── sceneAnim ─────────────────────────────────────────────────────────────────

type sceneAnim struct {
	parade    []string
	paradeIdx int
	x         int
	frameIdx  int
	play      playAnim
}

func newSceneAnim() sceneAnim {
	order := make([]string, len(zoo.SpeciesOrder))
	copy(order, zoo.SpeciesOrder)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })

	// Pick the right play font based on default tier (large); will reinit on first resize.
	pa := newPlayAnim(playLarge)

	return sceneAnim{
		parade: order,
		x:      -animalW,
		play:   pa,
	}
}

func (s sceneAnim) init() tea.Cmd {
	return tea.Batch(sceneMove(), s.play.tickCmd())
}

func (s sceneAnim) currentSpecies() string {
	if len(s.parade) == 0 {
		return "dog"
	}
	return s.parade[s.paradeIdx%len(s.parade)]
}

func (s sceneAnim) currentFrames() []string {
	species := s.currentSpecies()
	bySpecies := zoo.BySpecies()
	animals := bySpecies[species]
	if len(animals) == 0 {
		return nil
	}
	frames, _ := animals[0].FramesFor(zoo.StateWalking)
	// Scale down: keep every other row, every 3rd col → ~16 cols × 8 rows.
	scaled := make([]string, len(frames))
	for fi, frame := range frames {
		rows := strings.Split(strings.TrimRight(frame, "\n"), "\n")
		var sb strings.Builder
		for ri, row := range rows {
			if ri%2 != 0 {
				continue
			}
			runes := []rune(row)
			for ci := 0; ci < len(runes); ci += 3 {
				sb.WriteRune(runes[ci])
			}
			sb.WriteByte('\n')
		}
		scaled[fi] = sb.String()
	}
	return scaled
}

func (s sceneAnim) update(msg tea.Msg, sceneW int, playStr string) (sceneAnim, tea.Cmd) {
	switch msg.(type) {
	case sceneMoveMsg:
		s.x += 2
		if s.x%6 == 0 {
			frames := s.currentFrames()
			if len(frames) > 0 {
				s.frameIdx = (s.frameIdx + 1) % len(frames)
			}
		}
		if s.x > sceneW+animalW {
			s.paradeIdx++
			s.x = -animalW
			s.frameIdx = 0
		}
		return s, sceneMove()
	case playTickMsg:
		s.play = s.play.advance()
		return s, s.play.tickCmd()
	}
	return s, nil
}

// ── Compositor ────────────────────────────────────────────────────────────────

func isBlank(r rune) bool {
	return r == ' ' || r == '⠀' || r == 0
}

func (s sceneAnim) composite(tc tierConfig) []string {
	bgLines := strings.Split(strings.TrimRight(tc.bg, "\n"), "\n")
	for len(bgLines) < tc.sceneRows {
		bgLines = append(bgLines, strings.Repeat(" ", tc.sceneW))
	}

	grid := make([][]rune, tc.sceneRows)
	for i, l := range bgLines {
		if i >= tc.sceneRows {
			break
		}
		runes := []rune(l)
		for len(runes) < tc.sceneW {
			runes = append(runes, ' ')
		}
		if len(runes) > tc.sceneW {
			runes = runes[:tc.sceneW]
		}
		grid[i] = runes
	}

	frames := s.currentFrames()
	if len(frames) > 0 {
		frame := frames[s.frameIdx%len(frames)]
		frameLines := strings.Split(strings.TrimRight(frame, "\n"), "\n")

		for row := 0; row < tc.walkRowCount && row < len(frameLines); row++ {
			sceneRow := tc.walkRowStart + row
			if sceneRow >= tc.sceneRows {
				break
			}
			animalRunes := []rune(frameLines[row])
			for col := 0; col < len(animalRunes); col++ {
				sceneCol := s.x + col
				if sceneCol < 0 || sceneCol >= tc.sceneW {
					continue
				}
				ar := animalRunes[col]
				if !isBlank(ar) {
					grid[sceneRow][sceneCol] = ar
				}
			}
		}

		// Re-apply foreground to mask animal behind bushes.
		for i, l := range bgLines {
			if i >= tc.sceneRows || i < tc.walkRowStart {
				continue
			}
			bgRunes := []rune(l)
			for col := 0; col < len(bgRunes) && col < tc.sceneW; col++ {
				br := bgRunes[col]
				if !isBlank(br) {
					grid[i][col] = br
				}
			}
		}
	}

	out := make([]string, tc.sceneRows)
	for i, row := range grid {
		out[i] = string(row)
	}
	return out
}

// ── View ──────────────────────────────────────────────────────────────────────

func (s sceneAnim) view(w, h int) string {
	if w == 0 || h == 0 {
		return ""
	}
	tc := tierForSize(w, h)

	margin := (w - tc.sceneW) / 2
	if margin < 0 {
		margin = 0
	}
	indent := strings.Repeat(" ", margin)

	var b strings.Builder

	// Composited scene.
	for _, l := range s.composite(tc) {
		b.WriteString(indent)
		b.WriteString(l)
		b.WriteByte('\n')
	}

	// PLAY lines — trim trailing blanks.
	playLines := strings.Split(strings.TrimRight(tc.play, "\n"), "\n")
	for len(playLines) > 0 && strings.TrimSpace(playLines[len(playLines)-1]) == "" {
		playLines = playLines[:len(playLines)-1]
	}
	playRows := len(playLines)

	// Distribute remaining vertical space.
	remaining := h - tc.sceneRows - playRows - 2
	if remaining < 0 {
		remaining = 0
	}
	topPad := remaining / 2
	for i := 0; i < topPad; i++ {
		b.WriteByte('\n')
	}

	// Render each PLAY line with animation, centered on terminal width.
	rendered := strings.Split(s.play.render(playLines), "\n")
	for li, line := range rendered {
		// Center based on the raw (uncoloured) line width.
		rawLine := playLines[li]
		pad := (w - len([]rune(rawLine))) / 2
		if pad < 0 {
			pad = 0
		}
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(line)
		b.WriteByte('\n')
	}

	// Blank + hint centered.
	b.WriteByte('\n')
	hint := "[ enter to foster an animal ]"
	hintPad := (w - len(hint)) / 2
	if hintPad < 0 {
		hintPad = 0
	}
	b.WriteString(strings.Repeat(" ", hintPad))
	b.WriteString(hintStyle.Render(hint))

	// Clamp to h rows.
	out := b.String()
	lines := strings.Split(out, "\n")
	for len(lines) < h {
		lines = append(lines, "")
	}
	if len(lines) > h {
		lines = lines[:h]
	}
	return strings.Join(lines, "\n")
}
