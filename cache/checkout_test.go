package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/fsutil"
	"github.com/kevin-hanselman/duc/strategy"
	"github.com/kevin-hanselman/duc/testutil"
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

	if diff := cmp.Diff(statusWant, statusGot); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}
}

func TestDirectoryCheckout(t *testing.T) {
	t.Run("non-recursive", testDirectoryCheckoutNonRecursive)
	t.Run("recursive", testDirectoryCheckoutRecursive)
}

func testDirectoryCheckoutNonRecursive(t *testing.T) {
	dirQuickStatus := artifact.Status{
		WorkspaceFileStatus: fsutil.Absent,
		HasChecksum:         true,
		ChecksumInCache:     true,
	}

	strat := strategy.CopyStrategy

	fileArtifacts := []*artifact.Artifact{
		{Path: "file1"},
		{Path: "file2"},
		{Path: "file3"},
	}

	quickStatusOrig := quickStatus
	quickStatus = func(
		ch *LocalCache,
		workingDir string,
		art artifact.Artifact,
	) (status artifact.Status, cachePath, workPath string, err error) {
		status = dirQuickStatus
		workPath = filepath.Join(workingDir, art.Path)
		cachePath = "foobar"
		return
	}
	defer func() { quickStatus = quickStatusOrig }()

	checkoutFileOrig := checkoutFile
	checkoutFileArtifacts := []*artifact.Artifact{}
	checkoutFile = func(
		ch *LocalCache,
		workingDir string,
		art *artifact.Artifact,
		strat strategy.CheckoutStrategy,
	) error {
		checkoutFileArtifacts = append(checkoutFileArtifacts, art)
		return nil
	}
	defer func() { checkoutFile = checkoutFileOrig }()

	readDirManifestOrig := readDirManifest
	readDirManifest = func(path string) (directoryManifest, error) {
		man := directoryManifest{
			Path:     "art_dir",
			Contents: fileArtifacts,
		}
		return man, nil
	}
	defer func() { readDirManifest = readDirManifestOrig }()

	cache, err := NewLocalCache("/cache")
	if err != nil {
		t.Fatal(err)
	}
	dirArt := artifact.Artifact{IsDir: true, Checksum: "dummy_checksum", Path: "art_dir"}

	checkoutErr := cache.Checkout("work_dir", &dirArt, strat)
	if checkoutErr != nil {
		t.Fatal(checkoutErr)
	}

	if diff := cmp.Diff(fileArtifacts, checkoutFileArtifacts); diff != "" {
		t.Fatalf("checkoutFileArtifacts -want +got:\n%s", diff)
	}
}

func testDirectoryCheckoutRecursive(t *testing.T) {
	dirQuickStatus := artifact.Status{
		WorkspaceFileStatus: fsutil.Absent,
		HasChecksum:         true,
		ChecksumInCache:     true,
	}

	strat := strategy.CopyStrategy

	fileArtifactsParent := []*artifact.Artifact{
		{Path: "dir1", IsDir: true},
		{Path: "file1"},
		{Path: "file2"},
		{Path: "file3"},
	}

	fileArtifactsChild := []*artifact.Artifact{
		{Path: "nested_file1"},
		{Path: "nested_file2"},
	}

	expectedFileArtifacts := []*artifact.Artifact{}
	for i := range fileArtifactsChild {
		expectedFileArtifacts = append(expectedFileArtifacts, fileArtifactsChild[i])
	}
	for i := range fileArtifactsParent {
		if fileArtifactsParent[i].IsDir {
			continue
		}
		expectedFileArtifacts = append(expectedFileArtifacts, fileArtifactsParent[i])
	}

	quickStatusOrig := quickStatus
	quickStatus = func(
		ch *LocalCache,
		workingDir string,
		art artifact.Artifact,
	) (status artifact.Status, cachePath, workPath string, err error) {
		status = dirQuickStatus
		workPath = filepath.Join(workingDir, art.Path)
		cachePath = "foobar"
		return
	}
	defer func() { quickStatus = quickStatusOrig }()

	checkoutFileOrig := checkoutFile
	checkoutFileArtifacts := []*artifact.Artifact{}
	checkoutFile = func(
		ch *LocalCache,
		workingDir string,
		art *artifact.Artifact,
		strat strategy.CheckoutStrategy,
	) error {
		checkoutFileArtifacts = append(checkoutFileArtifacts, art)
		return nil
	}
	defer func() { checkoutFile = checkoutFileOrig }()

	readDirManCalls := 0
	readDirManifestOrig := readDirManifest
	readDirManifest = func(path string) (directoryManifest, error) {
		var man directoryManifest
		readDirManCalls++
		if readDirManCalls == 1 {
			man = directoryManifest{
				Path:     "art_dir",
				Contents: fileArtifactsParent,
			}
		} else if readDirManCalls == 2 {
			man = directoryManifest{
				Path:     "art_dir",
				Contents: fileArtifactsChild,
			}
		} else {
			t.Fatal("unexpected call to readDirManifest")
		}
		return man, nil
	}
	defer func() { readDirManifest = readDirManifestOrig }()

	cache, err := NewLocalCache("/cache")
	if err != nil {
		t.Fatal(err)
	}
	dirArt := artifact.Artifact{IsDir: true, Checksum: "dummy_checksum", Path: "art_dir"}

	checkoutErr := cache.Checkout("work_dir", &dirArt, strat)
	if checkoutErr != nil {
		t.Fatal(checkoutErr)
	}

	if diff := cmp.Diff(expectedFileArtifacts, checkoutFileArtifacts); diff != "" {
		t.Fatalf("checkoutFileArtifacts -want +got:\n%s", diff)
	}
}
