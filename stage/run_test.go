package stage

import (
	"testing"

	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/mocks"
)

func TestRun(t *testing.T) {

	var runCommandCalled bool
	var runCommandCalledWith string
	var resetRunCommandMock = func() {
		runCommandCalled = false
		runCommandCalledWith = ""
	}
	runCommandOrig := runCommand
	runCommand = func(command string) error {
		runCommandCalled = true
		runCommandCalledWith = command
		return nil
	}
	defer func() { runCommand = runCommandOrig }()

	t.Run("run reports up-to-date even if no command", func(t *testing.T) {
		defer resetRunCommandMock()
		stg := Stage{
			Command:    "",
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

		upToDate, err := stg.Run(&mockCache)
		if err != nil {
			t.Fatal(err)
		}
		if !upToDate {
			t.Fatal("expected Run() to return upToDate = true")
		}
		if runCommandCalled {
			t.Fatal("runCommand unexpectedly called")
		}
		mockCache.AssertExpectations(t)
	})

	t.Run("run reports out-of-date even if no command", func(t *testing.T) {
		defer resetRunCommandMock()
		stg := Stage{
			Command:    "",
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
			ChecksumInCache:     false,
			ContentsMatch:       false,
		}

		for _, art := range stg.Outputs {
			mockCache.On("Status", "workDir", art).Return(artStatus, nil)
		}

		upToDate, err := stg.Run(&mockCache)
		if err != nil {
			t.Fatal(err)
		}
		if upToDate {
			t.Fatal("expected Run() to return upToDate = false")
		}
		if runCommandCalled {
			t.Fatal("runCommand unexpectedly called")
		}
		mockCache.AssertExpectations(t)
	})

	t.Run("run is no-op if artifacts up-to-date", func(t *testing.T) {
		defer resetRunCommandMock()
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

		upToDate, err := stg.Run(&mockCache)
		if err != nil {
			t.Fatal(err)
		}
		if !upToDate {
			t.Fatal("expected Run() to return upToDate = true")
		}
		if runCommandCalled {
			t.Fatal("runCommand unexpectedly called")
		}
		mockCache.AssertExpectations(t)
	})

	t.Run("run executes if outputs out-of-date", func(t *testing.T) {
		defer resetRunCommandMock()
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

		upToDate, err := stg.Run(&mockCache)
		if err != nil {
			t.Fatal(err)
		}
		if upToDate {
			t.Fatal("expected Run() to return upToDate = false")
		}
		if !runCommandCalled {
			t.Fatal("runCommand not called")
		}
		if runCommandCalledWith != stg.Command {
			t.Fatalf("runCommand called with %#v, want %#v", runCommandCalledWith, stg.Command)
		}
		mockCache.AssertExpectations(t)
	})

	t.Run("run executes if dependencies out-of-date", func(t *testing.T) {
		t.Fatal("TODO")
	})
}
