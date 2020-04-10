package fsutil

import (
	"fmt"
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

func TestIsLinkIntegration(t *testing.T) {
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
		"./bar.txt":           true,
		"./fsutil.go":         false,
	}
	for path, shouldBeLink := range tests {
		isLink, err := IsLink(path)
		if err != nil {
			t.Fatal(err)
		}
		if isLink != shouldBeLink {
			t.Errorf("IsLink(%v) = %v", path, isLink)
		}
	}
	if _, err := IsLink("foobar"); err == nil {
		t.Errorf("IsLink to nonexistent file did not return an error")
	}
}

func TestIsRegularFileIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	if err := os.Symlink("fsutil.go", "fsutil.go.symlink"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("fsutil.go.symlink")

	tests := map[string]bool{
		"./fsutil.go.symlink": false,
		"./fsutil.go":         true,
	}
	for path, shouldBeReg := range tests {
		isReg, err := IsRegularFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if isReg != shouldBeReg {
			t.Errorf("IsRegularFile(%v) = %v", path, isReg)
		}
	}
	if _, err := IsRegularFile("foobar"); err == nil {
		t.Errorf("IsRegularFile to nonexistent file did not return an error")
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
