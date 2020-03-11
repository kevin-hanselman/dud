package fsutil

import (
	"bytes"
	"fmt"
	"github.com/c2h5oh/datasize"
	"math/rand"
	"os"
	"testing"
)

func TestExists(t *testing.T) {
	tests := map[string]bool{
		"./fsutil_test.go": true,
		"./foobar.txt":     false,
		"x/":               false,
		"../fsutil":        true,
		".":                true,
	}
	for path, shouldExist := range tests {
		// Since none of these files are symlinks, followLinks should be irrelevant.
		for _, followLinks := range [2]bool{true, false} {
			t.Run(
				fmt.Sprintf("%s_followLinks=%v", path, followLinks),
				func(t *testing.T) { testExists(path, followLinks, shouldExist, t) },
			)
		}
	}
}

func TestExistsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	if err := os.Symlink("fsutil.go", "fsutil.go.symlink"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("fsutil.go.symlink")
	if err := os.Symlink("foo.txt", "bar.txt"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("bar.txt")

	tests := map[string]bool{
		"./fsutil.go.symlink": true,
		"./bar.txt":           false,
	}
	for path, shouldExist := range tests {
		for _, followLinks := range [2]bool{true, false} {
			// Override test mapping if we're inspecting the links themselves,
			// as we created both of them above.
			if !followLinks {
				shouldExist = true
			}
			t.Run(
				fmt.Sprintf("%s_followLinks=%v", path, followLinks),
				func(t *testing.T) { testExists(path, followLinks, shouldExist, t) },
			)
		}
	}
}

func testExists(path string, followLinks, shouldExist bool, t *testing.T) {
	exists, err := Exists(path, followLinks)
	if err != nil {
		t.Errorf("Exists(%#v) raised error: %v", path, err)
	}
	if exists != shouldExist {
		t.Errorf("Exists(%#v) = %v", path, exists)
	}
}

func TestSameFileAndContentsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	if err := os.Symlink("fsutil.go", "fsutil.go.symlink"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("fsutil.go.symlink")
	if err := os.Link("fsutil.go", "fsutil.go.hardlink"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("fsutil.go.hardlink")

	tests := map[[2]string]bool{
		{"fsutil_test.go", "fsutil_test.go"}: true,
		{"fsutil.go", "fsutil_test.go"}:      false,
		{"fsutil_test.go", "fsutil.go"}:      false,
		{"fsutil.go", "fsutil.go.symlink"}:   true,
		{"fsutil.go", "fsutil.go.hardlink"}:  true,
	}

	for paths, shouldBeSame := range tests {
		t.Run(
			paths[0]+"=="+paths[1],
			func(t *testing.T) {
				testSameFile(paths[0], paths[1], shouldBeSame, t)
			},
		)
		t.Run(
			paths[0]+"=="+paths[1],
			func(t *testing.T) {
				testSameContents(paths[0], paths[1], shouldBeSame, t)
			},
		)
	}
}

func testSameFile(pathA, pathB string, shouldBeSame bool, t *testing.T) {
	same, err := SameFile(pathA, pathB)
	if err != nil {
		t.Errorf("SameFile(%#v, %#v) raised error: %v", pathA, pathB, err)
	}
	if same != shouldBeSame {
		t.Errorf("SameFile(%#v, %#v) = %v", pathA, pathB, same)
	}
}

func testSameContents(pathA, pathB string, shouldBeSame bool, t *testing.T) {
	same, err := SameContents(pathA, pathB)
	if err != nil {
		t.Errorf("SameFile(%#v, %#v) raised error: %v", pathA, pathB, err)
	}
	if same != shouldBeSame {
		t.Errorf("SameFile(%#v, %#v) = %v", pathA, pathB, same)
	}
}

func TestChecksumAndCopy(t *testing.T) {
	inputString := "Hello, World!"
	inputReader := bytes.NewBufferString(inputString)
	outputBuffer := bytes.NewBuffer(nil)
	want := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
	output, err := ChecksumAndCopy(inputReader, outputBuffer)
	if err != nil {
		t.Error(err)
	}
	if output != want {
		t.Errorf("ChecksumAndCopy(%#v) yielded hash '%s', want '%s'", inputString, output, want)
	}
	if outputBuffer.String() != inputString {
		t.Errorf("ChecksumAndCopy(%#v) wrote output '%s', want '%s'", inputString, outputBuffer, inputString)
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
		_, err := ChecksumAndCopy(reader, nil)
		b.StopTimer()
		if err != nil {
			b.Error(err)
		}
	}
}
