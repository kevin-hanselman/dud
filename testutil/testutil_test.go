package testutil

import (
	"testing"
	"github.com/kevlar1818/duc/fsutil"
	"os"
)

func TestCreateTempDirsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	cacheDir, workDir, err := CreateTempDirs()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workDir)
	defer os.RemoveAll(cacheDir)
	if exists, _ := fsutil.Exists(workDir); ! exists {
		t.Errorf("directory %#v doesn't exist", workDir)
	}
	if exists, _ := fsutil.Exists(cacheDir); ! exists {
		t.Errorf("directory %#v doesn't exist", cacheDir)
	}
}

func TestCreateArtifactTestCaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
//func CreateArtifactTestCase(args ArtifactTestCaseArgs) (*artifact.Artifact, error) {
}

