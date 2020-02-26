package add

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/c2h5oh/datasize"
	"hash"
	"io"
)

func hashToString(h hash.Hash) string {
	return hex.EncodeToString(h.Sum(nil))
}

// Add adds the contents of reader to the cache.
func Add(reader io.Reader) (string, error) {
	h := sha1.New()
	b := make([]byte, 1*datasize.KB)
	for {
		n, readErr := reader.Read(b)
		_, writeErr := h.Write(b[:n])
		if readErr == io.EOF {
			return hashToString(h), nil
		}
		// prefer propagating the Read error,
		// as it came first
		if readErr != nil {
			return "", readErr
		}
		if writeErr != nil {
			return "", writeErr
		}
	}
}
