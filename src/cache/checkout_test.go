package cache

import (
	"fmt"
	"os"
	"testing"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/kevin-hanselman/dud/src/testutil"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type testInput struct {
	Status           artifact.Status
	CheckoutStrategy strategy.CheckoutStrategy
}

type testExpectedOutput struct {
	Status artifact.Status
	Error  error
}

func TestFileCheckoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	allStrategies := []strategy.CheckoutStrategy{strategy.LinkStrategy, strategy.CopyStrategy}

	allFileStatuses := []fsutil.FileStatus{
		fsutil.StatusAbsent,
		fsutil.StatusRegularFile,
		fsutil.StatusLink,
		// TODO: consider adding StatusDirectory and StatusOther
	}
	allNonMatchingStatuses := []artifact.Status{
		{
			HasChecksum:     true,
			ChecksumInCache: true,
		},
		{
			HasChecksum:     true,
			ChecksumInCache: false,
		},
		{
			HasChecksum:     false,
			ChecksumInCache: false,
		},
	}

	t.Run("happy path", func(t *testing.T) {
		for _, strat := range allStrategies {
			in := testInput{
				Status: artifact.Status{
					WorkspaceFileStatus: fsutil.StatusAbsent,
					HasChecksum:         true,
					ChecksumInCache:     true,
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
				testFileCheckoutIntegration(in, out, t)
			})
		}
	})

	t.Run("missing/invalid checksum", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			hasChecksum := i > 0
			for _, fileStatus := range allFileStatuses {
				for _, strat := range allStrategies {
					in := testInput{
						Status: artifact.Status{
							WorkspaceFileStatus: fileStatus,
							HasChecksum:         hasChecksum,
						},
						CheckoutStrategy: strat,
					}
					out := testExpectedOutput{
						Status: artifact.Status{
							WorkspaceFileStatus: fileStatus,
							HasChecksum:         hasChecksum,
						},
					}
					if hasChecksum {
						out.Error = MissingFromCacheError{}
					} else {
						out.Error = InvalidChecksumError{}
					}

					testName := fmt.Sprintf("%s %s HasChecksum: %v", fileStatus, strat, hasChecksum)
					t.Run(testName, func(t *testing.T) {
						testFileCheckoutIntegration(in, out, t)
					})
				}
			}
		}
	})

	t.Run("skip cache", func(t *testing.T) {
		for _, status := range allNonMatchingStatuses {
			for _, fileStatus := range allFileStatuses {
				for _, strat := range allStrategies {
					status.WorkspaceFileStatus = fileStatus
					status.Artifact = artifact.Artifact{SkipCache: true}
					in := testInput{
						Status:           status,
						CheckoutStrategy: strat,
					}
					out := testExpectedOutput{
						Status: status,
						Error:  nil,
					}
					testName := fmt.Sprintf("%s %s", status, strat)
					t.Run(testName, func(t *testing.T) {
						testFileCheckoutIntegration(in, out, t)
					})
				}
			}
		}
	})

	t.Run("workspace file exists", func(t *testing.T) {
		for _, fileStatus := range []fsutil.FileStatus{fsutil.StatusRegularFile, fsutil.StatusLink} {
			for _, strat := range allStrategies {
				status := artifact.Status{
					WorkspaceFileStatus: fileStatus,
					HasChecksum:         true,
					ChecksumInCache:     true,
				}
				in := testInput{
					Status:           status,
					CheckoutStrategy: strat,
				}
				out := testExpectedOutput{
					Status: status,
					Error:  os.ErrExist,
				}

				testName := fmt.Sprintf("%s %s", status, strat)
				t.Run(testName, func(t *testing.T) {
					testFileCheckoutIntegration(in, out, t)
				})
			}
		}
	})

	t.Run("up-to-date link", func(t *testing.T) {
		for _, strat := range allStrategies {
			status := artifact.Status{
				WorkspaceFileStatus: fsutil.StatusLink,
				HasChecksum:         true,
				ChecksumInCache:     true,
				ContentsMatch:       true,
			}
			in := testInput{
				Status:           status,
				CheckoutStrategy: strat,
			}
			out := testExpectedOutput{
				Status: status,
				Error:  nil,
			}
			if strat == strategy.CopyStrategy {
				out.Status.WorkspaceFileStatus = fsutil.StatusRegularFile
			} else if strat == strategy.LinkStrategy {
				out.Status.WorkspaceFileStatus = fsutil.StatusLink
			} else {
				panic("unknown strategy")
			}

			testName := fmt.Sprintf("%s %s", status, strat)
			t.Run(testName, func(t *testing.T) {
				testFileCheckoutIntegration(in, out, t)
			})
		}
	})
}

func testFileCheckoutIntegration(in testInput, expectedOut testExpectedOutput, t *testing.T) {
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

	checkoutErr := cache.Checkout(dirs.WorkDir, art, in.CheckoutStrategy, nil)

	// Strip any context from the error (e.g. "checkout hello.txt:").
	checkoutErr = errors.Cause(checkoutErr)

	assertErrorMatches(t, expectedOut.Error, checkoutErr)

	statusGot, err := cache.Status(dirs.WorkDir, art, false)
	if err != nil {
		t.Fatal(err)
	}

	expectedOut.Status.Artifact = art

	if diff := cmp.Diff(expectedOut.Status, statusGot); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}
}

func assertErrorMatches(t *testing.T, want, got error) {
	switch want {
	case nil:
		if got != nil {
			t.Fatalf("expected no error, got %v", got)
		}
	case os.ErrExist:
		if !os.IsExist(got) {
			t.Fatalf("expected Exist error, got %v", got)
		}
	case os.ErrNotExist:
		if !os.IsNotExist(got) {
			t.Fatalf("expected NotExist error, got %v", got)
		}
	default:
		if !assert.IsType(t, want, got) {
			t.Fatalf("expected error %v, got %v", want, got)
		}
	}
}
