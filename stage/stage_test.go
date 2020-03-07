package stage

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"testing"
)

func TestSetChecksum(t *testing.T) {
	s := Stage{
		Checksum:   "",
		WorkingDir: "foo",
		Outputs: []artifact.Artifact{
			{
				Checksum: "abc",
				Path:     "bar.txt",
			},
		},
	}

	s.SetChecksum()

	if s.Checksum == "" {
		t.Fatal("stage.SetChecksum() didn't change (empty) checksum")
	}

	expected := s

	s.Checksum = "this should not affect the checksum"

	s.SetChecksum()

	if diff := cmp.Diff(expected, s); diff != "" {
		t.Fatalf("stage.SetChecksum() -want +got:\n%s", diff)
	}

	origChecksum := s.Checksum
	s.WorkingDir = "this should affect the checksum"

	s.SetChecksum()

	if s.Checksum == origChecksum {
		t.Fatal("changing stage.WorkingDir should have affected checksum")
	}
}
