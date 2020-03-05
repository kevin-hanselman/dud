package stage

import (
	"crypto/sha1"
	"encoding/gob"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
)

// A Stage holds all information required to reproduce data. It is the primary
// artifact of DUC.
type Stage struct {
	Checksum   string
	WorkingDir string
	Outputs    []Artifact
}

// An Artifact is a file that is a result of reproducing a stage.
type Artifact struct {
	Checksum string
	Path     string
}

// SetChecksum calculates the checksum of a Stage (sans Checksum field)
// then sets the Checksum field accordingly.
func (s *Stage) SetChecksum() error {
	s.Checksum = ""

	h := sha1.New()
	enc := gob.NewEncoder(h)
	if err := enc.Encode(s); err != nil {
		return err
	}
	s.Checksum = fsutil.HashToHexString(h)
	return nil
}

// CheckoutStrategy enumerates the strategies for checking out files from the cache
type CheckoutStrategy int

const (
	// LinkStrategy creates read-only links to files in the cache (prefers hard links to symbolic)
	LinkStrategy CheckoutStrategy = iota
	// CopyStrategy creates copies of files in the cache
	CopyStrategy
)

// Commit calculates the checksums of all outputs of a stage and adds the outputs to the DUC cache.
// TODO: This function should have 100% coverage
func (s *Stage) Commit(cacheDir string, strategy CheckoutStrategy) error {
	for i, output := range s.Outputs {
		srcPath := path.Join(s.WorkingDir, output.Path)
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
		s.Outputs[i].Checksum = checksum
		if strategy == LinkStrategy {
			if err := os.Remove(srcPath); err != nil {
				return errors.Wrapf(err, "rm %#v failed", srcPath)
			}
			if err = os.Symlink(cachePath, srcPath); err != nil {
				return errors.Wrapf(err, "link %#v -> %#v failed", cachePath, srcPath)
			}
		}
	}
	return nil
}
