package fsutil

import (
	"bytes"
	"io"
	"os"

	"github.com/c2h5oh/datasize"
	"github.com/pkg/errors"
)

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
		isEndOfFileA := errA == io.EOF
		if errA != nil && !isEndOfFileA {
			return false, errors.Wrapf(errA, "read %#v failed", pathA)
		}
		nBytesReadB, errB := fileB.Read(bytesB)
		isEndOfFileB := errB == io.EOF
		if errB != nil && !isEndOfFileB {
			return false, errors.Wrapf(errB, "read %#v failed", pathB)
		}
		if nBytesReadA != nBytesReadB {
			return false, nil
		}
		if !bytes.Equal(bytesA, bytesB) {
			return false, nil
		}
		if isEndOfFileA != isEndOfFileB {
			return false, nil
		} else if isEndOfFileA {
			return true, nil
		}
	}
}
