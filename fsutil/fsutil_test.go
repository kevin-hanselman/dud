package fsutil

import (
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
