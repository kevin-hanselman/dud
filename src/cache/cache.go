package cache

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/mattn/go-isatty"
)

const (
	cacheFilePerms = 0o444

	// Template for progress report.
	//
	// rtime docs copied from cheggaaa/pb:
	// First string will be used as value for format time duration string, default is "%s".
	// Second string will be used when bar finished and value indicates elapsed time, default is "%s"
	// Third string will be used when value not available, default is "?"
	progressTemplate pb.ProgressBarTemplate = `  {{string . "prefix"}}  {{counters . }}` +
		`  {{percent . "%3.0f%%"}}  {{speed . "%s/s" "?/s"}}  {{rtime . "ETA %s" "%s total"}}`
)

// These are somewhat arbitrary numbers. We need to profile more.
var (
	// The number of concurrent workers available to a top-level directory
	// artifact and all its child artifacts.
	maxSharedWorkers = 64
	// The number of concurrent workers available to each individual directory
	// artifact (and not its children). Dedicated workers are necessary because
	// without them, deadlocks can occur when maxSharedWorkers is less than the
	// directory depth.
	maxDedicatedWorkers = 1
)

// A Cache provides a means to store Artifacts.
type Cache interface {
	Commit(
		workDir string,
		art *artifact.Artifact,
		s strategy.CheckoutStrategy,
		l *agglog.AggLogger,
	) error
	Checkout(
		workDir string,
		art artifact.Artifact,
		s strategy.CheckoutStrategy,
		p *pb.ProgressBar,
	) error
	Status(workDir string, art artifact.Artifact) (artifact.ArtifactWithStatus, error)
	Fetch(workDir, remoteSrc string, art artifact.Artifact) error
	Push(workDir, remoteDst string, art artifact.Artifact) error
}

// A LocalCache is a Cache that uses a directory on a local filesystem.
type LocalCache struct {
	dir string
}

// NewLocalCache initializes a LocalCache with a valid cache directory.
func NewLocalCache(dir string) (ch LocalCache, err error) {
	if dir == "" {
		return ch, errors.New("cache directory path must be set")
	}
	ch.dir, err = filepath.Abs(dir)
	if err != nil {
		return ch, err
	}
	return ch, nil
}

// PathForChecksum returns the expected location of an object with the
// given checksum in the cache. If the checksum has an invalid (e.g. empty)
// checksum value, this function returns an error.
func (ch LocalCache) PathForChecksum(checksum string) (string, error) {
	if len(checksum) < 3 {
		return "", InvalidChecksumError{checksum: checksum}
	}
	return filepath.Join(checksum[:2], checksum[2:]), nil
}

type directoryManifest struct {
	Path     string                        `json:"path,"`
	Contents map[string]*artifact.Artifact `json:"contents,"`
}

// InvalidChecksumError is an error case where a valid checksum was expected
// but not found.
type InvalidChecksumError struct {
	checksum string
}

func (err InvalidChecksumError) Error() string {
	if err.checksum == "" {
		return "no checksum"
	}
	return fmt.Sprintf("invalid checksum: %#v", err.checksum)
}

// MissingFromCacheError is an error case where a cache file was expected but
// not found.
type MissingFromCacheError struct {
	checksum string
}

func (err MissingFromCacheError) Error() string {
	return fmt.Sprintf("checksum missing from cache: %#v", err.checksum)
}

func newProgress(prefix string) (progress *pb.ProgressBar) {
	// Only show the progress report if stderr is a terminal. Otherwise, don't
	// bother updating the progress report and send any incidental output to
	// /dev/null. Either way we instantiate the progress tracker because we
	// still need it to tell us how many bytes we've read/written.
	progress = progressTemplate.New(0)
	if isatty.IsTerminal(os.Stderr.Fd()) {
		progress.SetRefreshRate(100 * time.Millisecond).SetWriter(os.Stderr)
		progress.SetMaxWidth(120).Set(pb.TimeRound, time.Millisecond)
		progress.Set("prefix", fmt.Sprintf("%-20s", prefix))
	} else {
		progress.SetRefreshRate(time.Hour).SetWriter(ioutil.Discard)
	}
	return
}
