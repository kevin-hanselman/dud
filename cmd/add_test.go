package cmd

import (
	"github.com/go-yaml/yaml"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/index"
	"github.com/kevlar1818/duc/track"
	"testing"
)

func TestAddStages(t *testing.T) {
	fromFileOrig := index.FromFile
	index.FromFile = func(path string, v interface{}) error {
		return nil
	}
	defer func() { index.FromFile = fromFileOrig }()

	stages := []string{"a.duc", "b.duc", "c.duc"}

	t.Run("add new stages", func(t *testing.T) {
		idx := make(index.Index)

		if err := addStages(stages, &idx); err != nil {
			t.Fatal(err)
		}

		for _, stage := range stages {
			inCommitList, added := idx[stage]
			if !added {
				t.Fatal("path wasn't added to the index")
			}
			if !inCommitList {
				t.Fatal("path wasn't added to the commit list")
			}
		}
	})

	t.Run("add existing stages", func(t *testing.T) {
		idx := make(index.Index)

		for _, stage := range stages {
			idx[stage] = false
		}

		if err := addStages(stages, &idx); err != nil {
			t.Fatal(err)
		}

		for _, stage := range stages {
			inCommitList, added := idx[stage]
			if !added {
				t.Fatal("path wasn't added to the index")
			}
			if !inCommitList {
				t.Fatal("path wasn't added to the commit list")
			}
		}
	})
}

func TestAddFiles(t *testing.T) {
	fromFileOrig := index.FromFile
	files := []string{"file1", "file2", "file3"}

	index.FromFile = func(path string, v interface{}) error {
		return nil
	}
	defer func() { index.FromFile = fromFileOrig }()

	toFileOrig := toFile
	toFile = func(path string, v interface{}) error {
		return nil
	}
	defer func() { toFile = toFileOrig }()

	fileStatusFromPathOrig := track.FileStatusFromPath
	track.FileStatusFromPath = func(path string) (fsutil.FileStatus, error) {
		return fsutil.RegularFile, nil
	}
	defer func() { track.FileStatusFromPath = fileStatusFromPathOrig }()

	stagePath := "out.duc"

	t.Run("add new files", func(t *testing.T) {
		idx := make(index.Index)

		if err := addArtifacts(files, &idx, stagePath, false); err != nil {
			t.Fatal(err)
		}

		inCommitList, added := idx[stagePath]
		if !added {
			t.Fatal("path wasn't added to index")
		}
		if !inCommitList {
			t.Fatal("path wasn't added to the commit list")
		}
	})
}

func TestCheckAddTypes(t *testing.T) {

	t.Run("check all stages", func(t *testing.T) {
		fromFileOrig := fromFile
		fromFile = func(path string, v interface{}) error {
			return nil
		}
		defer func() { fromFile = fromFileOrig }()

		stages := []string{"a.duc", "b.duc", "c.duc"}

		pathType, err := checkAddTypes(stages)

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

		stages := []string{"file1", "file2", "file3"}

		pathType, err := checkAddTypes(stages)

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

		stages := []string{"file1", "a.duc"}

		_, err := checkAddTypes(stages)

		if err == nil {
			t.Fatal("expected error thrown by using files and stages")
		}

	})
}
