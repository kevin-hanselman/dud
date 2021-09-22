package cache

import (
	"os"
	"testing"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/fsutil"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/goleak"
)

func TestDirectoryStatusIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	defer goleak.VerifyNone(t)

	notCommitted := func(art artifact.Artifact) *artifact.Status {
		return &artifact.Status{
			WorkspaceFileStatus: fsutil.StatusRegularFile,
			Artifact:            art,
			HasChecksum:         false,
			ChecksumInCache:     false,
			ContentsMatch:       false,
		}
	}

	maxSharedWorkers = 1
	maxDedicatedWorkers = 1

	t.Run("untracked directory", func(t *testing.T) {
		dirs, art, cache := setupDirTest(t)
		defer os.RemoveAll(dirs.CacheDir)
		defer os.RemoveAll(dirs.WorkDir)

		actualStatus, err := cache.Status(dirs.WorkDir, art, false)
		if err != nil {
			t.Fatal(err)
		}

		expectedStatus := artifact.Status{
			Artifact:            art,
			WorkspaceFileStatus: fsutil.StatusDirectory,
			HasChecksum:         false,
			ChecksumInCache:     false,
			ContentsMatch:       false,
			ChildrenStatus: map[string]*artifact.Status{
				"1.txt": notCommitted(artifact.Artifact{Path: "1.txt"}),
				"2.txt": notCommitted(artifact.Artifact{Path: "2.txt"}),
				"3.txt": notCommitted(artifact.Artifact{Path: "3.txt"}),
				"4.txt": notCommitted(artifact.Artifact{Path: "4.txt"}),
				"5.txt": notCommitted(artifact.Artifact{Path: "5.txt"}),
				"bar": {
					Artifact:            artifact.Artifact{Path: "bar", IsDir: true},
					WorkspaceFileStatus: fsutil.StatusDirectory,
					HasChecksum:         false,
					ChecksumInCache:     false,
					ContentsMatch:       false,
					ChildrenStatus: map[string]*artifact.Status{
						"4.txt": notCommitted(artifact.Artifact{Path: "4.txt"}),
						"5.txt": notCommitted(artifact.Artifact{Path: "5.txt"}),
						"6.txt": notCommitted(artifact.Artifact{Path: "6.txt"}),
						"7.txt": notCommitted(artifact.Artifact{Path: "7.txt"}),
						"8.txt": notCommitted(artifact.Artifact{Path: "8.txt"}),
					},
				},
			},
		}

		if diff := cmp.Diff(expectedStatus, actualStatus); diff != "" {
			t.Fatalf("Status -want +got:\n%s", diff)
		}
	})
}
