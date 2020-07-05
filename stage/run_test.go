package stage

import (
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/mocks"
	"testing"
)

func TestRun(t *testing.T) {

	var runCommandCalledWith string
	runCommandOrig := runCommand
	runCommand = func(command string) error {
		runCommandCalledWith = command
		return nil
	}
	defer func() { runCommand = runCommandOrig }()

	t.Run("error if no command", func(t *testing.T) {
		stg := Stage{
			Command:    "",
			WorkingDir: "workDir",
			Outputs: []artifact.Artifact{
				{Path: "foo.txt"},
				{Path: "bar.txt"},
			},
		}
		mockCache := mocks.Cache{}

		if _, err := stg.Run(&mockCache); err == nil {
			t.Fatal("expected Stage.Run() to return an error")
		}

		mockCache.AssertNotCalled(t, "Status")
	})

	t.Run("will NOT run if artifacts up to date", func(t *testing.T) {
		stg := Stage{
			Command:    "echo 'hello world'",
			WorkingDir: "workDir",
			Outputs: []artifact.Artifact{
				{Path: "foo.txt"},
				{Path: "bar.txt"},
			},
		}
		mockCache := mocks.Cache{}

		artStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.RegularFile,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}

		for _, art := range stg.Outputs {
			mockCache.On("Status", "workDir", art).Return(artStatus, nil)
		}

		ran, err := stg.Run(&mockCache)
		if err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if ran {
			t.Fatal("expected Stage.Run() to return ran = false")
		}
	})

	t.Run("will run if artifacts NOT up to date", func(t *testing.T) {
		runCommandCalledWith = ""
		stg := Stage{
			Command:    "echo 'hello world'",
			WorkingDir: "workDir",
			Outputs: []artifact.Artifact{
				{Path: "foo.txt"},
				{Path: "bar.txt"},
			},
		}
		mockCache := mocks.Cache{}

		artStatus := artifact.Status{
			WorkspaceFileStatus: fsutil.RegularFile,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       false,
		}

		for _, art := range stg.Outputs {
			mockCache.On("Status", "workDir", art).Return(artStatus, nil)
		}

		ran, err := stg.Run(&mockCache)
		if err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if !ran {
			t.Fatal("expected Stage.Run() to return ran = true")
		}

		if runCommandCalledWith == "" {
			t.Fatal("runCommand not called")
		} else if runCommandCalledWith != stg.Command {
			t.Fatalf("runCommand called with %#v, want %#v", runCommandCalledWith, stg.Command)
		}
	})
}
