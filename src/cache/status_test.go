package cache

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/testutil"
)

func TestStatusIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	for _, testCase := range testutil.AllFileTestCases() {
		t.Run(fmt.Sprintf("%+v", testCase), func(t *testing.T) {
			testStatusIntegration(testCase, t)
		})
	}
}

func testStatusIntegration(statusWant artifact.Status, t *testing.T) {
	dirs, art, err := testutil.CreateArtifactTestCase(statusWant)
	defer os.RemoveAll(dirs.CacheDir)
	defer os.RemoveAll(dirs.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	statusWant.Artifact = art

	cache, err := NewLocalCache(dirs.CacheDir)
	if err != nil {
		t.Fatal(err)
	}

	statusGot, err := cache.Status(dirs.WorkDir, art)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(statusWant, statusGot); diff != "" {
		t.Fatalf("Status() -want +got:\n%s", diff)
	}
}
