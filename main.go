package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"github.com/zkaplan/terminal-site/internal/pages/bio"
)

const (
	host    = "localhost"
	port    = "2222"
	hostKey = ".ssh/host_ed25519"
)

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(hostKey),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("could not start server", "error", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("could not stop server", "error", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	m := rootModel{
		page:   bio.New(),
		width:  pty.Window.Width,
		height: pty.Window.Height,
	}
	// Prime the page with its first size message so it can render
	// before any user input arrives.
	m.page, _ = m.page.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

// rootModel is a thin wrapper that handles global keys (quit) and
// forwards everything else to the active page. Phase 2 will replace
// this with a real router that swaps pages based on SSH username.
type rootModel struct {
	page   tea.Model
	width  int
	height int
}

func (m rootModel) Init() tea.Cmd { return m.page.Init() }

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.page, cmd = m.page.Update(msg)
	return m, cmd
}

func (m rootModel) View() string { return m.page.View() }
