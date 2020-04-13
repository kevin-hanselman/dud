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
	CachePathForChecksum(checksum string) (string, error)
	Status(workingDir string, art artifact.Artifact) (artifact.Status, error)
}

// A LocalCache is a concrete Cache that uses a directory on a local filesystem.
type LocalCache struct {
	Dir string
}

// CachePathForChecksum returns the expected location of an object with the
// given checksum in the cache. If the checksum has an invalid (e.g. empty)
// checksum value, this function returns an error.
func (cache *LocalCache) CachePathForChecksum(checksum string) (string, error) {
	if len(checksum) < 3 {
		return "", fmt.Errorf("invalid checksum: %#v", checksum)
	}
	return path.Join(cache.Dir, checksum[:2], checksum[2:]), nil
}

type directoryManifest struct {
	Checksum string
	Path     string
	Contents []*artifact.Artifact
}

func (man *directoryManifest) GetChecksum() string {
	return man.Checksum
}

func (man *directoryManifest) SetChecksum(checksum string) {
	man.Checksum = checksum
}
