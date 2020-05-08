package checksum

import (
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"github.com/c2h5oh/datasize"
	"hash"
	"io"
)

// Checksummable objects can get and set their checksum
// TODO: better name? ChecksummableStruct?
type Checksummable interface {
	GetChecksum() string
	SetChecksum(string)
}

// Update calculates the checksum of a Checksummable (sans Checksum field)
// then sets the Checksum field accordingly.
func Update(c Checksummable) error {
	c.SetChecksum("")

	h := sha1.New()
	enc := gob.NewEncoder(h)
	if err := enc.Encode(c); err != nil {
		return err
	}
	c.SetChecksum(hashToHexString(h))
	return nil
}

// hashToHexString returns the sum of the hash object encoded as a hex string.
func hashToHexString(h hash.Hash) string {
	return hex.EncodeToString(h.Sum(nil))
}

// Checksum reads from reader and returns the hash of the bytes as a hex string.
func Checksum(reader io.Reader, size uint64) (string, error) {
	// Benchmarking (on a Linux laptop) shows significantly improved performance
	// on small files (KBs to 10's of MBs) using a 1KB buffer, while
	// maintaining a larger buffer on larger files.
	bufSize := datasize.KB
	if size >= 500*datasize.MB.Bytes() {
		bufSize = 8 * datasize.MB
	}
	h := sha1.New()
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
