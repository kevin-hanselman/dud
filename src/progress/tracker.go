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

func (s *state) started() bool {
	return !s.start.IsZero()
}

func (s *state) percent() float64 {
	if s.total > 0 {
		return float64(s.current) / float64(s.total) * 100
	}
	return 0
}

func (s *state) perSecond() float64 {
	if s.started() {
		return float64(s.current) / time.Since(s.start).Seconds()
	}
	return 0
}

func (s *state) eta() time.Duration {
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

func (s *state) add(n int64) {
	if !s.started() {
		s.start = time.Now()
	}
	s.current += n
}

func (s *state) addTotal(n int64) {
	s.total += n
}

type lockedStateWriter struct {
	s    *state
	lock *sync.Mutex
}

func (w *lockedStateWriter) Write(b []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	n := len(b)
	w.s.add(int64(n))
	return n, nil
}

type ProgressTracker struct {
	bytes *state
	files *state
	lock  sync.Mutex
}

func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		bytes: &state{},
		files: &state{},
	}
}

// AddFileBytes adds one file and N bytes to the progress tracker.
// If N is zero or negative, the bytes progress is not updated; such values
// indicate a file that's not being read (such as when checking artifact status
// or gathering cache files to push or fetch).
func (pt *ProgressTracker) AddFileBytes(numBytes int64) {
	pt.lock.Lock()
	defer pt.lock.Unlock()
	if numBytes > 0 {
		pt.bytes.addTotal(numBytes)
	}
	pt.files.addTotal(1)
}

func (pt *ProgressTracker) NewProxyReader(r io.Reader) io.Reader {
	pr := &lockedStateWriter{pt.bytes, &pt.lock}
	return io.TeeReader(r, pr)
}

func (pt *ProgressTracker) FileDone() {
	pt.lock.Lock()
	defer pt.lock.Unlock()
	pt.files.add(1)
}

func (pt *ProgressTracker) states() (bytes, files *state) {
	pt.lock.Lock()
	defer pt.lock.Unlock()
	return pt.bytes, pt.files
}
