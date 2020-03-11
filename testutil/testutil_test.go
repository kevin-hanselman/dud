package testutil

import (
	"fmt"
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/fsutil"
	"os"
	"path"
	"testing"
)

func TestCreateTempDirsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dirs, err := CreateTempDirs()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirs.WorkDir)
	defer os.RemoveAll(dirs.CacheDir)
	if exists, _ := fsutil.Exists(dirs.WorkDir, false); !exists {
		t.Errorf("directory %#v doesn't exist", dirs.WorkDir)
	}
	if exists, _ := fsutil.Exists(dirs.CacheDir, false); !exists {
		t.Errorf("directory %#v doesn't exist", dirs.CacheDir)
	}
}

func TestCreateArtifactTestCaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	for _, inCache := range []bool{true, false} {
		for _, wspaceStatus := range []ArtifactWorkspaceStatus{IsAbsent, IsRegularFile, IsLink} {
			t.Run(fmt.Sprintf("%#v_%T", inCache, wspaceStatus), func(t *testing.T) {
				testCreateArtifactTestCaseIntegration(inCache, wspaceStatus, t)
			})
		}
	}
}

func testCreateArtifactTestCaseIntegration(inCache bool, wspaceStatus ArtifactWorkspaceStatus, t *testing.T) {
	dirs, art, err := CreateArtifactTestCase(inCache, wspaceStatus)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirs.WorkDir)
	defer os.RemoveAll(dirs.CacheDir)

	workPath := path.Join(dirs.WorkDir, art.Path)
	ch := cache.LocalCache{Dir: dirs.CacheDir}
	cachePath, err := ch.CachePathForArtifact(art)

	// TODO perform similar (the same?) tests as in cache_test.assertCheckoutExpectations
	exists, err := fsutil.Exists(workPath, false)
	if err != nil {
		t.Fatal(err)
	}
	shouldExist := wspaceStatus != IsAbsent
	if exists != shouldExist {
		t.Errorf("Exists(%#v) = %#v", workPath, exists)
	}

	exists, err = fsutil.Exists(cachePath, false)
	if err != nil {
		t.Fatal(err)
	}
	if exists != inCache {
		t.Errorf("Exists(%#v) = %#v", cachePath, exists)
	}
}
