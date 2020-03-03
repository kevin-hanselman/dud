package commit

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/kevlar1818/duc/stage"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path"
)

// Commit calculates the checksums of all outputs of a stage and adds the outputs to the DUC cache.
// TODO: This function should have 100% coverage
func Commit(s *stage.Stage, cacheDir string) error {
	for i, output := range s.Outputs {
		srcFile, err := os.Open(path.Join(s.WorkingDir, output.Path))
		defer srcFile.Close()
		if err != nil {
			return err
		}
		dstFile, err := ioutil.TempFile(cacheDir, "")
		defer dstFile.Close()
		if err != nil {
			return err
		}

		checksum, err := checksumAndCopy(srcFile, dstFile)
		if err != nil {
			return err
		}

		dstDir := path.Join(cacheDir, checksum[:2])
		if err = os.MkdirAll(dstDir, 0755); err != nil {
			return err
		}
		if err = os.Rename(dstFile.Name(), path.Join(dstDir, checksum[2:])); err != nil {
			return err
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
func checksumAndCopy(reader io.Reader, writer io.Writer) (string, error) {
	h := sha1.New()
	b := make([]byte, 8*datasize.MB)
	for {
		nBytesRead, readErr := reader.Read(b)
		if readErr != nil && readErr != io.EOF {
			return "", readErr
		}

		nBytesHashed, hashErr := h.Write(b[:nBytesRead])
		if hashErr != nil {
			return "", hashErr
		}
		if nBytesHashed != nBytesRead {
			return "", fmt.Errorf("commit: read %d byte(s), hashed %d byte(s)", nBytesRead, nBytesHashed)
		}

		if writer != nil {
			nBytesWritten, writeErr := writer.Write(b[:nBytesRead])
			if writeErr != nil {
				return "", writeErr
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
