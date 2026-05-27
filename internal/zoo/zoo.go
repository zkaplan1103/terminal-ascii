package zoo

import "time"

// State represents what an animal is currently doing.
type State int

const (
	StateIdle State = iota
	StateEating
	StateWalking
	StateSleeping
)

func (s State) String() string {
	switch s {
	case StateEating:
		return "eating"
	case StateWalking:
		return "walking"
	case StateSleeping:
		return "sleeping"
	default:
		return "idle"
	}
}

// Animal holds species data and its animation frames per state.
// Frames are pre-baked ASCII strings (one per animation frame).
// A missing state falls back to StateIdle frames.
type Animal struct {
	Name    string // proper name e.g. "Biscuit"
	Species string // species key e.g. "dog"
	Breed   string // variant key e.g. "beagle"
	Frames  map[State][]string
	Delays  map[State][]time.Duration
}

// FramesFor returns the frames for a given state, falling back to idle.
func (a *Animal) FramesFor(s State) ([]string, []time.Duration) {
	if f, ok := a.Frames[s]; ok && len(f) > 0 {
		return f, a.Delays[s]
	}
	return a.Frames[StateIdle], a.Delays[StateIdle]
}

// Session is the live per-visitor state. No persistence — dies with the SSH session.
type Session struct {
	Animal    *Animal
	State     State
	Hunger    int // 0..100, ticks up over time
	Energy    int // 0..100, ticks down over time
	Happiness int // 0..100, decays slowly

	// actionUntil: if non-zero, return to idle after this time.
	ActionUntil time.Time
}

// NewSession creates a fresh session with a chosen animal.
func NewSession(a *Animal) *Session {
	return &Session{
		Animal:    a,
		State:     StateIdle,
		Hunger:    20,
		Energy:    80,
		Happiness: 70,
	}
}

// Tick advances the session stats by one second. Returns true if any stat changed.
func (s *Session) Tick(now time.Time) bool {
	changed := false

	// Return to idle after a timed action.
	if !s.ActionUntil.IsZero() && now.After(s.ActionUntil) {
		s.ActionUntil = time.Time{}
		s.State = StateIdle
		changed = true
	}

	// Stats change every N ticks — tracked by wall time.
	// Hunger: +1 per 60s
	// Energy: -1 per 60s (awake), +5 per 10s (sleeping)
	// Happiness: -1 per 120s
	// We approximate by checking seconds modulo intervals.
	sec := now.Unix()
	if sec%60 == 0 {
		if s.Hunger < 100 {
			s.Hunger++
			changed = true
		}
		if s.State != StateSleeping && s.Energy > 0 {
			s.Energy--
			changed = true
		}
	}
	if sec%10 == 0 && s.State == StateSleeping && s.Energy < 100 {
		s.Energy += 5
		if s.Energy > 100 {
			s.Energy = 100
		}
		changed = true
	}
	if sec%120 == 0 && s.Happiness > 0 {
		s.Happiness--
		changed = true
	}

	return changed
}

// Feed reduces hunger and briefly triggers eating state.
func (s *Session) Feed(now time.Time) {
	s.Hunger -= 30
	if s.Hunger < 0 {
		s.Hunger = 0
	}
	s.Happiness += 5
	if s.Happiness > 100 {
		s.Happiness = 100
	}
	s.State = StateEating
	s.ActionUntil = now.Add(3 * time.Second)
}

// Walk reduces energy and briefly triggers walking state.
func (s *Session) Walk(now time.Time) {
	s.Energy -= 15
	if s.Energy < 0 {
		s.Energy = 0
	}
	s.Happiness += 10
	if s.Happiness > 100 {
		s.Happiness = 100
	}
	s.State = StateWalking
	s.ActionUntil = now.Add(4 * time.Second)
}

// Sleep sets energy to full and enters sleeping state (no auto-return).
func (s *Session) Sleep() {
	s.Energy = 100
	s.State = StateSleeping
	s.ActionUntil = time.Time{} // manual exit only
}

// Pet increases happiness without changing state.
func (s *Session) Pet() {
	s.Happiness += 5
	if s.Happiness > 100 {
		s.Happiness = 100
	}
}

// WakeUp returns to idle from sleep.
func (s *Session) WakeUp() {
	if s.State == StateSleeping {
		s.State = StateIdle
		s.ActionUntil = time.Time{}
	}
}
