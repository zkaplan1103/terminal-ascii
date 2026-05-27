package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zkaplan/terminal-site/internal/zoo"
)

// AnimalTickMsg fires on each animation frame.
// Exported so adopt page can type-switch on it when needed.
type AnimalTickMsg struct{}

// AnimalView is the animated animal component. It owns its own tick loop
// independent of the zoo stats tick — animation can run at 2-12 fps while
// stats only update every second.
type AnimalView struct {
	session *zoo.Session
	frame   int
	frames  []string
	delays  []time.Duration
}

// NewAnimalView creates an AnimalView for the given session.
func NewAnimalView(s *zoo.Session) AnimalView {
	frames, delays := s.Animal.FramesFor(s.State)
	return AnimalView{
		session: s,
		frame:   0,
		frames:  frames,
		delays:  delays,
	}
}

// Init starts the animation tick.
func (a AnimalView) Init() tea.Cmd {
	return a.scheduleTick()
}

func (a AnimalView) scheduleTick() tea.Cmd {
	if len(a.delays) == 0 {
		return tea.Tick(400*time.Millisecond, func(time.Time) tea.Msg { return AnimalTickMsg{} })
	}
	d := a.delays[a.frame%len(a.delays)]
	return tea.Tick(d, func(time.Time) tea.Msg { return AnimalTickMsg{} })
}

// Update handles animation ticks and state changes from the parent.
func (a AnimalView) Update(msg tea.Msg) (AnimalView, tea.Cmd) {
	switch msg.(type) {
	case AnimalTickMsg:
		// If the session state changed, swap to the new frame set.
		frames, delays := a.session.Animal.FramesFor(a.session.State)
		if &frames[0] != &a.frames[0] {
			a.frames = frames
			a.delays = delays
			a.frame = 0
		} else {
			a.frame = (a.frame + 1) % len(a.frames)
		}
		return a, a.scheduleTick()
	}
	return a, nil
}

// View returns the current animation frame string.
func (a AnimalView) View() string {
	if len(a.frames) == 0 {
		return ""
	}
	return a.frames[a.frame%len(a.frames)]
}
