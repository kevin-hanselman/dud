package index

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/artifact"
	"github.com/kevin-hanselman/duc/fsutil"
	"github.com/kevin-hanselman/duc/mocks"
	"github.com/kevin-hanselman/duc/stage"
)

func assertCorrectCommand(stg stage.Stage, commands map[string]*exec.Cmd, t *testing.T) {
	cmd, ok := commands[stg.Command]
	if !ok {
		t.Fatalf("%#v not found in commands", stg.Command)
	}
	lastArg := cmd.Args[len(cmd.Args)-1]
	if lastArg != stg.Command {
		t.Fatalf("cmd.Args[-1] = %#v, want %#v", lastArg, stg.Command)
	}
	if cmd.Dir != filepath.Clean(stg.WorkingDir) {
		t.Fatalf("cmd.Dir = %#v, want %#v", cmd.Dir, stg.WorkingDir)
	}
}

func TestRun(t *testing.T) {

	upToDate := artifact.Status{
		WorkspaceFileStatus: fsutil.Link,
		HasChecksum:         true,
		ChecksumInCache:     true,
		ContentsMatch:       true,
	}

	outOfDate := artifact.Status{
		WorkspaceFileStatus: fsutil.RegularFile,
		HasChecksum:         true,
		ChecksumInCache:     false,
		ContentsMatch:       false,
	}

	var commands map[string]*exec.Cmd
	var resetRunCommandMock = func() {
		commands = make(map[string]*exec.Cmd)
	}
	runCommandOrig := runCommand
	runCommand = func(cmd *exec.Cmd) error {
		lastArg := cmd.Args[len(cmd.Args)-1]
		commands[lastArg] = cmd
		return nil
	}
	defer func() { runCommand = runCommandOrig }()

	t.Run("up-to-date stage without command doesn't suggest run", func(t *testing.T) {
		defer resetRunCommandMock()
		stgA := stage.Stage{
			WorkingDir: "a",
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgA, &mockCache, upToDate)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("foo.yaml", &mockCache, ran, inProgress); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if len(commands) > 0 {
			t.Fatal("runCommand called unexpectedly")
		}

		expectedRan := map[string]bool{
			"foo.yaml": false,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}
	})

	t.Run("out-of-date stage without command does suggest run", func(t *testing.T) {
		defer resetRunCommandMock()
		stgA := stage.Stage{
			WorkingDir: "a",
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgA, &mockCache, outOfDate)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("foo.yaml", &mockCache, ran, inProgress); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if len(commands) > 0 {
			t.Fatal("runCommand called unexpectedly")
		}

		expectedRan := map[string]bool{
			"foo.yaml": true,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}
	})

	t.Run("stage with command and no deps always runs (outputs up-to-date)", func(t *testing.T) {
		defer resetRunCommandMock()
		stgA := stage.Stage{
			Command:    "echo 'Running Stage A'",
			WorkingDir: "a",
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}

		mockCache := mocks.Cache{}

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("foo.yaml", &mockCache, ran, inProgress); err != nil {
			t.Fatal(err)
		}

		if len(commands) != 1 {
			t.Fatalf("runCommand called %d time(s), want 1", len(commands))
		}

		assertCorrectCommand(stgA, commands, t)

		expectedRan := map[string]bool{
			"foo.yaml": true,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}
	})

	t.Run("stage with command and no deps always runs (outputs out-of-date)", func(t *testing.T) {
		defer resetRunCommandMock()
		stgA := stage.Stage{
			Command:    "echo 'Running Stage A'",
			WorkingDir: "a",
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgA, &mockCache, outOfDate)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("foo.yaml", &mockCache, ran, inProgress); err != nil {
			t.Fatal(err)
		}

		if len(commands) != 1 {
			t.Fatalf("runCommand called %d time(s), want 1", len(commands))
		}

		assertCorrectCommand(stgA, commands, t)

		expectedRan := map[string]bool{
			"foo.yaml": true,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}
	})

	t.Run("two stages, both up-to-date", func(t *testing.T) {
		defer resetRunCommandMock()
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		stgB := stage.Stage{
			Command: "echo 'run stage B'",
			Dependencies: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}
		idx["bar.yaml"] = &entry{Stage: stgB}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgA, &mockCache, upToDate)
		expectStageStatusCalled(&stgB, &mockCache, upToDate)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bar.yaml", &mockCache, ran, inProgress); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if len(commands) > 0 {
			t.Fatal("runCommand called unexpectedly")
		}

		expectedRan := map[string]bool{
			"foo.yaml": false,
			"bar.yaml": false,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}
	})

	t.Run("two stages, upstream out-of-date", func(t *testing.T) {
		defer resetRunCommandMock()
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		stgB := stage.Stage{
			Command: "echo 'run stage B'",
			Dependencies: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}
		idx["bar.yaml"] = &entry{Stage: stgB}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgA, &mockCache, outOfDate)
		// Don't expect downstream Stage status to be checked, as the upstream being
		// out-of-date will force the run.

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bar.yaml", &mockCache, ran, inProgress); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if len(commands) != 1 {
			t.Fatalf("runCommand called %d time(s), want 1", len(commands))
		}

		assertCorrectCommand(stgB, commands, t)

		expectedRan := map[string]bool{
			"foo.yaml": true,
			"bar.yaml": true,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}
	})

	t.Run("two stages, downstream out-of-date", func(t *testing.T) {
		defer resetRunCommandMock()
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		stgB := stage.Stage{
			Command: "echo 'run stage B'",
			Dependencies: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		idx := make(Index)
		idx["foo.yaml"] = &entry{Stage: stgA}
		idx["bar.yaml"] = &entry{Stage: stgB}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgA, &mockCache, upToDate)
		expectStageStatusCalled(&stgB, &mockCache, outOfDate)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bar.yaml", &mockCache, ran, inProgress); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if len(commands) != 1 {
			t.Fatalf("runCommand called %d times, want 1", len(commands))
		}

		assertCorrectCommand(stgB, commands, t)

		expectedRan := map[string]bool{
			"foo.yaml": false,
			"bar.yaml": true,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}
	})

	t.Run("ensure all deps are checked", func(t *testing.T) {
		defer resetRunCommandMock()
		depA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"bish.bin": {Path: "bish.bin"},
			},
		}
		depB := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"bash.bin": {Path: "bash.bin"},
			},
		}
		downstream := stage.Stage{
			Command: "echo 'generating bosh.bin'",
			Dependencies: map[string]*artifact.Artifact{
				"bish.bin": {Path: "bish.bin"},
				"bash.bin": {Path: "bash.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bosh.bin": {Path: "bosh.bin"},
			},
		}
		idx := make(Index)
		idx["bish.yaml"] = &entry{Stage: depA}
		idx["bash.yaml"] = &entry{Stage: depB}
		idx["bosh.yaml"] = &entry{Stage: downstream}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&depA, &mockCache, outOfDate)
		expectStageStatusCalled(&depB, &mockCache, upToDate)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bosh.yaml", &mockCache, ran, inProgress); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if len(commands) != 1 {
			t.Fatalf("runCommand called %d times, want 1", len(commands))
		}

		assertCorrectCommand(downstream, commands, t)

		expectedRan := map[string]bool{
			"bish.yaml": true,
			"bash.yaml": false,
			"bosh.yaml": true,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}
	})

	t.Run("handle relative paths to other work dirs", func(t *testing.T) {
		defer resetRunCommandMock()
		depA := stage.Stage{
			WorkingDir: "binDir",
			Outputs: map[string]*artifact.Artifact{
				"bish.bin": {Path: "bish.bin"},
			},
		}
		downstream := stage.Stage{
			WorkingDir: "mainDir",
			Command:    "echo 'generating bosh.bin'",
			Dependencies: map[string]*artifact.Artifact{
				"../binDir/bish.bin": {Path: "../binDir/bish.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"../binDir/bosh.bin": {Path: "../binDir/bosh.bin"},
			},
		}
		idx := make(Index)
		idx["bish.yaml"] = &entry{Stage: depA}
		idx["bosh.yaml"] = &entry{Stage: downstream}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&depA, &mockCache, outOfDate)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bosh.yaml", &mockCache, ran, inProgress); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if len(commands) != 1 {
			t.Fatalf("runCommand called %d times, want 1", len(commands))
		}

		assertCorrectCommand(downstream, commands, t)

		expectedRan := map[string]bool{
			"bish.yaml": true,
			"bosh.yaml": true,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}
	})

	t.Run("cycles are prevented", func(t *testing.T) {
		// stgA <-- stgB <-- stgC --> stgD
		//    |---------------^
		stgA := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"c.bin": {Path: "c.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"a.bin": {Path: "a.bin"},
			},
		}
		stgB := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"a.bin": {Path: "a.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"b.bin": {Path: "b.bin"},
			},
		}
		stgC := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"b.bin": {Path: "b.bin"},
				"d.bin": {Path: "d.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"c.bin": {Path: "c.bin"},
			},
		}
		stgD := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"d.bin": {Path: "d.bin"},
			},
		}
		idx := make(Index)
		idx["a.yaml"] = &entry{Stage: stgA}
		idx["b.yaml"] = &entry{Stage: stgB}
		idx["c.yaml"] = &entry{Stage: stgC}
		idx["d.yaml"] = &entry{Stage: stgD}

		mockCache := mocks.Cache{}
		// Stage D is the only Stage that could possibly be ran successfully.
		// We mock it to prevent a panic, but we don't enforce that it must be
		// called (due to random order).
		expectStageStatusCalled(&stgD, &mockCache, upToDate)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		err := idx.Run("c.yaml", &mockCache, ran, inProgress)
		if err == nil {
			t.Fatal("expected error")
		}

		expectedError := "cycle detected"
		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
		}

		expectedInProgress := map[string]bool{
			"c.yaml": true,
			"b.yaml": true,
			"a.yaml": true,
		}
		if diff := cmp.Diff(expectedInProgress, inProgress); diff != "" {
			t.Fatalf("inProgress -want +got:\n%s", diff)
		}
	})
}
