package cache

import (
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/pkg/errors"
)

// Fetch downloads an Artifact from a remote location to the local cache.
//
// This uses a map of Artifacts instead of a slice to ease both testing and
// calling code. Primarily, a Stage's outputs will be passed to this function,
// so it's convenient to pass stage.Outputs directly. This also eases testing,
// because transcribing the map into a slice would introduce non-determinism.
func (ch LocalCache) Fetch(
	remoteSrc string,
	artifacts map[string]*artifact.Artifact,
) error {
	fetchFiles := make(map[string]struct{})
	dirArtifacts := make(map[string]*artifact.Artifact)
	// It's important not to use/assume what the string key in 'artifacts'
	// represents. Before recursing below, we change the keys to checksums to
	// prevent Artifacts with the same relative path from clobbering each
	// other.
	for _, art := range artifacts {
		if art.SkipCache {
			continue
		}
		status, cachePath, _, err := checksumStatus(ch, *art)
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
			fetchFiles[cachePath] = struct{}{}
		}
		if art.IsDir {
			dirArtifacts[cachePath] = art
		}
	}

	// This length check could/should be handled in remoteCopy, but the tests
	// currently expect remoteCopy not to be called if there's nothing to
	// fetch.
	if len(fetchFiles) > 0 {
		if err := remoteCopy(remoteSrc, ch.dir, fetchFiles); err != nil {
			return errors.Wrap(err, "fetch")
		}
	}

	children := make(map[string]*artifact.Artifact)
	// Collect all children of directory artifacts and call Fetch
	// on all of them at once.
	for cachePath, dirArt := range dirArtifacts {
		man, err := readDirManifest(filepath.Join(ch.dir, cachePath))
		if err != nil {
			return errors.Wrapf(err, "fetch %s", dirArt.Path)
		}
		for _, art := range man.Contents {
			// Use the Artifact's checksum as a key to ensure Artifacts with
			// the same relative path (from different parent Artifacts) don't
			// clobber each other.
			children[art.Checksum] = art
		}
	}
	if len(children) == 0 {
		return nil
	}
	// Don't wrap any error here because we're recursing.
	return ch.Fetch(remoteSrc, children)
}
