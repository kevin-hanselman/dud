package artifact

import (
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
	a.Checksum = checksum
	if strategy == cache.LinkStrategy {
		if err := os.Remove(srcPath); err != nil {
			return errors.Wrapf(err, "rm %#v failed", srcPath)
		}
		if err = os.Symlink(cachePath, srcPath); err != nil {
			return errors.Wrapf(err, "link %#v -> %#v failed", cachePath, srcPath)
		}
	}
	return nil
}
