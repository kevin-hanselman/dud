package cache

import (
	"fmt"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/checksum"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
)

// Checkout finds the artifact in the cache and adds a copy of/link to said
// artifact in the working directory.
func (cache *LocalCache) Checkout(
	workingDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	if art.IsDir {
		return checkoutDir(cache, workingDir, art, strat)
	}
	return checkoutFile(cache, workingDir, art, strat)
}

// InvalidChecksumError represents a error case where a valid checksum was expected
// but not found.
type InvalidChecksumError struct {
	invalidChecksum string
}

func (err InvalidChecksumError) Error() string {
	return fmt.Sprintf("invalid checksum: %#v", err.invalidChecksum)
}

// MissingFromCacheError represents an error case where the cache file was
// expected but not found.
type MissingFromCacheError struct {
	checksum string
}

func (err MissingFromCacheError) Error() string {
	return fmt.Sprintf("file missing from cache: %#v", err.checksum)
}

var checkoutFile = func(
	ch *LocalCache,
	workingDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	status, cachePath, workPath, err := quickStatus(ch, workingDir, *art)
	errorPrefix := fmt.Sprintf("checkout %#v", workPath)
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	if art.SkipCache {
		return nil
	}
	if !status.HasChecksum {
		return errors.Wrap(InvalidChecksumError{art.Checksum}, errorPrefix)
	}
	if !status.ChecksumInCache {
		return errors.Wrap(MissingFromCacheError{art.Checksum}, errorPrefix)
	}
	if err := os.MkdirAll(filepath.Dir(workPath), 0755); err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	switch strat {
	case strategy.CopyStrategy:
		srcFile, err := os.Open(cachePath)
		if err != nil {
			return errors.Wrap(err, errorPrefix)
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(workPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return errors.Wrap(err, errorPrefix)
		}
		defer dstFile.Close()

		checksum, err := checksum.Checksum(io.TeeReader(srcFile, dstFile), 0)
		if err != nil {
			return errors.Wrap(err, errorPrefix)
		}
		if checksum != art.Checksum {
			return fmt.Errorf("%s: found checksum %#v, expected %#v", errorPrefix, checksum, art.Checksum)
		}
	case strategy.LinkStrategy:
		// TODO: hardlink when possible?
		if err := os.Symlink(cachePath, workPath); err != nil {
			return errors.Wrap(err, errorPrefix)
		}
	}
	return nil
}

func checkoutDir(
	ch *LocalCache,
	workingDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	status, cachePath, workPath, err := quickStatus(ch, workingDir, *art)
	if err != nil {
		return errors.Wrap(err, "checkoutDir")
	}
	if !status.HasChecksum {
		return fmt.Errorf("checkoutDir: artifact has invalid checksum: %v", art.Checksum)
	}
	if !status.ChecksumInCache {
		return fmt.Errorf("checkoutDir: checksum %v not found in cache", art.Checksum)
	}
	if !(status.WorkspaceFileStatus == fsutil.Absent || status.WorkspaceFileStatus == fsutil.Directory) {
		return fmt.Errorf("checkoutDir: expected target to be empty or a directory, found %s", status.WorkspaceFileStatus)
	}
	man, err := readDirManifest(cachePath)
	for _, childArt := range man.Contents {
		if err := ch.Checkout(workPath, childArt, strat); err != nil {
			// TODO: undo previous file checkouts?
			return errors.Wrap(err, "checkoutDir")
		}
	}
	return nil
}
