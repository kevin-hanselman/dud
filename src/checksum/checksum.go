package checksum

import (
	"encoding/hex"
	"hash"
	"io"
	"sync"

	"github.com/c2h5oh/datasize"
	"github.com/zeebo/blake3"
)

// DefaultBufferSize is the size of the default internal buffer used by Checksum.
const DefaultBufferSize = 64 * datasize.KB

// Use pools to help the Go runtime save allocations and GCs. These pools drive
// a significant reduction in memory allocations in the Checksum benchmarks and
// result in an appreciable increase in throughput. On integration benchmarks,
// using these pools dramatically reduces the number and frequency of GC events
// when committing a large number of small files (e.g.
// integration/benchmarks/many_small_files). The overall variance in the
// integration benchmarks makes it hard to judge in runtime, but
// I would guesstimate around 5-15% runtime reduction.
var bufferPool = sync.Pool{
	New: func() interface{} {
		buffer := make([]byte, DefaultBufferSize)
		return &buffer
	},
}

var hasherPool = sync.Pool{
	New: func() interface{} {
		return blake3.New()
	},
}

func hashToHexString(h hash.Hash) string {
	return hex.EncodeToString(h.Sum(nil))
}

// Checksum reads from reader and returns the hash of the bytes as a hex string.
// If the buffer passed is zero-length, an internal buffer will be used.
func Checksum(reader io.Reader, buffer []byte) (string, error) {
	if len(buffer) == 0 {
		buffer = *bufferPool.Get().(*[]byte)
		defer bufferPool.Put(&buffer)
	}
	h := hasherPool.Get().(*blake3.Hasher)
	defer hasherPool.Put(h)
	h.Reset()
	if _, err := io.CopyBuffer(h, reader, buffer); err != nil {
		return "", err
	}
	return hashToHexString(h), nil
}
