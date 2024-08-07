package progress

import (
	"io"
	"os"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/mattn/go-isatty"
)

const (
	filesTemplate pb.ProgressBarTemplate = `{{string . "prefix"}}files: {{counters .}}  {{percent . "%3.0f%%"}}` +
		`  {{speed . "%s files/s" ""}}`

	// rtime docs copied from cheggaaa/pb:
	// First string will be used as value for format time duration string, default is "%s".
	// Second string will be used when bar finished and value indicates elapsed time, default is "%s"
	// Third string will be used when value not available, default is "?"
	bytesTemplate pb.ProgressBarTemplate = `{{string . "prefix"}}bytes: {{counters .}}  {{percent . "%3.0f%%"}}` +
		`  {{speed . "%s/s" ""}}  {{rtime . "ETA: %s" "Elapsed: %s" ""}}`
)

type Progress interface {
	// AddBytes adds to the progress report's total bytes.
	AddBytes(count int64)

	// AddFiles adds to the progress report's total files.
	AddFiles(count int)

	// ProxyReader returns a io.Reader that updates the bytes progress report when
	// read.
	ProxyReader(r io.Reader) io.Reader

	// DoneFile marks one file as completed.
	DoneFile()

	// Start starts the progress report.
	Start() error

	// Finish ends the progress report.
	Finish() error

	// CurrentFiles reports the current number of files process.
	CurrentFiles() int64

	// TotalFiles reports the current total of files to process.
	TotalFiles() int64
}

type FilesProgress struct {
	files *pb.ProgressBar
}

func (p *FilesProgress) AddBytes(count int64) {}

func (p *FilesProgress) AddFiles(count int) {
	p.files.AddTotal(int64(count))
}

func (p *FilesProgress) ProxyReader(r io.Reader) io.Reader {
	return r
}

func (p *FilesProgress) DoneFile() {
	p.files.Increment()
}

func (p *FilesProgress) Start() error {
	p.files.Start()
	return nil
}

func (p *FilesProgress) Finish() error {
	p.files.Finish()
	return nil
}

func (p *FilesProgress) CurrentFiles() int64 {
	return p.files.Current()
}

func (p *FilesProgress) TotalFiles() int64 {
	return p.files.Total()
}

type FilesBytesProgress struct {
	FilesProgress
	bytes *pb.ProgressBar
	pool  *pb.Pool
}

func (p *FilesBytesProgress) AddBytes(count int64) {
	p.bytes.AddTotal(count)
}

func (p *FilesBytesProgress) ProxyReader(r io.Reader) io.Reader {
	return p.bytes.NewProxyReader(r)
}

func (p *FilesBytesProgress) Start() error {
	p.initPool()
	return p.pool.Start()
}

func (p *FilesBytesProgress) Finish() error {
	// Both ProgressBar.Finish and Pool.Stop should be called for
	// correct final display of the progress report.
	p.files.Finish()
	p.bytes.Finish()
	return p.pool.Stop()
}

func (p *FilesBytesProgress) initPool() {
	// NewPool starts the ProgressBars, so we delay calling this until
	// Progress.Start().
	pool := pb.NewPool(p.files, p.bytes)
	p.pool = pool
}

func NewProgress(includeBytes bool, prefix string) Progress {
	// Only show the progress report if stderr is a terminal. Otherwise, don't
	// bother updating the progress report and send any incidental output to
	// /dev/null. Either way we instantiate the progress tracker because we
	// still need it to tell us how many files processed.
	if !isatty.IsTerminal(os.Stderr.Fd()) {
		return NewHiddenProgress()
	}

	filesProgress := &FilesProgress{
		files: newProgressBar(filesTemplate, prefix),
	}

	if !includeBytes {
		filesProgress.files.Set(pb.CleanOnFinish, false)
		return filesProgress
	}

	return &FilesBytesProgress{
		FilesProgress: *filesProgress,
		bytes:         newProgressBar(bytesTemplate, prefix),
	}
}

func NewHiddenProgress() Progress {
	return &FilesProgress{
		files: newHiddenProgressBar(),
	}
}

func newProgressBar(template pb.ProgressBarTemplate, prefix string) (p *pb.ProgressBar) {
	p = template.New(0)
	p.SetRefreshRate(100 * time.Millisecond)
	p.SetWriter(os.Stderr)
	p.SetMaxWidth(120)
	p.Set(pb.TimeRound, time.Millisecond)
	p.Set("prefix", prefix)
	// ProgressBars in a Pool require CleanOnFinish to be true
	// to avoid duplicated display when both ProgressBar.Finish
	// and Pool.Stop are called.
	p.Set(pb.CleanOnFinish, true)
	return p
}

func newHiddenProgressBar() (p *pb.ProgressBar) {
	p = pb.New(0)
	p.SetRefreshRate(time.Hour)
	p.SetWriter(io.Discard)
	return p
}
