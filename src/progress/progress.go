package progress

import (
	"fmt"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
	bar "github.com/schollz/progressbar/v3"
)

type ProgressTracker struct {
	bytes *bar.ProgressBar
	files *bar.ProgressBar
}

func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		bytes: bar.DefaultBytesSilent(0),
		files: bar.DefaultSilent(0),
	}
}

func (ui *ProgressTracker) AddFileBytes(numBytes int64) {
	if numBytes > 0 {
		ui.bytes.AddMax64(numBytes)
	}
	ui.files.AddMax(1)
}

func (ui *ProgressTracker) NewProxyReader(r io.Reader) io.Reader {
	return io.TeeReader(r, ui.bytes)
}

func (ui *ProgressTracker) FileDone() {
	ui.files.Add(1)
}

type commandExitMsg struct {
	err error
}

type ProgressUI struct {
	command tea.Cmd
	refresh time.Duration
	tracker *ProgressTracker
	Error   error
}

func NewProgressUI(tracker *ProgressTracker, f func() error) *ProgressUI {
	cmd := func() tea.Msg { return commandExitMsg{f()} }
	return &ProgressUI{
		command: cmd,
		refresh: 100 * time.Millisecond,
		tracker: tracker,
		Error:   nil,
	}
}

type UserInterrupt struct{}

func (u UserInterrupt) Error() string {
	return "encountered user interrupt"
}

type tickMsg struct{}

func (ui *ProgressUI) tick() tea.Cmd {
	return tea.Every(ui.refresh, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (ui *ProgressUI) Init() tea.Cmd {
	return tea.Batch(ui.tick(), ui.command)
}

func (ui *ProgressUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// If a key is pressed, ignore it unless it's a dedicated
	// "quit" key.
	case tea.KeyMsg:
		tea.Printf("key pressed: %v", msg.String())
		if msg.Type == tea.KeyCtrlC {
			ui.Error = UserInterrupt{}
			return ui, tea.Quit
		}
	case commandExitMsg:
		tea.Println("command exit message found")
		ui.Error = msg.err
		return ui, tea.Quit
	}
	return ui, ui.tick()
}

func (ui *ProgressUI) View() string {
	s := ui.tracker.bytes.State()
	bytesProgress := fmt.Sprintf(
		"bytes: %s / %s  %s/s  %3.0f%%\n",
		humanize.Bytes(uint64(s.CurrentBytes)),
		humanize.Bytes(uint64(s.Max)),
		humanize.SI(s.KBsPerSecond*1024, "B"),
		s.CurrentPercent*100,
	)
	s = ui.tracker.files.State()
	filesProgress := fmt.Sprintf(
		"files: %s / %s  %s file/s %3.0f%%\n",
		humanize.Comma(int64(s.CurrentBytes)),
		humanize.Comma(s.Max),
		humanize.SI(s.KBsPerSecond*1024, ""),
		s.CurrentPercent*100,
	)
	return bytesProgress + filesProgress
}
