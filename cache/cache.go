package cache

import (
	"fmt"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/strategy"
	"path"
)

// A Cache provides a means to store Artifacts.
type Cache interface {
	Commit(workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error
	Checkout(workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error
	CachePathForArtifact(art artifact.Artifact) (string, error)
	Status(workingDir string, art artifact.Artifact) (artifact.Status, error)
}

// A LocalCache is a concrete Cache that uses a directory on a local filesystem.
type LocalCache struct {
	Dir string
}

// CachePathForArtifact returns the expected location of the given artifact in the cache.
// If the artifact has an invalid (e.g. empty) checksum value, this function returns an error.
func (cache *LocalCache) CachePathForArtifact(art artifact.Artifact) (string, error) {
	if len(art.Checksum) < 3 {
		return "", fmt.Errorf("invalid checksum: %#v", art.Checksum)
	}
	return path.Join(cache.Dir, art.Checksum[:2], art.Checksum[2:]), nil
}
