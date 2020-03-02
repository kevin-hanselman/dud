package stage

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestSetChecksum(t *testing.T) {
	s := Stage{
		Checksum:   "",
		WorkingDir: "foo",
		Outputs: []Artifact{
			Artifact{
				Checksum: "abc",
				Path:     "bar.txt",
			},
		},
	}

	expected := s
	expected.Checksum = "ab6bf37cc6943336aa2ebcdfac102984c74f1e1f"

	s.SetChecksum()

	if diff := cmp.Diff(expected, s); diff != "" {
		t.Errorf("stage.SetChecksum() -want +got:\n%s", diff)
	}

	s.Checksum = "this should not affect the checksum"

	s.SetChecksum()

	if diff := cmp.Diff(expected, s); diff != "" {
		t.Errorf("stage.SetChecksum() -want +got:\n%s", diff)
	}
}
