package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
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

func TestWorkspaceStatusFromPathIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	if err := os.Symlink("fsutil.go", "fsutil.go.symlink"); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("fsutil.go.symlink")

	tests := map[string]FileStatus{
		"./fsutil_test.go":    StatusRegularFile,
		"./foobar.txt":        StatusAbsent,
		"../fsutil":           StatusDirectory,
		"./fsutil.go.symlink": StatusLink,
	}

	for path, expectedWorkspaceStatus := range tests {
		wspaceStatus, err := FileStatusFromPath(path)
		if err != nil {
			t.Error(err)
			continue
		}
		if wspaceStatus != expectedWorkspaceStatus {
			t.Errorf(
				"WorkspaceStatusFromPath(%#v) = %s, want %s",
				path,
				wspaceStatus,
				expectedWorkspaceStatus,
			)
		}
	}
}

func TestSameFilesystemIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Run("same fs", func(t *testing.T) {
		pathA := filepath.Join("..", "..", "README.md")
		pathB := "fsutil.go"
		sameFs, err := SameFilesystem(pathA, pathB)
		if err != nil {
			t.Fatal(err)
		}
		if !sameFs {
			t.Fatalf("SameFilesystem(%#v, %#v) returned false", pathA, pathB)
		}
	})

	t.Run("different fs", func(t *testing.T) {
		pathA := "/dev/null"
		pathB := "fsutil.go"
		sameFs, err := SameFilesystem(pathA, pathB)
		if err != nil {
			t.Fatal(err)
		}
		if sameFs {
			t.Fatalf("SameFilesystem(%#v, %#v) returned true", pathA, pathB)
		}
	})
}
