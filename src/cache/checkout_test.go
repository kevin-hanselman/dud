package cache

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/src/artifact"
	"github.com/kevin-hanselman/duc/src/fsutil"
	"github.com/kevin-hanselman/duc/src/strategy"
	"github.com/kevin-hanselman/duc/src/testutil"
	"github.com/pkg/errors"
)

func TestFileCheckoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	for _, testCase := range testutil.AllFileTestCases() {
		for _, strat := range []strategy.CheckoutStrategy{strategy.CopyStrategy, strategy.LinkStrategy} {
			t.Run(fmt.Sprintf("%s %+v", strat, testCase), func(t *testing.T) {
				testFileCheckoutIntegration(strat, testCase, t)
			})
		}
	}
}

func testFileCheckoutIntegration(strat strategy.CheckoutStrategy, statusStart artifact.ArtifactWithStatus, t *testing.T) {
	dirs, art, err := testutil.CreateArtifactTestCase(statusStart)
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := NewLocalCache(dirs.CacheDir)
	if err != nil {
		t.Fatal(err)
	}

	checkoutErr := cache.Checkout(dirs.WorkDir, &art, strat)

	causeErr := errors.Cause(checkoutErr)

	statusWant := statusStart.Status // By default we expect nothing to change.

	// SkipCache is checked first because HasChecksum and ChecksumInCache are
	// likely false when skipping the cache, but we don't expect any errors in this case.
	if statusStart.SkipCache {
		if checkoutErr != nil {
			t.Fatalf("expected Checkout to return no error, got %v", checkoutErr)
		}
	} else if !statusStart.HasChecksum {
		if _, ok := causeErr.(InvalidChecksumError); !ok {
			t.Fatalf("expected Checkout to return InvalidChecksumError, got %v", causeErr)
		}
	} else if !statusStart.ChecksumInCache {
		if _, ok := causeErr.(MissingFromCacheError); !ok {
			t.Fatalf("expected Checkout to return MissingFromCacheError, got %v", causeErr)
		}
	} else if statusStart.WorkspaceFileStatus != fsutil.Absent {
		// TODO: this clause should change if checkout should be a no-op if up-to-date
		if !os.IsExist(causeErr) {
			t.Fatalf("expected Checkout to return Exist error, got %#v", checkoutErr)
		}
	} else if checkoutErr != nil {
		t.Fatalf("expected Checkout to return no error, got %v", checkoutErr)
	} else {
		// If none of the above conditions are met, we expect Checkout to have
		// done it's job.
		statusWant = artifact.Status{
			HasChecksum:     true,
			ChecksumInCache: true,
			ContentsMatch:   true,
		}
		switch strat {
		case strategy.CopyStrategy:
			statusWant.WorkspaceFileStatus = fsutil.RegularFile
		case strategy.LinkStrategy:
			statusWant.WorkspaceFileStatus = fsutil.Link
		}
	}

	statusGot, err := cache.Status(dirs.WorkDir, art)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(statusWant, statusGot.Status); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}
}
