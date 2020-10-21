package checksum

import (
	"encoding/hex"
	"hash"
	"io"

	"github.com/c2h5oh/datasize"
	"github.com/zeebo/blake3"
)

// hashToHexString returns the sum of the hash object encoded as a hex string.
func hashToHexString(h hash.Hash) string {
	return hex.EncodeToString(h.Sum(nil))
}

// Checksum reads from reader and returns the hash of the bytes as a hex string.
func Checksum(reader io.Reader, bufSize int64) (string, error) {
	if bufSize <= 0 {
		bufSize = int64(64 * datasize.KB)
	}
	h := blake3.New()
	buf := make([]byte, bufSize)
	if _, err := io.CopyBuffer(h, reader, buf); err != nil {
		return "", err
	}
	return hashToHexString(h), nil
}
