package commit

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/c2h5oh/datasize"
	"hash"
	"io"
)

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
