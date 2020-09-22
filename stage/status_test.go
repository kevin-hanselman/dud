package stage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/fsutil"
	"github.com/kevin-hanselman/duc/mocks"
)

func TestStatus(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		stg := Stage{
			WorkingDir: "workDir",
			Outputs: map[string]*artifact.Artifact{
				"foo.txt": {Path: "foo.txt"},
				"bar.txt": {Path: "bar.txt"},
			},
		}

		artStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.RegularFile,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		mockCache := mocks.Cache{}
		expectedStageStatus := Status{}
		for _, art := range stg.Outputs {
			mockCache.On("Status", "workDir", *art).Return(artStatus, nil)
			expectedStageStatus[art.Path] = artifact.ArtifactWithStatus{
				Artifact: *art,
				Status:   artStatus,
			}
		}

		stageStatus, err := stg.Status(&mockCache, false)
		if err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if diff := cmp.Diff(expectedStageStatus, stageStatus); diff != "" {
			t.Fatalf("stage.Status() -want +got:\n%s", diff)
		}
	})

	t.Run("include dependencies", func(t *testing.T) {
		stg := Stage{
			WorkingDir: "workDir",
			Dependencies: map[string]*artifact.Artifact{
				"foo.txt": {Path: "foo.txt"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.txt": {Path: "bar.txt"},
			},
		}

		artStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.RegularFile,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		mockCache := mocks.Cache{}
		expectedStageStatus := Status{}
		for _, art := range stg.Dependencies {
			mockCache.On("Status", "workDir", *art).Return(artStatus, nil)
			expectedStageStatus[art.Path] = artifact.ArtifactWithStatus{
				Artifact: *art,
				Status:   artStatus,
			}
		}

		stageStatus, err := stg.Status(&mockCache, true)
		if err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if diff := cmp.Diff(expectedStageStatus, stageStatus); diff != "" {
			t.Fatalf("stage.Status() -want +got:\n%s", diff)
		}
	})
}
