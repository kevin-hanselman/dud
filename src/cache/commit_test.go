package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/kevin-hanselman/dud/src/testutil"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

func TestFileCommitIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	allStrategies := []strategy.CheckoutStrategy{strategy.LinkStrategy, strategy.CopyStrategy}

	t.Run("happy path", func(t *testing.T) {
		for _, strat := range allStrategies {
			in := testInput{
				Status: artifact.ArtifactWithStatus{
					Status: artifact.Status{
						WorkspaceFileStatus: fsutil.StatusRegularFile,
					},
				},
				CheckoutStrategy: strat,
			}
			out := testExpectedOutput{
				Status: artifact.Status{
					HasChecksum:     true,
					ChecksumInCache: true,
					ContentsMatch:   true,
				},
				Error: nil,
			}
			if strat == strategy.CopyStrategy {
				out.Status.WorkspaceFileStatus = fsutil.StatusRegularFile
			} else if strat == strategy.LinkStrategy {
				out.Status.WorkspaceFileStatus = fsutil.StatusLink
			} else {
				panic("unknown strategy")
			}

			t.Run(strat.String(), func(t *testing.T) {
				testCommitIntegration(in, out, t)
			})
		}
	})

	t.Run("already up-to-date", func(t *testing.T) {
		for _, strat := range allStrategies {
			status := artifact.Status{
				WorkspaceFileStatus: fsutil.StatusLink,
				HasChecksum:         true,
				ChecksumInCache:     true,
				ContentsMatch:       true,
			}
			in := testInput{
				Status:           artifact.ArtifactWithStatus{Status: status},
				CheckoutStrategy: strat,
			}
			// If we started out up-to-date, we don't change workspace state,
			// even if the checkout strategy differs. We may reconsider this in
			// the future.
			out := testExpectedOutput{
				Status: status,
				Error:  nil,
			}

			t.Run(strat.String(), func(t *testing.T) {
				testCommitIntegration(in, out, t)
			})
		}
	})

	t.Run("missing from workspace", func(t *testing.T) {
		for _, strat := range allStrategies {
			in := testInput{
				Status: artifact.ArtifactWithStatus{
					Status: artifact.Status{
						WorkspaceFileStatus: fsutil.StatusAbsent,
					},
				},
				CheckoutStrategy: strat,
			}
			out := testExpectedOutput{
				Status: artifact.Status{
					WorkspaceFileStatus: fsutil.StatusAbsent,
				},
				Error: os.ErrNotExist,
			}

			t.Run(strat.String(), func(t *testing.T) {
				testCommitIntegration(in, out, t)
			})
		}
	})

	t.Run("invalid workspace file type", func(t *testing.T) {
		fileStatuses := []fsutil.FileStatus{
			fsutil.StatusLink,
			fsutil.StatusDirectory,
		}
		for _, fileStatus := range fileStatuses {
			for _, strat := range allStrategies {
				in := testInput{
					Status: artifact.ArtifactWithStatus{
						Status: artifact.Status{
							WorkspaceFileStatus: fileStatus,
						},
					},
					CheckoutStrategy: strat,
				}
				out := testExpectedOutput{
					Status: artifact.Status{
						WorkspaceFileStatus: fileStatus,
					},
					Error: errors.New("not a regular file"),
				}

				t.Run(strat.String(), func(t *testing.T) {
					testCommitIntegration(in, out, t)
				})
			}
		}
	})

	t.Run("skip cache", func(t *testing.T) {
		for _, strat := range allStrategies {
			in := testInput{
				Status: artifact.ArtifactWithStatus{
					Artifact: artifact.Artifact{SkipCache: true},
					Status: artifact.Status{
						WorkspaceFileStatus: fsutil.StatusRegularFile,
					},
				},
				CheckoutStrategy: strat,
			}
			out := testExpectedOutput{
				Status: artifact.Status{
					WorkspaceFileStatus: fsutil.StatusRegularFile,
					HasChecksum:         true,
					ChecksumInCache:     false,
					ContentsMatch:       true,
				},
				Error: nil,
			}

			t.Run(strat.String(), func(t *testing.T) {
				testCommitIntegration(in, out, t)
			})
		}
	})
}

func testCommitIntegration(in testInput, expectedOut testExpectedOutput, t *testing.T) {
	dirs, art, err := testutil.CreateArtifactTestCase(in.Status)
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := NewLocalCache(dirs.CacheDir)
	if err != nil {
		t.Fatal(err)
	}

	commitErr := cache.Commit(dirs.WorkDir, &art, in.CheckoutStrategy)

	// Strip any context from the error (e.g. "commit hello.txt:").
	commitErr = errors.Cause(commitErr)

	assertErrorMatches(t, expectedOut.Error, commitErr)

	statusGot, err := cache.Status(dirs.WorkDir, art)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expectedOut.Status, statusGot.Status); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}

	if expectedOut.Status.ChecksumInCache {
		testCachePermissions(cache, art, t)
	}
}

func testCachePermissions(cache LocalCache, art artifact.Artifact, t *testing.T) {
	fileCachePath, err := cache.PathForChecksum(art.Checksum)
	if err != nil {
		t.Fatal(err)
	}
	fileCachePath = filepath.Join(cache.dir, fileCachePath)
	cachedFileInfo, err := os.Stat(fileCachePath)
	if err != nil {
		t.Fatal(err)
	}
	// TODO: check this in cache.Status?
	if cachedFileInfo.Mode() != 0o444 {
		t.Fatalf("%#v has perms %#o, want %#o", fileCachePath, cachedFileInfo.Mode(), 0o444)
	}
}
