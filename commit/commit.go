package commit

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/kevlar1818/duc/stage"
	"github.com/pkg/errors"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path"
)

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
func Commit(s *stage.Stage, cacheDir string, strategy CheckoutStrategy) error {
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
		checksum, err := checksumAndCopy(srcFile, dstFile)
		if err != nil {
			return errors.Wrapf(err, "checksum of %#v failed", srcPath)
		}
		dstDir := path.Join(cacheDir, checksum[:2])
		if err = os.MkdirAll(dstDir, 0755); err != nil {
			return errors.Wrapf(err, "mkdirs %#v failed", dstDir)
		}
		if err = os.Rename(dstFile.Name(), path.Join(dstDir, checksum[2:])); err != nil {
			return errors.Wrapf(err, "mv %#v failed", dstFile)
		}
		s.Outputs[i].Checksum = checksum
	}
	return nil
}

func hashToString(h hash.Hash) string {
	return hex.EncodeToString(h.Sum(nil))
}

// checksumAndCopy copies the contents of reader to writer and returns the hash of the
// bytes as a hex string.
//
// os.Rename will be faster than copying if we aren't crossing filesystems. The
// caller of this function should account for this and pass nil as writer to
// prevent copying data.
// TODO: This function should have 100% coverage
func checksumAndCopy(reader io.Reader, writer io.Writer) (string, error) {
	h := sha1.New()
	b := make([]byte, 8*datasize.MB)
	for {
		nBytesRead, readErr := reader.Read(b)
		if readErr != nil && readErr != io.EOF {
			return "", errors.Wrap(readErr, "read failed")
		}

		nBytesHashed, hashErr := h.Write(b[:nBytesRead])
		if hashErr != nil {
			return "", errors.Wrap(hashErr, "updating hash failed")
		}
		if nBytesHashed != nBytesRead {
			return "", fmt.Errorf("commit: read %d byte(s), hashed %d byte(s)", nBytesRead, nBytesHashed)
		}

		if writer != nil {
			nBytesWritten, writeErr := writer.Write(b[:nBytesRead])
			if writeErr != nil {
				return "", errors.Wrap(writeErr, "write failed")
			}
			if nBytesWritten != nBytesRead {
				return "", fmt.Errorf("commit: read %d byte(s), wrote %d byte(s)", nBytesRead, nBytesWritten)
			}
		}

		if readErr == io.EOF {
			return hashToString(h), nil
		}
	}
}
