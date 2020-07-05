package cache

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFileCheckoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	for _, testCase := range testutil.AllTestCases() {
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

	if !statusStart.HasChecksum {
		if checkoutErr != nil {
			return
		}
		t.Fatal("expected Checkout to raise invalid checksum error")
	}

	if !statusStart.ChecksumInCache {
		if checkoutErr != nil {
			return
		}
		t.Fatal("expected Checkout to raise missing checksum in cache error")
	}

	t.Log("checking SkipCache")
	if art.SkipCache {
		t.Log("SkipCache set")
		if checkoutErr != nil {
			t.Logf("error thrown as expected: %v", checkoutErr)
			return
		}
		t.Fatal("expected Checkout to raise error due to SkipCache = true")
	}

	if statusStart.WorkspaceFileStatus != fsutil.Absent {
		if os.IsExist(errors.Cause(checkoutErr)) {
			return
		}
		t.Fatalf("expected Checkout to raise Exist error, got %#v", checkoutErr)
	} else if checkoutErr != nil {
		t.Fatal(checkoutErr)
	}

	statusGot, err := cache.Status(dirs.WorkDir, art)
	if err != nil {
		t.Fatal(err)
	}
	statusWant := artifact.Status{
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
