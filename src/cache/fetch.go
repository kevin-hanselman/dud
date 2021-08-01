package cache

import (
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/pkg/errors"
)

// Fetch downloads an Artifact from a remote location to the local cache.
func (ch LocalCache) Fetch(
	remoteSrc string,
	artifacts ...artifact.Artifact,
) error {
	toFetch := make(map[string]struct{})
	dirArtifacts := make(map[string]artifact.Artifact)
	for _, art := range artifacts {
		if art.SkipCache {
			continue
		}
		status, cachePath, _, err := checksumStatus(ch, art)
		if err != nil {
			return errors.Wrapf(err, "fetch %s", art.Path)
		}
		if !status.HasChecksum {
			return errors.Wrapf(
				InvalidChecksumError{art.Checksum},
				"fetch %s",
				art.Path,
			)
		}
		// Fetch an artifact if it's missing from the cache.
		if !status.ChecksumInCache {
			toFetch[cachePath] = struct{}{}
		}
		if art.IsDir {
			dirArtifacts[cachePath] = art
		}
	}

	// This length check could/should be handled in remoteCopy, but the tests
	// currently expect remoteCopy not to be called if there's nothing to
	// fetch.
	if len(toFetch) > 0 {
		if err := remoteCopy(remoteSrc, ch.dir, toFetch); err != nil {
			return errors.Wrap(err, "fetch")
		}
	}

	children := []artifact.Artifact{}
	// Collect all children of directory artifacts and recursively call Fetch
	// on all of them.
	for cachePath, dirArt := range dirArtifacts {
		man, err := readDirManifest(filepath.Join(ch.dir, cachePath))
		if err != nil {
			return errors.Wrapf(err, "fetch %s", dirArt.Path)
		}
		for _, art := range man.Contents {
			children = append(children, *art)
		}
	}
	if len(children) == 0 {
		return nil
	}
	// Don't wrap any error here because we're recursing.
	return ch.Fetch(remoteSrc, children...)
}
