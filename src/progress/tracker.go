package progress

import (
	"io"
	"sync"
	"time"
)

type state struct {
	current int64
	total   int64
	start   time.Time
}

func (s state) started() bool {
	return !s.start.IsZero()
}

func (s state) percent() float64 {
	if s.total > 0 {
		return float64(s.current) / float64(s.total) * 100
	}
	return 0
}

func (s state) perSecond() float64 {
	if s.started() {
		return float64(s.current) / time.Since(s.start).Seconds()
	}
	return 0
}

func (s state) eta() time.Duration {
	// current        elapsed
	// ------- = -------------------
	//  total    elapsed + remaining
	//
	//                          total
	// remaining = elapsed * ( ------- - 1 )
	//                         current
	// https://www.wolframalpha.com/input?i=c+%2F+t+%3D+e+%2F+%28e+%2B+r%29%2C+solve+for+r
	if s.started() {
		secondsLeft := time.Since(s.start).Seconds() * (float64(s.total)/float64(s.current) - 1)
		return time.Duration(secondsLeft * float64(time.Second))
	}
	return 0
}

type progress struct {
	lock sync.Mutex
	s    state
}

func (p *progress) add(n int64) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.s.started() {
		p.s.start = time.Now()
	}
	p.s.current += n
}

func (p *progress) addTotal(n int64) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.s.total += n
}

func (p *progress) Write(b []byte) (int, error) {
	n := len(b)
	p.add(int64(n))
	return n, nil
}

func (p *progress) state() state {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.s
}

type ProgressTracker struct {
	bytes *progress
	files *progress
}

func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		bytes: &progress{},
		files: &progress{},
	}
}

// AddFileBytes adds one file and N bytes to the progress tracker.
// If N is zero or negative, the bytes progress is not updated; such values
// indicate a file that's not being read (such as when checking artifact status
// or gathering cache files to push or fetch).
func (pt *ProgressTracker) AddFileBytes(numBytes int64) {
	if numBytes > 0 {
		pt.bytes.addTotal(numBytes)
	}
	pt.files.addTotal(1)
}

func (pt *ProgressTracker) NewProxyReader(r io.Reader) io.Reader {
	return io.TeeReader(r, pt.bytes)
}

func (pt *ProgressTracker) FileDone() {
	pt.files.add(1)
}
