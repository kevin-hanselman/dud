package fsutil

import (
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
		t.Run(path, func(t *testing.T) { testExists(path, shouldExist, t) })
	}
}

func testExists(path string, shouldExist bool, t *testing.T) {
	exists, err := Exists(path)
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
		t.Fatal("Failed to create symlink")
	}
	defer os.Remove("fsutil.go.symlink")
	if err := os.Link("fsutil.go", "fsutil.go.hardlink"); err != nil {
		t.Fatal("Failed to create hardlink")
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
			paths[0]+", "+paths[1],
			func(t *testing.T) {
				testSameFile(paths[0], paths[1], shouldBeSame, t)
			},
		)
		t.Run(
			paths[0]+", "+paths[1],
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
