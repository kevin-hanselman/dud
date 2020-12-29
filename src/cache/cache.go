package cache

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/strategy"
)

// A Cache provides a means to store Artifacts.
type Cache interface {
	Commit(workspaceDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error
	Checkout(workspaceDir string, art artifact.Artifact, strat strategy.CheckoutStrategy) error
	PathForChecksum(checksum string) (string, error)
	Status(workspaceDir string, art artifact.Artifact) (artifact.ArtifactWithStatus, error)
	Fetch(workspaceDir, remoteSrc string, art artifact.Artifact) error
	Push(workspaceDir, remoteDst string, art artifact.Artifact) error
}

// A LocalCache is a concrete Cache that uses a directory on a local filesystem.
type LocalCache struct {
	dir string
}

// NewLocalCache initializes a LocalCache with a valid cache directory.
func NewLocalCache(dir string) (*LocalCache, error) {
	if dir == "" {
		return nil, errors.New("cache directory path must be set")
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	return &LocalCache{dir: absDir}, nil
}

// Dir returns the root directory for the LocalCache.
func (ch *LocalCache) Dir() string {
	return ch.dir
}

// PathForChecksum returns the expected location of an object with the
// given checksum in the cache. If the checksum has an invalid (e.g. empty)
// checksum value, this function returns an error.
func (ch *LocalCache) PathForChecksum(checksum string) (string, error) {
	if len(checksum) < 3 {
		return "", fmt.Errorf("invalid checksum: %#v", checksum)
	}
	return filepath.Join(checksum[:2], checksum[2:]), nil
}

type directoryManifest struct {
	Path     string
	Contents map[string]*artifact.Artifact
}
