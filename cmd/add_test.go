package cmd

import (
	"github.com/go-yaml/yaml"
	"testing"
)

func TestCheckAddTypes(t *testing.T) {

	t.Run("check all stages", func(t *testing.T) {
		fromFileOrig := fromFile
		fromFile = func(path string, v interface{}) error {
			return nil
		}
		defer func() { fromFile = fromFileOrig }()

		stagePaths := []string{"a.duc", "b.duc", "c.duc"}

		pathType, err := checkAddTypes(stagePaths)

		if err != nil {
			t.Fatal(err)
		}

		if pathType != stageType {
			t.Fatalf("expected pathType: stageType got pathType: %s", pathType)
		}
	})

	t.Run("check all files", func(t *testing.T) {
		fromFileOrig := fromFile
		fromFile = func(path string, v interface{}) error {
			return &yaml.TypeError{}
		}
		defer func() { fromFile = fromFileOrig }()

		stagePaths := []string{"file1", "file2", "file3"}

		pathType, err := checkAddTypes(stagePaths)

		if err != nil {
			t.Fatal(err)
		}

		if pathType != artifactType {
			t.Fatal("expected stageType")
		}
	})

	t.Run("check files and stages throws error", func(t *testing.T) {
		fromFileOrig := fromFile

		first := true
		fromFile = func(path string, v interface{}) error {
			if first {
				first = false
				return &yaml.TypeError{}
			}
			return nil
		}
		defer func() { fromFile = fromFileOrig }()

		stagePaths := []string{"file1", "a.duc"}

		_, err := checkAddTypes(stagePaths)

		if err == nil {
			t.Fatal("expected error thrown by using files and stages")
		}

	})
}
