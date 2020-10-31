package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/kevin-hanselman/dud/src/testutil"
	"github.com/pkg/errors"
)

func TestCommitIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	for _, testCase := range testutil.AllFileTestCases() {
		for _, strat := range []strategy.CheckoutStrategy{strategy.CopyStrategy, strategy.LinkStrategy} {
			t.Run(fmt.Sprintf("%s %+v", strat, testCase), func(t *testing.T) {
				testCommitIntegration(strat, testCase, t)
			})
		}
	}
}

func testCommitIntegration(strat strategy.CheckoutStrategy, statusStart artifact.ArtifactWithStatus, t *testing.T) {
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

	commitErr := cache.Commit(dirs.WorkDir, &art, strat)
	causeErr := errors.Cause(commitErr)

	// By default, expect commit to fail and leave the workspace/cache untouched.
	statusWant := statusStart.Status

	if statusStart.WorkspaceFileStatus == fsutil.StatusAbsent {
		if !os.IsNotExist(causeErr) {
			t.Fatalf("expected Commit to return a NotExist error, got %#v", causeErr)
		}
	} else if statusStart.WorkspaceFileStatus != fsutil.StatusRegularFile && !statusStart.ContentsMatch {
		if causeErr == nil {
			t.Fatal("expected Commit to return an error")
		}
	} else if commitErr != nil {
		t.Fatalf("unexpected error: %v", commitErr)
	} else {
		statusWant.HasChecksum = true
		statusWant.ContentsMatch = true
		// If Artifact.SkipCache is true, commit will not modify the cache or
		// the workspace file, so we retain the starting values for
		// ChecksumInCache and WorkspaceFileStatus.
		if !art.SkipCache {
			statusWant.ChecksumInCache = true
			switch strat {
			case strategy.CopyStrategy:
				statusWant.WorkspaceFileStatus = fsutil.StatusRegularFile
			case strategy.LinkStrategy:
				statusWant.WorkspaceFileStatus = fsutil.StatusLink
			}
			// If we started out up-to-date, we shouldn't change workspace state.
			if statusStart.WorkspaceFileStatus == fsutil.StatusLink && statusStart.ContentsMatch {
				statusWant.WorkspaceFileStatus = statusStart.WorkspaceFileStatus
			}
		}
	}

	statusGot, err := cache.Status(dirs.WorkDir, art)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(statusWant, statusGot.Status); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}

	if statusWant.ChecksumInCache {
		testCachePermissions(cache, art, t)
	}
}

func testCachePermissions(cache *LocalCache, art artifact.Artifact, t *testing.T) {
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
	if cachedFileInfo.Mode() != 0444 {
		t.Fatalf("%#v has perms %#o, want %#o", fileCachePath, cachedFileInfo.Mode(), 0444)
	}
}
