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
func (cache LocalCache) Checkout(
	workspaceDir string,
	art artifact.Artifact,
	strat strategy.CheckoutStrategy,
) (err error) {
	if art.SkipCache {
		return
	}
	if art.IsDir {
		err = checkoutDir(cache, workspaceDir, art, strat)
	} else {
		err = checkoutFile(cache, workspaceDir, art, strat)
	}
	return errors.Wrapf(err, "checkout %s", art.Path)
}

func checkoutFile(
	ch LocalCache,
	workspaceDir string,
	art artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	status, cachePath, workPath, err := quickStatus(ch, workspaceDir, art)
	if err != nil {
		return err
	}
	if !status.HasChecksum {
		return InvalidChecksumError{art.Checksum}
	}
	if !status.ChecksumInCache {
		return MissingFromCacheError{art.Checksum}
	}
	if err := os.MkdirAll(filepath.Dir(workPath), 0o755); err != nil {
		return err
	}
	cachePath = filepath.Join(ch.dir, cachePath)
	switch strat {
	case strategy.CopyStrategy:
		srcFile, err := os.Open(cachePath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// ContentsMatch is set true in quickStatus only when the workspace
		// file is a link to the correct file in the cache. In this case, we
		// can safely remove the link to allow the copy checkout to proceed.
		// Otherwise, it's best to let os.OpenFile fail below to make the user
		// fix the issue.
		if status.ContentsMatch {
			if err := os.Remove(workPath); err != nil {
				return err
			}
		}

		dstFile, err := os.OpenFile(workPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		checksum, err := checksum.Checksum(io.TeeReader(srcFile, dstFile))
		if err != nil {
			return err
		}
		if checksum != art.Checksum {
			return fmt.Errorf("found checksum %#v, expected %#v", checksum, art.Checksum)
		}
	case strategy.LinkStrategy:
		if status.ContentsMatch {
			return nil
		}
		if err := os.Symlink(cachePath, workPath); err != nil {
			return err
		}
	}
	return nil
}

func checkoutDir(
	ch LocalCache,
	workspaceDir string,
	art artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	status, cachePath, workPath, err := quickStatus(ch, workspaceDir, art)
	if err != nil {
		return err
	}
	cachePath = filepath.Join(ch.dir, cachePath)
	if !status.HasChecksum {
		return InvalidChecksumError{art.Checksum}
	}
	if !status.ChecksumInCache {
		return MissingFromCacheError{art.Checksum}
	}
	if !(status.WorkspaceFileStatus == fsutil.StatusAbsent || status.WorkspaceFileStatus == fsutil.StatusDirectory) {
		return fmt.Errorf(
			"expected target to be empty or a directory, found %s",
			status.WorkspaceFileStatus,
		)
	}
	man, err := readDirManifest(cachePath)
	if err != nil {
		return err
	}
	for _, childArt := range man.Contents {
		if err := ch.Checkout(workPath, *childArt, strat); err != nil {
			return err
		}
	}
	return nil
}
