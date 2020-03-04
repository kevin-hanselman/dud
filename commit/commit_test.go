package commit

import (
	"bytes"
	"github.com/c2h5oh/datasize"
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/stage"
	"github.com/kevlar1818/duc/testutil"
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
	t.Run("Copy", func(t *testing.T) { testCommitIntegration(CopyStrategy, t) })
}

func testCommitIntegration(strategy CheckoutStrategy, t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	cacheDir, workDir, err := testutil.CreateTempDirs()
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(cacheDir)
	defer os.RemoveAll(workDir)

	fileWorkspacePath := path.Join(workDir, "foo.txt")
	if err = ioutil.WriteFile(fileWorkspacePath, []byte("Hello, World!"), 0644); err != nil {
		t.Error(err)
	}
	fileChecksum := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
	fileCachePath := path.Join(cacheDir, fileChecksum[:2], fileChecksum[2:])

	s := stage.Stage{
		WorkingDir: workDir,
		Outputs: []stage.Artifact{
			stage.Artifact{
				Checksum: "",
				Path:     "foo.txt",
			},
		},
	}

	expected := stage.Stage{
		WorkingDir: workDir,
		Outputs: []stage.Artifact{
			stage.Artifact{
				Checksum: fileChecksum,
				Path:     "foo.txt",
			},
		},
	}

	if err := Commit(&s, cacheDir, strategy); err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(expected, s); diff != "" {
		t.Errorf("Commit(stage) -want +got:\n%s", diff)
	}

	exists, err := fsutil.Exists(fileWorkspacePath)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("File %#v should exist", fileWorkspacePath)
	}
	exists, err = fsutil.Exists(fileCachePath)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("File %#v should exist", fileCachePath)
	}
	if strategy == CopyStrategy {
		// TODO check that files are distinct, with the same contents
	}
	if strategy == LinkStrategy {
		// TODO check that workspace file is a link to cache file
	}
}

func BenchmarkChecksum(b *testing.B) {
	b.Run("1KB", func(b *testing.B) { benchmarkChecksum(1*datasize.KB, b) })
	b.Run("1MB", func(b *testing.B) { benchmarkChecksum(1*datasize.MB, b) })
	b.Run("1GB", func(b *testing.B) { benchmarkChecksum(1*datasize.GB, b) })
}

func benchmarkChecksum(inputSize datasize.ByteSize, b *testing.B) {
	b.StopTimer()
	b.ResetTimer()
	input := make([]byte, inputSize)
	rand.Read(input)
	reader := bytes.NewReader(input)
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, err := checksumAndCopy(reader, nil)
		b.StopTimer()
		if err != nil {
			b.Error(err)
		}
	}
}
