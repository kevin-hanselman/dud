package fsutil

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/pkg/errors"
	"hash"
	"io"
	"os"
)

// Exists returns true if path is an existing file or directory, otherwise it
// returns false. If followLinks is true, then Exists will attempt to follow
// links to their target and report said target's existence. If followLinks is
// false, Exist will operate on the link itself.
func Exists(path string, followLinks bool) (bool, error) {
	var err error
	if followLinks {
		_, err = os.Stat(path)
	} else {
		_, err = os.Lstat(path)
	}
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.Wrapf(err, "stat %#v failed", path)
}

// SameFile wraps os.SameFile and handles calling os.Stat on both paths.
func SameFile(pathA, pathB string) (bool, error) {
	aInfo, err := os.Stat(pathA)
	if err != nil {
		return false, errors.Wrapf(err, "stat %#v failed", pathA)
	}
	bInfo, err := os.Stat(pathB)
	if err != nil {
		return false, errors.Wrapf(err, "stat %#v failed", pathB)
	}
	return os.SameFile(aInfo, bInfo), nil
}

// SameContents checks that two files contain the same bytes
func SameContents(pathA, pathB string) (bool, error) {
	fileA, err := os.Open(pathA)
	if err != nil {
		return false, errors.Wrapf(err, "open %#v failed", pathA)
	}
	defer fileA.Close()
	fileB, err := os.Open(pathB)
	if err != nil {
		return false, errors.Wrapf(err, "open %#v failed", pathB)
	}
	defer fileB.Close()

	bytesA := make([]byte, 8*datasize.MB)
	bytesB := make([]byte, 8*datasize.MB)
	for {
		nBytesReadA, errA := fileA.Read(bytesA)
		if errA != nil && errA != io.EOF {
			return false, errors.Wrapf(errA, "read %#v failed", pathA)
		}
		nBytesReadB, errB := fileB.Read(bytesB)
		if errB != nil && errB != io.EOF {
			return false, errors.Wrapf(errB, "read %#v failed", pathB)
		}
		if nBytesReadA != nBytesReadB {
			return false, nil
		}
		if !bytes.Equal(bytesA, bytesB) {
			return false, nil
		}
		if errA == io.EOF {
			if errB == io.EOF {
				return true, nil
			}
			return false, nil
		}
	}
}

// HashToHexString returns the sum of the hash object encoded as a hex string.
func HashToHexString(h hash.Hash) string {
	return hex.EncodeToString(h.Sum(nil))
}

// ChecksumAndCopy copies the contents of reader to writer and returns the hash of the
// bytes as a hex string. Passing nil as writer skips writing and simply checksums reader.
func ChecksumAndCopy(reader io.Reader, writer io.Writer) (string, error) {
	// TODO: This function should have 100% coverage
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
			return HashToHexString(h), nil
		}
	}
}
