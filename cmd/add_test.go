package cmd

import (
	"testing"

	"github.com/go-yaml/yaml"
	"github.com/kevin-hanselman/duc/stage"
)

func TestCheckAddTypes(t *testing.T) {

	t.Run("check all stages", func(t *testing.T) {
		stageFromFileOrig := stageFromFile
		stageFromFile = func(path string) (stg stage.Stage, isLock bool, err error) {
			return
		}
		defer func() { stageFromFile = stageFromFileOrig }()

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
		stageFromFileOrig := stageFromFile
		stageFromFile = func(path string) (stg stage.Stage, isLock bool, err error) {
			err = &yaml.TypeError{}
			return
		}
		defer func() { stageFromFile = stageFromFileOrig }()

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
		stageFromFileOrig := stageFromFile

		first := true
		stageFromFile = func(path string) (stg stage.Stage, isLock bool, err error) {
			if first {
				first = false
				err = &yaml.TypeError{}
			}
			return
		}
		defer func() { stageFromFile = stageFromFileOrig }()

		stagePaths := []string{"file1", "a.duc"}

		_, err := checkAddTypes(stagePaths)

		if err == nil {
			t.Fatal("expected error thrown by using files and stages")
		}

	})
}
