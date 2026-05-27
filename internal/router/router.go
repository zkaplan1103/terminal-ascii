package router

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/zkaplan/terminal-site/internal/page"
	"github.com/zkaplan/terminal-site/internal/pages/adopt"
	"github.com/zkaplan/terminal-site/internal/pages/bio"
	"github.com/zkaplan/terminal-site/internal/pages/contact"
	"github.com/zkaplan/terminal-site/internal/pages/projects"
)

// usernameEntry maps the SSH login username to the page key the visitor lands
// on. Unknown usernames (and "nav"/"menu" aliases) fall through to "bio" —
// the bio page IS the home/menu now.
var usernameEntry = map[string]string{
	"bio":      "bio",
	"adopt":    "adopt",
	"projects": "projects",
	"contact":  "contact",
}

// pageFor builds a fresh tea.Model for the given page key. Returns bio as the
// fallback for any unknown key.
func pageFor(key string) tea.Model {
	switch key {
	case "adopt":
		return adopt.New()
	case "projects":
		return projects.New()
	case "contact":
		return contact.New()
	default:
		return bio.New()
	}
}

// EntryPageFor returns the page key for a given SSH username. The router
// constructor calls this to decide where to land the visitor.
func EntryPageFor(username string) string {
	if page, ok := usernameEntry[username]; ok {
		return page
	}
	return "bio"
}

// Model is the root tea.Model — wraps a current page, forwards messages, and
// handles global key bindings (quit) + NavigateMsg page swaps.
type Model struct {
	current     tea.Model
	currentKey  string
	width       int
	height      int
}

// New constructs a router rooted at the entry page for the given SSH username.
// The constructor primes the initial page with the current window size so it
// can render before any input arrives.
func New(username string, width, height int) Model {
	key := EntryPageFor(username)
	m := Model{
		currentKey: key,
		current:    pageFor(key),
		width:      width,
		height:     height,
	}
	m.current, _ = m.current.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return m
}

func (m Model) Init() tea.Cmd { return m.current.Init() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case page.NavigateMsg:
		m.currentKey = msg.To
		m.current = pageFor(msg.To)
		// Prime the new page with the current size so its first View call
		// has dimensions to work with.
		m.current, _ = m.current.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		return m, m.current.Init()
	}
	var cmd tea.Cmd
	m.current, cmd = m.current.Update(msg)
	return m, cmd
}

func (m Model) View() string { return m.current.View() }
