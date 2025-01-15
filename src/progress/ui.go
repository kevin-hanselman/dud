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
	complete   bool
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
		m.complete = true
		return m, tea.Quit
	}
	return m, m.tick()
}

func (m *model) View() string {
	// TODO: DRY this out a bit
	// TODO: Consider only showing one ETA/elapsed time--maybe an average
	s := m.tracker.bytes.state()
	var timeDisplay string
	if m.complete {
		timeDisplay = fmt.Sprintf("%s elapsed", time.Since(s.start))
	} else {
		timeDisplay = fmt.Sprintf("ETA %s", s.eta())
	}
	rate, units := humanize.ComputeSI(s.perSecond())
	bytesProgress := fmt.Sprintf(
		"bytes: %s / %s  %.2f %sB/s  %3.0f%%  %s\n",
		humanize.Bytes(uint64(s.current)),
		humanize.Bytes(uint64(s.total)),
		rate,
		units,
		s.percent(),
		timeDisplay,
	)
	s = m.tracker.files.state()
	if m.complete {
		timeDisplay = fmt.Sprintf("%s elapsed", time.Since(s.start))
	} else {
		timeDisplay = fmt.Sprintf("ETA %s", s.eta())
	}
	rate, units = humanize.ComputeSI(s.perSecond())
	filesProgress := fmt.Sprintf(
		"files: %s / %s  %.2f%s files/s  %3.0f%%  %s\n",
		humanize.Comma(s.current),
		humanize.Comma(s.total),
		rate,
		units,
		s.percent(),
		timeDisplay,
	)
	return bytesProgress + filesProgress
}
