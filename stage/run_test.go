package stage

import (
	"os/exec"
	"testing"

	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/fsutil"
	"github.com/kevin-hanselman/duc/mocks"
)

func TestRun(t *testing.T) {

	var runCommandCalled bool
	var command *exec.Cmd
	var resetRunCommandMock = func() {
		runCommandCalled = false
		command = new(exec.Cmd)
	}
	runCommandOrig := runCommand
	runCommand = func(cmd *exec.Cmd) error {
		runCommandCalled = true
		command = cmd
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
		assertCorrectCommand(stg, command, t)
		mockCache.AssertExpectations(t)
	})

	t.Run("run executes if dependencies out-of-date", func(t *testing.T) {
		defer resetRunCommandMock()
		stg := Stage{
			Command:    "echo 'hello world'",
			WorkingDir: "workDir",
			Dependencies: []artifact.Artifact{
				{Path: "foo.txt"},
			},
			Outputs: []artifact.Artifact{
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

		artStatus.ContentsMatch = false
		for _, art := range stg.Dependencies {
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
		assertCorrectCommand(stg, command, t)
		mockCache.AssertExpectations(t)
	})
}

func assertCorrectCommand(stg Stage, cmd *exec.Cmd, t *testing.T) {
	lastArg := cmd.Args[len(cmd.Args)-1]
	if lastArg != stg.Command {
		t.Fatalf("cmd.Args[-1] = %#v, want %#v", lastArg, stg.Command)
	}
	dirWant := "workDir"
	if cmd.Dir != dirWant {
		t.Fatalf("cmd.Dir = %#v, want %#v", cmd.Dir, dirWant)
	}
}
