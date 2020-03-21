package cache

import (
	"fmt"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"github.com/pkg/errors"
	"os"
	"path"
)

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
