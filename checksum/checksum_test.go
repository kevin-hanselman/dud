package checksum

import (
	"bytes"
	"github.com/c2h5oh/datasize"
	"io"
	"math/rand"
	"testing"
)

func TestChecksum(t *testing.T) {
	inputString := "Hello, World!"
	inputReader := bytes.NewBufferString(inputString)
	outputBuffer := bytes.NewBuffer(nil)
	want := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
	output, err := Checksum(io.TeeReader(inputReader, outputBuffer), 0)
	if err != nil {
		t.Error(err)
	}
	if output != want {
		t.Errorf("Checksum(%#v) yielded hash '%s', want '%s'", inputString, output, want)
	}
	if outputBuffer.String() != inputString {
		t.Errorf("Checksum(%#v) wrote output '%s', want '%s'", inputString, outputBuffer, inputString)
	}
}

func BenchmarkChecksum(b *testing.B) {
	b.Run("1KB",   func(b *testing.B) { benchmarkChecksum(1*datasize.KB, b) })
	b.Run("1MB",   func(b *testing.B) { benchmarkChecksum(1*datasize.MB, b) })
	b.Run("50MB",  func(b *testing.B) { benchmarkChecksum(1*datasize.MB, b) })
	b.Run("500MB", func(b *testing.B) { benchmarkChecksum(500*datasize.MB, b) })
	b.Run("1GB",   func(b *testing.B) { benchmarkChecksum(1*datasize.GB, b) })
}

func benchmarkChecksum(inputSize datasize.ByteSize, b *testing.B) {
	b.StopTimer()
	b.ResetTimer()
	input := make([]byte, inputSize)
	rand.Read(input)
	reader := bytes.NewReader(input)
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, err := Checksum(reader, inputSize.Bytes())
		b.StopTimer()
		if err != nil {
			b.Error(err)
		}
	}
}
