package progress

import (
	"io"

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

// AddFileBytes adds one file and N bytes to the progress tracker.
// If N is zero or negative, the bytes progress is not updated; such values
// indicate a file that's not being read (such as when checking artifact status
// or gathering cache files to push or fetch).
func (pt *ProgressTracker) AddFileBytes(numBytes int64) {
	if numBytes > 0 {
		pt.bytes.AddMax64(numBytes)
	}
	pt.files.AddMax(1)
}

func (pt *ProgressTracker) NewProxyReader(r io.Reader) io.Reader {
	return io.TeeReader(r, pt.bytes)
}

func (pt *ProgressTracker) FileDone() {
	pt.files.Add(1)
}
