package progress

import (
	"fmt"
	"strings"
	"text/tabwriter"
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
		refresh:    300 * time.Millisecond,
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
	var (
		out                strings.Builder
		durFmt             string
		durBytes, durFiles time.Duration
	)
	writer := tabwriter.NewWriter(&out, 0, 8, 0, '\t', 0)

	bytes, files := m.tracker.states()
	s := bytes
	if m.complete {
		durFmt = "%s total\n"
		durBytes = time.Since(s.start)
	} else {
		durFmt = "ETA %s\n"
		durBytes = s.eta()
	}
	rate, units := humanize.ComputeSI(s.perSecond())
	fmt.Fprintf(
		writer,
		"bytes:\t%s / %s\t%.2f %sB/s\t%3.0f%%\n",
		humanize.Bytes(uint64(s.current)),
		humanize.Bytes(uint64(s.total)),
		rate,
		units,
		s.percent(),
	)

	s = files
	if m.complete {
		durFiles = time.Since(s.start)
	} else {
		durFiles = s.eta()
	}
	rate, units = humanize.ComputeSI(s.perSecond())
	fmt.Fprintf(
		writer,
		"files:\t%s / %s\t%.2f%s files/s\t%3.0f%%\n",
		humanize.Comma(s.current),
		humanize.Comma(s.total),
		rate,
		units,
		s.percent(),
	)
	dur := max(durFiles, durBytes).Round(time.Millisecond)
	fmt.Fprintf(writer, durFmt, dur)
	writer.Flush()
	return out.String()
}
