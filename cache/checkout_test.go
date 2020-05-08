package cache

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/strategy"
	"github.com/kevlar1818/duc/testutil"
	"github.com/pkg/errors"
	"os"
	"path"
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

func testFileCheckoutIntegration(strat strategy.CheckoutStrategy, statusStart artifact.Status, t *testing.T) {
	dirs, art, err := testutil.CreateArtifactTestCase(statusStart)
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	cache := LocalCache{Dir: dirs.CacheDir}

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

	if statusStart.WorkspaceFileStatus != artifact.Absent {
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
		statusWant.WorkspaceFileStatus = artifact.RegularFile
	case strategy.LinkStrategy:
		statusWant.WorkspaceFileStatus = artifact.Link
	}

	if diff := cmp.Diff(statusWant, statusGot); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}
}

func TestDirectoryCheckout(t *testing.T) {
	dirQuickStatus := artifact.Status{
		WorkspaceFileStatus: artifact.Absent,
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
	quickStatus = func(ch *LocalCache, workingDir string, art artifact.Artifact) (status artifact.Status, cachePath, workPath string, err error) {
		status = dirQuickStatus
		workPath = path.Join(workingDir, art.Path)
		cachePath = "foobar"
		return
	}
	defer func() { quickStatus = quickStatusOrig }()

	checkoutFileOrig := checkoutFile
	checkoutFileArtifacts := []*artifact.Artifact{}
	checkoutFile = func(ch *LocalCache, workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error {
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

	cache := LocalCache{Dir: "/cache"}
	dirArt := artifact.Artifact{IsDir: true, Checksum: "dummy_checksum", Path: "art_dir"}

	checkoutErr := cache.Checkout("work_dir", &dirArt, strat)
	if checkoutErr != nil {
		t.Fatal(checkoutErr)
	}

	if diff := cmp.Diff(fileArtifacts, checkoutFileArtifacts); diff != "" {
		t.Fatalf("checkoutFileArtifacts -want +got:\n%s", diff)
	}
}
