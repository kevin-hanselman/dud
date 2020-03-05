package fsutil

import (
	"bytes"
	"github.com/c2h5oh/datasize"
	"github.com/pkg/errors"
	"io"
	"os"
)

// Exists returns true if a file or directory exists, otherwise false.
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
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
