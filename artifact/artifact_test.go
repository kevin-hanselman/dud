package artifact

import (
	"github.com/kevlar1818/duc/cache"
	"github.com/kevlar1818/duc/testutil"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func testCheckoutIntegration(t *testing.T, strategy cache.CheckoutStrategy) {
	if testing.Short() {
		t.Skip()
	}

	cacheDir, workDir, err := testutil.CreateTempDirs()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cacheDir)
	defer os.RemoveAll(workDir)

	fileChecksum := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
	fileCacheDir := path.Join(cacheDir, fileChecksum[:2])
	fileCachePath := path.Join(fileCacheDir, fileChecksum[2:])
	if err := os.Mkdir(fileCacheDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err = ioutil.WriteFile(fileCachePath, []byte("Hello, World!"), 0444); err != nil {
		t.Fatal(err)
	}

	art := Artifact{Checksum: fileChecksum, Path: "foo.txt"}

	if err := art.Checkout(workDir, cacheDir, strategy); err != nil {
		t.Fatal(err)
	}
}
