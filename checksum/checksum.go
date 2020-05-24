package checksum

import (
	"encoding/gob"
	"encoding/hex"
	"github.com/c2h5oh/datasize"
	"github.com/zeebo/blake3"
	"hash"
	"io"
)

// ChecksumObject returns the checksum of an object's encoded bytes.
func ChecksumObject(v interface{}) (string, error) {
	h := blake3.New()
	enc := gob.NewEncoder(h)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return hashToHexString(h), nil
}

// hashToHexString returns the sum of the hash object encoded as a hex string.
func hashToHexString(h hash.Hash) string {
	return hex.EncodeToString(h.Sum(nil))
}

// Checksum reads from reader and returns the hash of the bytes as a hex string.
func Checksum(reader io.Reader, bufSize int64) (string, error) {
	if bufSize <= 0 {
		bufSize = int64(1 * datasize.MB)
	}
	h := blake3.New()
	buf := make([]byte, bufSize)
	tee := io.TeeReader(reader, h)
	for {
		if _, err := tee.Read(buf); err != nil {
			if err == io.EOF {
				return hashToHexString(h), nil
			}
			return "", err
		}
	}
}
