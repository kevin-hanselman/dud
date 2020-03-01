package commit

import (
	"bytes"
	"github.com/c2h5oh/datasize"
	"github.com/kevlar1818/duc/stage"
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"
)

func TestChecksumAndCopy(t *testing.T) {
	inputString := "Hello, World!"
	inputReader := bytes.NewBufferString(inputString)
	outputBuffer := bytes.NewBuffer(nil)
	want := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
	output, err := checksumAndCopy(inputReader, outputBuffer)
	if err != nil {
		t.Error(err)
	}
	if output != want {
		t.Errorf("checksumAndCopy(%#v) yielded hash '%s', want '%s'", inputString, output, want)
	}
	if outputBuffer.String() != inputString {
		t.Errorf("checksumAndCopy(%#v) wrote output '%s', want '%s'", inputString, outputBuffer, inputString)
	}
}

func TestCommitIntegration(t *testing.T) {

	if testing.Short() {
		t.Skip()
	}

	cacheDir, err := ioutil.TempDir("", "duc_cache")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(cacheDir)

	workDir, err := ioutil.TempDir("", "duc_wspace")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(workDir)

	filePath := path.Join(workDir, "foo.txt")
	if err = ioutil.WriteFile(filePath, []byte("Hello, World!"), 0644); err != nil {
		t.Error(err)
	}

	s := stage.Stage{
		Checksum: "",
		Outputs: []stage.Artifact{
			stage.Artifact{
				Checksum: "",
				Path:     filePath,
			},
		},
	}

	expected := stage.Stage{
		Checksum: "", // TODO
		Outputs: []stage.Artifact{
			stage.Artifact{
				Checksum: "0a0a9f2a6772942557ab5355d76af442f8f65e01",
				Path:     filePath,
			},
		},
	}

	if err := Commit(&s, cacheDir); err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(expected, s); diff != "" {
		t.Errorf("Commit(stage) -want +got:\n%s", diff)
	}
}

func benchmarkChecksumAndCopy(inputSize datasize.ByteSize, b *testing.B) {
	b.StopTimer()
	b.ResetTimer()
	input := make([]byte, inputSize)
	rand.Read(input)
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, err := checksumAndCopy(bytes.NewReader(input), nil)
		b.StopTimer()
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkCommit1KB(b *testing.B) { benchmarkChecksumAndCopy(1*datasize.KB, b) }
func BenchmarkCommit1MB(b *testing.B) { benchmarkChecksumAndCopy(1*datasize.MB, b) }
func BenchmarkCommit1GB(b *testing.B) { benchmarkChecksumAndCopy(1*datasize.GB, b) }
