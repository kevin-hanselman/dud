package artifact

import (
	"fmt"
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
)

// An Artifact is a file tracked by DUC
type Artifact struct {
	Checksum string
	Path     string
}

// A CacheItem is something that can be committed and checked out from a cache.
type CacheItem interface {
	// TODO: interface primarily for mocking in Stage unit tests
	Commit(workingDir, cacheDir string, strategy cache.CheckoutStrategy) error
	Checkout(workingDir, cacheDir string, strategy cache.CheckoutStrategy) error
}

// Commit calculates the checksum the artifact file, moves it to the cache, then checks it out.
func (a *Artifact) Commit(workingDir, cacheDir string, strategy cache.CheckoutStrategy) error {
	srcPath := path.Join(workingDir, a.Path)
	srcFile, err := os.Open(srcPath)
	defer srcFile.Close()
	if err != nil {
		return errors.Wrapf(err, "opening %#v failed", srcPath)
	}
	dstFile, err := ioutil.TempFile(cacheDir, "")
	defer dstFile.Close()
	if err != nil {
		return errors.Wrapf(err, "creating tempfile in %#v failed", cacheDir)
	}
	// TODO: only copy if the cache is on a different filesystem (os.Rename if possible)
	// OR, if we're using CopyStrategy
	checksum, err := fsutil.ChecksumAndCopy(srcFile, dstFile)
	if err != nil {
		return errors.Wrapf(err, "checksum of %#v failed", srcPath)
	}
	dstDir := path.Join(cacheDir, checksum[:2])
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
	a.Checksum = checksum
	// There's no need to call Checkout if using CopyStrategy; the original file still exists.
	if strategy == cache.LinkStrategy {
		// TODO: add rm to checkout as "force" option
		if err := os.Remove(srcPath); err != nil {
			return errors.Wrapf(err, "rm %#v failed", srcPath)
		}
		return a.Checkout(workingDir, cacheDir, strategy)
	}
	return nil
}

// Checkout finds the artifact in the cache and adds a copy of/link to said
// artifact in the working directory.
func (a *Artifact) Checkout(workingDir, cacheDir string, strategy cache.CheckoutStrategy) error {
	dstPath := path.Join(workingDir, a.Path)
	srcPath := path.Join(cacheDir, a.Checksum[:2], a.Checksum[2:])
	switch strategy {
	case cache.CopyStrategy:
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
		if checksum != a.Checksum {
			return fmt.Errorf("checkout %#v: found checksum %#v, expected %#v", dstPath, checksum, a.Checksum)
		}
	case cache.LinkStrategy:
		// TODO: hardlink when possible
		if err := os.Symlink(srcPath, dstPath); err != nil {
			return errors.Wrapf(err, "link %#v -> %#v failed", srcPath, dstPath)
		}
	}
	return nil
}
