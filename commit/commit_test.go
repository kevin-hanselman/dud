package commit

import (
	"bytes"
	"github.com/c2h5oh/datasize"
	"math/rand"
	"testing"
)

func TestCommit(t *testing.T) {
	inputString := "Hello, World!"
	inputReader := bytes.NewBufferString(inputString)
	outputBuffer := bytes.NewBuffer(nil)
	want := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
	output, err := Commit(inputReader, outputBuffer)
	if err != nil {
		t.Error(err)
	}
	if output != want {
		t.Errorf("Commit(%v) yielded hash '%s', want '%s'", inputString, output, want)
	}
	if outputBuffer.String() != inputString {
		t.Errorf("Commit(%v) yielded output '%s', want '%s'", inputString, outputBuffer, inputString)
	}
}

func benchmarkCommit(inputSize datasize.ByteSize, b *testing.B) {
	input := make([]byte, inputSize)
	rand.Read(input)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Commit(bytes.NewReader(input), nil)
	}
}

func BenchmarkCommit1KB(b *testing.B) { benchmarkCommit(1*datasize.KB, b) }
func BenchmarkCommit1MB(b *testing.B) { benchmarkCommit(1*datasize.MB, b) }
func BenchmarkCommit1GB(b *testing.B) { benchmarkCommit(1*datasize.GB, b) }
