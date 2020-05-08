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
	"path"
)

// Checkout finds the artifact in the cache and adds a copy of/link to said
// artifact in the working directory.
func (cache *LocalCache) Checkout(workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error {
	if art.IsDir {
		return checkoutDir(cache, workingDir, art, strat)
	}
	return checkoutFile(cache, workingDir, art, strat)
}

var checkoutFile = func(ch *LocalCache, workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error {
	// TODO: refactor to use quickStatus
	dstPath := path.Join(workingDir, art.Path)
	srcPath, err := ch.PathForChecksum(art.Checksum)
	if err != nil {
		return errors.Wrap(err, "checkoutFile")
	}
	switch strat {
	case strategy.CopyStrategy:
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return errors.Wrap(err, "checkoutFile")
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return errors.Wrap(err, "checkoutFile")
		}
		defer dstFile.Close()

		checksum, err := checksum.Checksum(io.TeeReader(srcFile, dstFile), 0)
		if err != nil {
			return errors.Wrap(err, "checkoutFile")
		}
		if checksum != art.Checksum {
			return fmt.Errorf("checkout %#v: found checksum %#v, expected %#v", dstPath, checksum, art.Checksum)
		}
	case strategy.LinkStrategy:
		isRegFile, err := fsutil.IsRegularFile(srcPath)
		if err != nil {
			return errors.Wrap(err, "checkoutFile")
		}
		if !isRegFile {
			return fmt.Errorf("checkoutFile: file %#v is not a regular file", srcPath)
		}
		// TODO: hardlink when possible?
		if err := os.Symlink(srcPath, dstPath); err != nil {
			return errors.Wrap(err, "checkoutFile")
		}
	}
	return nil
}

func checkoutDir(ch *LocalCache, workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error {
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
	if !(status.WorkspaceFileStatus == artifact.Absent || status.WorkspaceFileStatus == artifact.Directory) {
		return fmt.Errorf("checkoutDir: expected target to be empty or a directory, found %v", status.WorkspaceFileStatus)
	}
	man, err := readDirManifest(cachePath)
	for _, fileArt := range man.Contents {
		if err := checkoutFile(ch, workPath, fileArt, strat); err != nil {
			// TODO: undo previous file checkouts?
			return errors.Wrap(err, "checkoutDir")
		}
	}
	return nil
}
