package cache

import (
	"fmt"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
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

// Commit calculates the checksum of the artifact, moves it to the cache, then performs a checkout.
func (cache *LocalCache) Commit(workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error {
	srcPath := path.Join(workingDir, art.Path)
	srcFile, err := os.Open(srcPath)
	defer srcFile.Close()
	if err != nil {
		return errors.Wrapf(err, "opening %#v failed", srcPath)
	}
	dstFile, err := ioutil.TempFile(cache.Dir, "")
	defer dstFile.Close()
	if err != nil {
		return errors.Wrapf(err, "creating tempfile in %#v failed", cache.Dir)
	}
	// TODO: only copy if the cache is on a different filesystem (os.Rename if possible)
	// OR, if we're using CopyStrategy
	checksum, err := fsutil.ChecksumAndCopy(srcFile, dstFile)
	if err != nil {
		return errors.Wrapf(err, "checksum of %#v failed", srcPath)
	}
	dstDir := path.Join(cache.Dir, checksum[:2])
	if err = os.MkdirAll(dstDir, 0755); err != nil {
		return errors.Wrapf(err, "mkdirs %#v failed", dstDir)
	}
	cachePath := path.Join(dstDir, checksum[2:])
	if err = os.Rename(dstFile.Name(), cachePath); err != nil {
		return errors.Wrapf(err, "mv %#v failed", dstFile)
	}
	if err := os.Chmod(cachePath, 0444); err != nil {
		return errors.Wrapf(err, "chmod %#v failed", cachePath)
	}
	art.Checksum = checksum
	// There's no need to call Checkout if using CopyStrategy; the original file still exists.
	if strat == strategy.LinkStrategy {
		// TODO: add rm to checkout as "force" option
		if err := os.Remove(srcPath); err != nil {
			return errors.Wrapf(err, "rm %#v failed", srcPath)
		}
		return cache.Checkout(workingDir, art, strat)
	}
	return nil
}

// Checkout finds the artifact in the cache and adds a copy of/link to said
// artifact in the working directory.
func (cache *LocalCache) Checkout(workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error {
	dstPath := path.Join(workingDir, art.Path)
	srcPath, err := cache.CachePathForArtifact(*art)
	if err != nil {
		return err
	}
	switch strat {
	case strategy.CopyStrategy:
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return errors.Wrap(err, "checkout")
		}
		defer srcFile.Close()
		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return errors.Wrap(err, "checkout")
		}
		defer dstFile.Close()
		checksum, err := fsutil.ChecksumAndCopy(srcFile, dstFile)
		if err != nil {
			return errors.Wrap(err, "checkout")
		}
		if checksum != art.Checksum {
			return fmt.Errorf("checkout %#v: found checksum %#v, expected %#v", dstPath, checksum, art.Checksum)
		}
	case strategy.LinkStrategy:
		// TODO: hardlink when possible?
		if err := os.Symlink(srcPath, dstPath); err != nil {
			return errors.Wrapf(err, "link %#v -> %#v failed", srcPath, dstPath)
		}
	}
	return nil
}

// Status reports the status of an Artifact in the Cache.
func (cache *LocalCache) Status(workingDir string, art artifact.Artifact) (artifact.Status, error) {
	var status artifact.Status
	workPath := path.Join(workingDir, art.Path)
	cachePath, err := cache.CachePathForArtifact(art)
	if err != nil {
		// TODO: don't necessarily throw error, report invalid (or missing?) checksum
		return status, err
	}
	exists, err := fsutil.Exists(cachePath, false)
	if err != nil {
		return status, err
	}
	status.InCache = exists

	exists, err = fsutil.Exists(workPath, false)
	if err != nil {
		return status, err
	}

	if exists {
		// TODO: check file contents, and for incorrect link location
		linkDst, _ := os.Readlink(workPath)
		if linkDst == cachePath {
			status.FileStatus = artifact.IsLink
		} else {
			status.FileStatus = artifact.IsRegularFile
		}
	} else {
		status.FileStatus = artifact.IsAbsent
	}

	return status, nil
}
