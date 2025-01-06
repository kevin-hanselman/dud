package progress

import (
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func Run(tracker *ProgressTracker, f func() error) error {
	ui := newProgressUI(tracker, f)

	outFile := os.Stderr
	var outWriter io.Writer = outFile
	if !term.IsTerminal(int(outFile.Fd())) {
		outWriter = io.Discard
	}

	prog := tea.NewProgram(ui, tea.WithOutput(outWriter))
	if _, err := prog.Run(); err != nil {
		return err
	}

	return ui.finalError
}
