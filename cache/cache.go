package cache

import (
	"fmt"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/strategy"
	"path/filepath"
)

// A Cache provides a means to store Artifacts.
type Cache interface {
	Commit(workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error
	Checkout(workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error
	PathForChecksum(checksum string) (string, error)
	Status(workingDir string, art artifact.Artifact) (artifact.Status, error)
}

// A LocalCache is a concrete Cache that uses a directory on a local filesystem.
type LocalCache struct {
	dir string
}

// NewLocalCache initializes a LocalCache with a valid cache directory.
func NewLocalCache(dir string) (*LocalCache, error) {
	if !filepath.IsAbs(dir) {
		return nil, fmt.Errorf("NewLocalCache: %#v not an absolute path", dir)
	}
	return &LocalCache{dir: dir}, nil
}

// Dir returns the root directory for the LocalCache.
func (ch *LocalCache) Dir() string {
	return ch.dir
}

// Equal returns true if the two LocalCaches are equivalent.
// This is necessary for go-cmp to properly compare LocalCache objects (e.g. in
// file commit args).  TODO: Needing to compare LocalCaches in tests is a bit
// smelly. Consider removing this method?
func (ch *LocalCache) Equal(other *LocalCache) bool {
	return ch.Dir() == other.Dir()
}

// PathForChecksum returns the expected location of an object with the
// given checksum in the cache. If the checksum has an invalid (e.g. empty)
// checksum value, this function returns an error.
func (ch *LocalCache) PathForChecksum(checksum string) (string, error) {
	if len(checksum) < 3 {
		return "", fmt.Errorf("invalid checksum: %#v", checksum)
	}
	return filepath.Join(ch.dir, checksum[:2], checksum[2:]), nil
}

type directoryManifest struct {
	Path     string
	Contents []*artifact.Artifact
}
