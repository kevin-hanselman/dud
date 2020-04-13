package checksum

import (
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/pkg/errors"
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

// CalculateAndCopy copies the contents of reader to writer and returns the hash of the
// bytes as a hex string. Passing nil as writer skips writing and simply checksums reader.
func CalculateAndCopy(reader io.Reader, writer io.Writer) (string, error) {
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
			return hashToHexString(h), nil
		}
	}
}
