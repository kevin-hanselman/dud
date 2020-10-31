package cache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/checksum"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/pkg/errors"
)

// Checkout finds the artifact in the cache and adds a copy of/link to said
// artifact in the working directory.
func (cache *LocalCache) Checkout(
	workspaceDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	if art.SkipCache {
		return nil
	}
	if art.IsDir {
		return checkoutDir(cache, workspaceDir, art, strat)
	}
	return checkoutFile(cache, workspaceDir, art, strat)
}

// InvalidChecksumError is an error case where a valid checksum was expected
// but not found.
type InvalidChecksumError struct {
	checksum string
}

func (err InvalidChecksumError) Error() string {
	return fmt.Sprintf("invalid checksum: %#v", err.checksum)
}

// MissingFromCacheError is an error case where a cache file was expected but
// not found.
type MissingFromCacheError struct {
	checksum string
}

func (err MissingFromCacheError) Error() string {
	return fmt.Sprintf("file missing from cache: %#v", err.checksum)
}

func checkoutFile(
	ch *LocalCache,
	workspaceDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	status, cachePath, workPath, err := quickStatus(ch, workspaceDir, *art)
	errorPrefix := fmt.Sprintf("checkout %#v", workPath)
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	cachePath = filepath.Join(ch.dir, cachePath)
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
		if err := os.Symlink(cachePath, workPath); err != nil {
			return errors.Wrap(err, errorPrefix)
		}
	}
	return nil
}

func checkoutDir(
	ch *LocalCache,
	workspaceDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	status, cachePath, workPath, err := quickStatus(ch, workspaceDir, *art)
	if err != nil {
		return err
	}
	cachePath = filepath.Join(ch.dir, cachePath)
	if !status.HasChecksum {
		return fmt.Errorf("checkoutDir: artifact has invalid checksum: %v", art.Checksum)
	}
	if !status.ChecksumInCache {
		return fmt.Errorf("checkoutDir: checksum %v not found in cache", art.Checksum)
	}
	if !(status.WorkspaceFileStatus == fsutil.StatusAbsent || status.WorkspaceFileStatus == fsutil.StatusDirectory) {
		return fmt.Errorf("checkoutDir: expected target to be empty or a directory, found %s", status.WorkspaceFileStatus)
	}
	man, err := readDirManifest(cachePath)
	if err != nil {
		return err
	}
	for _, childArt := range man.Contents {
		if err := ch.Checkout(workPath, childArt, strat); err != nil {
			return err
		}
	}
	return nil
}
