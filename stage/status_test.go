package stage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/mocks"
)

func TestStatus(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		stg := Stage{
			WorkingDir: "workDir",
			Outputs: []artifact.Artifact{
				{Path: "foo.txt"},
				{Path: "bar.txt"},
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
			mockCache.On("Status", "workDir", art).Return(artStatus, nil)
			expectedStageStatus[art.Path] = artifact.ArtifactWithStatus{
				Artifact: art,
				Status:   artStatus,
			}
		}

		stageStatus, err := stg.Status(&mockCache)
		if err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if diff := cmp.Diff(expectedStageStatus, stageStatus); diff != "" {
			t.Fatalf("stage.Status() -want +got:\n%s", diff)
		}
	})

	t.Run("include dependencies", func(t *testing.T) {
		//t.Fatal("TODO")
	})
}
