package progress

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

type commandExitMsg struct {
	err error
}

type model struct {
	command    tea.Cmd
	refresh    time.Duration
	tracker    *ProgressTracker
	finalError error
}

func newProgressUI(tracker *ProgressTracker, f func() error) *model {
	cmd := func() tea.Msg { return commandExitMsg{f()} }
	return &model{
		command:    cmd,
		refresh:    100 * time.Millisecond,
		tracker:    tracker,
		finalError: nil,
	}
}

type UserInterrupt struct{}

func (u UserInterrupt) Error() string {
	return "encountered user interrupt"
}

type tickMsg struct{}

func (m *model) tick() tea.Cmd {
	return tea.Every(m.refresh, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.tick(), m.command)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// If a key is pressed, ignore it unless it's a dedicated
	// "quit" key.
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.finalError = UserInterrupt{}
			return m, tea.Quit
		}
	case commandExitMsg:
		m.finalError = msg.err
		return m, tea.Quit
	}
	return m, m.tick()
}

func (m *model) View() string {
	s := m.tracker.bytes.State()
	bytesProgress := fmt.Sprintf(
		"bytes: %s / %s  %s  %3.0f%%\n",
		humanize.Bytes(uint64(s.CurrentBytes)),
		humanize.Bytes(uint64(s.Max)),
		humanize.SI(s.KBsPerSecond*1024, "B/s"),
		s.CurrentPercent*100,
	)
	s = m.tracker.files.State()
	filesProgress := fmt.Sprintf(
		"files: %s / %s  %s %3.0f%%\n",
		humanize.Comma(s.CurrentNum),
		humanize.Comma(s.Max),
		humanize.SI(s.KBsPerSecond*1024, " files/s"),
		s.CurrentPercent*100,
	)
	return bytesProgress + filesProgress
}
