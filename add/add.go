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

func Add(reader io.Reader) (string, error) {
	h := sha1.New()
	b := make([]byte, 1*datasize.KB)
	for {
		n_read, read_err := reader.Read(b)
		_, write_err := h.Write(b[:n_read])
		if read_err == io.EOF {
			return hashToString(h), nil
		}
		// prefer propagating the Read error,
		// as it came first
		if read_err != nil {
			return "", read_err
		}
		if write_err != nil {
			return "", write_err
		}
	}
}
