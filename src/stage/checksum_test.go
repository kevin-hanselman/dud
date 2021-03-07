package stage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
)

func TestCalculateChecksum(t *testing.T) {
	newStage := func() Stage {
		return Stage{
			Checksum:   "foobar",
			WorkingDir: "dir",
			Outputs: map[string]*artifact.Artifact{
				"foo.txt": {Checksum: "bish", Path: "foo.txt"},
				"bar.txt": {Checksum: "bash", Path: "bar.txt"},
			},
			Dependencies: map[string]*artifact.Artifact{
				"b": {Checksum: "bosh", Path: "b", IsDir: true},
			},
		}
	}

	t.Run("calculating checksum doesn't change stage", func(t *testing.T) {
		stg := newStage()
		checksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}
		if checksum == "" {
			t.Fatal("CalculateChecksum returned empty string")
		}

		// Stage shouldn't be modified by call to CalculateChecksum.
		expectedStage := newStage()
		if diff := cmp.Diff(expectedStage, stg); diff != "" {
			t.Fatalf("Stage -want +got:\n%s", diff)
		}
	})

	t.Run("stage checksum field should not affect checksum", func(t *testing.T) {
		stg := newStage()
		expectedChecksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}

		stg.Checksum = "this should not affect the checksum"

		checksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(expectedChecksum, checksum); diff != "" {
			t.Fatalf("CalculateChecksum -want +got:\n%s", diff)
		}
	})

	t.Run("stage workdir should affect checksum", func(t *testing.T) {
		stg := newStage()
		originalChecksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}

		stg.WorkingDir = "this/should/affect/the/checksum"

		newChecksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}

		if originalChecksum == newChecksum {
			t.Fatal("changing stage.WorkingDir should have affected checksum")
		}
	})

	t.Run("stage command should affect checksum", func(t *testing.T) {
		stg := newStage()
		originalChecksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}

		stg.Command = "ls this/should/affect/the/checksum"

		newChecksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}

		if originalChecksum == newChecksum {
			t.Fatal("changing stage.Command should have affected checksum")
		}
	})

	t.Run("output artifact path should affect checksum", func(t *testing.T) {
		stg := newStage()
		originalChecksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}

		stg.Outputs["foo.txt"].Path = "cat.png"

		newChecksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}

		if originalChecksum == newChecksum {
			t.Fatal("changing stage artifact path should have affected checksum")
		}
	})

	t.Run("output artifact checksums should not affect checksum", func(t *testing.T) {
		stg := newStage()
		originalChecksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}

		stg.Outputs["foo.txt"].Checksum = "123456789"

		newChecksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(originalChecksum, newChecksum); diff != "" {
			t.Fatalf("CalculateChecksum -want +got:\n%s", diff)
		}
	})

	t.Run("artifact flags should affect checksum", func(t *testing.T) {
		stg := newStage()
		originalChecksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}

		stg.Dependencies["b"].DisableRecursion = true

		newChecksum, err := stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}
		if originalChecksum == newChecksum {
			t.Fatal("changing stage.Dependencies should have affected checksum")
		}
	})
}
