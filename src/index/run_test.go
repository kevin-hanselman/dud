package index

import (
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/mocks"
	"github.com/kevin-hanselman/dud/src/stage"
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
	upToDate := func() artifact.Status {
		return artifact.Status{
			WorkspaceFileStatus: fsutil.StatusLink,
			HasChecksum:         true,
			ChecksumInCache:     true,
			ContentsMatch:       true,
		}
	}

	outOfDate := func() artifact.Status {
		return artifact.Status{
			WorkspaceFileStatus: fsutil.StatusRegularFile,
			HasChecksum:         true,
			ChecksumInCache:     false,
			ContentsMatch:       false,
		}
	}

	rootDir := "project/root"

	var commands map[string]*exec.Cmd
	runCommandOrig := runCommand
	runCommand = func(cmd *exec.Cmd) error {
		lastArg := cmd.Args[len(cmd.Args)-1]
		commands[lastArg] = cmd
		return nil
	}
	defer func() { runCommand = runCommandOrig }()

	updateChecksum := func(stg *stage.Stage, t *testing.T) {
		var err error
		stg.Checksum, err = stg.CalculateChecksum()
		if err != nil {
			t.Fatal(err)
		}
	}

	var infoLog strings.Builder
	logger := agglog.NewNullLogger()

	resetTestHarness := func() {
		commands = make(map[string]*exec.Cmd)
		infoLog = strings.Builder{}
		logger.Info = log.New(&infoLog, "", 0)
	}

	t.Run("up-to-date stage without command doesn't suggest run", func(t *testing.T) {
		resetTestHarness()
		stgA := stage.Stage{
			WorkingDir: "a",
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		updateChecksum(&stgA, t)
		idx := Index{"foo.yaml": &stgA}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgA, &mockCache, rootDir, upToDate(), true)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("foo.yaml", &mockCache, rootDir, true, ran, inProgress, logger); err != nil {
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

		wantLog := "nothing to do for stage foo.yaml (up-to-date)\n"
		if diff := cmp.Diff(wantLog, infoLog.String()); diff != "" {
			t.Fatalf("log -want +got:\n%s", diff)
		}
	})

	t.Run("out-of-date stage without command does suggest run", func(t *testing.T) {
		resetTestHarness()
		stgA := stage.Stage{
			WorkingDir: "a",
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		updateChecksum(&stgA, t)
		idx := Index{"foo.yaml": &stgA}

		mockCache := mocks.Cache{}
		expectStageStatusCalled(&stgA, &mockCache, rootDir, outOfDate(), true)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("foo.yaml", &mockCache, rootDir, true, ran, inProgress, logger); err != nil {
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

		wantLog := "nothing to do for stage foo.yaml (output out-of-date, but no command)\n"
		if diff := cmp.Diff(wantLog, infoLog.String()); diff != "" {
			t.Fatalf("log -want +got:\n%s", diff)
		}
	})

	t.Run("stage with command and no inputs always runs (outputs up-to-date)", func(t *testing.T) {
		resetTestHarness()
		stgA := stage.Stage{
			Command:    "echo 'Running Stage A'",
			WorkingDir: "a",
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		updateChecksum(&stgA, t)
		idx := Index{
			"foo.yaml": &stgA,
		}

		mockCache := mocks.Cache{}

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("foo.yaml", &mockCache, rootDir, true, ran, inProgress, logger); err != nil {
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

		wantLog := "running stage foo.yaml (has command and no inputs)\n"
		if diff := cmp.Diff(wantLog, infoLog.String()); diff != "" {
			t.Fatalf("log -want +got:\n%s", diff)
		}
	})

	t.Run("stage with command and no inputs always runs (outputs out-of-date)", func(t *testing.T) {
		resetTestHarness()
		stg := stage.Stage{
			Command:    "echo 'Running Stage A'",
			WorkingDir: "a",
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		updateChecksum(&stg, t)
		idx := Index{
			"foo.yaml": &stg,
		}

		mockCache := mocks.Cache{}

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("foo.yaml", &mockCache, rootDir, true, ran, inProgress, logger); err != nil {
			t.Fatal(err)
		}

		if len(commands) != 1 {
			t.Fatalf("runCommand called %d time(s), want 1", len(commands))
		}

		assertCorrectCommand(stg, commands, t)

		expectedRan := map[string]bool{
			"foo.yaml": true,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}

		wantLog := "running stage foo.yaml (has command and no inputs)\n"
		if diff := cmp.Diff(wantLog, infoLog.String()); diff != "" {
			t.Fatalf("log -want +got:\n%s", diff)
		}
	})

	t.Run("two stages, both up-to-date", func(t *testing.T) {
		resetTestHarness()
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		updateChecksum(&stgA, t)
		stgB := stage.Stage{
			Command: "echo 'run stage B'",
			Inputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		updateChecksum(&stgB, t)
		idx := Index{
			"foo.yaml": &stgA,
			"bar.yaml": &stgB,
		}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgA, &mockCache, rootDir, upToDate(), true)
		expectStageStatusCalled(&stgB, &mockCache, rootDir, upToDate(), true)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bar.yaml", &mockCache, rootDir, true, ran, inProgress, logger); err != nil {
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

		wantLog := "nothing to do for stage foo.yaml (up-to-date)\n" +
			"nothing to do for stage bar.yaml (up-to-date)\n"
		if diff := cmp.Diff(wantLog, infoLog.String()); diff != "" {
			t.Fatalf("log -want +got:\n%s", diff)
		}
	})

	t.Run("two stages, upstream out-of-date", func(t *testing.T) {
		resetTestHarness()
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		updateChecksum(&stgA, t)
		stgB := stage.Stage{
			Command: "echo 'run stage B'",
			Inputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		updateChecksum(&stgB, t)
		idx := Index{
			"foo.yaml": &stgA,
			"bar.yaml": &stgB,
		}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgA, &mockCache, rootDir, outOfDate(), true)
		// Don't expect downstream Stage status to be checked, as the upstream being
		// out-of-date will force the run.

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bar.yaml", &mockCache, rootDir, true, ran, inProgress, logger); err != nil {
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

		wantLog := "nothing to do for stage foo.yaml (output out-of-date, but no command)\n" +
			"running stage bar.yaml (upstream stage out-of-date)\n"
		if diff := cmp.Diff(wantLog, infoLog.String()); diff != "" {
			t.Fatalf("log -want +got:\n%s", diff)
		}
	})

	t.Run("two stages, downstream out-of-date", func(t *testing.T) {
		resetTestHarness()
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		updateChecksum(&stgA, t)
		stgB := stage.Stage{
			Command: "echo 'run stage B'",
			Inputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		updateChecksum(&stgB, t)
		idx := Index{
			"foo.yaml": &stgA,
			"bar.yaml": &stgB,
		}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgA, &mockCache, rootDir, upToDate(), true)
		expectStageStatusCalled(&stgB, &mockCache, rootDir, outOfDate(), true)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bar.yaml", &mockCache, rootDir, true, ran, inProgress, logger); err != nil {
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

		wantLog := "nothing to do for stage foo.yaml (up-to-date)\n" +
			"running stage bar.yaml (output out-of-date)\n"
		if diff := cmp.Diff(wantLog, infoLog.String()); diff != "" {
			t.Fatalf("log -want +got:\n%s", diff)
		}
	})

	t.Run("ensure all inputs are checked", func(t *testing.T) {
		resetTestHarness()
		inA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"bish.bin": {Path: "bish.bin"},
			},
		}
		updateChecksum(&inA, t)
		inB := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"bash.bin": {Path: "bash.bin"},
			},
		}
		updateChecksum(&inB, t)
		downstream := stage.Stage{
			Command: "echo 'generating bosh.bin'",
			Inputs: map[string]*artifact.Artifact{
				"bish.bin": {Path: "bish.bin"},
				"bash.bin": {Path: "bash.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bosh.bin": {Path: "bosh.bin"},
			},
		}
		updateChecksum(&downstream, t)
		idx := Index{
			"bish.yaml": &inA,
			"bash.yaml": &inB,
			"bosh.yaml": &downstream,
		}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&inA, &mockCache, rootDir, outOfDate(), true)
		expectStageStatusCalled(&inB, &mockCache, rootDir, upToDate(), true)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bosh.yaml", &mockCache, rootDir, true, ran, inProgress, logger); err != nil {
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

		gotLogs := make(map[string]bool)
		for _, line := range strings.Split(infoLog.String(), "\n") {
			if len(line) > 0 {
				gotLogs[line] = true
			}
		}
		wantLogs := map[string]bool{
			"nothing to do for stage bish.yaml (output out-of-date, but no command)": true,
			"nothing to do for stage bash.yaml (up-to-date)":                         true,
			"running stage bosh.yaml (upstream stage out-of-date)":                   true,
		}
		if diff := cmp.Diff(wantLogs, gotLogs); diff != "" {
			t.Fatalf("log -want +got:\n%s", diff)
		}
	})

	t.Run("cycles are prevented", func(t *testing.T) {
		resetTestHarness()
		// stgA <-- stgB <-- stgC --> stgD
		//    |---------------^
		stgA := stage.Stage{
			Inputs: map[string]*artifact.Artifact{
				"c.bin": {Path: "c.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"a.bin": {Path: "a.bin"},
			},
		}
		stgB := stage.Stage{
			Inputs: map[string]*artifact.Artifact{
				"a.bin": {Path: "a.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"b.bin": {Path: "b.bin"},
			},
		}
		stgC := stage.Stage{
			Inputs: map[string]*artifact.Artifact{
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
		idx := Index{
			"a.yaml": &stgA,
			"b.yaml": &stgB,
			"c.yaml": &stgC,
			"d.yaml": &stgD,
		}

		mockCache := mocks.Cache{}
		// Stage D is the only Stage that could possibly be ran successfully.
		// We mock it to prevent a panic, but we don't enforce that it must be
		// called (due to random order).
		expectStageStatusCalled(&stgD, &mockCache, rootDir, upToDate(), true)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		err := idx.Run("c.yaml", &mockCache, rootDir, true, ran, inProgress, logger)
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

	t.Run("run when any orphan input is out-of-date", func(t *testing.T) {
		resetTestHarness()
		bish := upToDate()
		bish.Artifact = artifact.Artifact{Path: "bish.bin"}

		bash := outOfDate()
		bash.Artifact = artifact.Artifact{Path: "bash.bin"}

		stg := stage.Stage{
			Command: "echo 'generating bosh.bin'",
			Inputs: map[string]*artifact.Artifact{
				"bish.bin": &(bish.Artifact),
				"bash.bin": &(bash.Artifact),
			},
			Outputs: map[string]*artifact.Artifact{
				"bosh.bin": {Path: "bosh.bin"},
			},
		}
		idx := Index{
			"bosh.yaml": &stg,
		}

		mockCache := mocks.Cache{}

		mockCache.On("Status", rootDir, bish.Artifact, true).Return(bish, nil).Once()
		mockCache.On("Status", rootDir, bash.Artifact, true).Return(bash, nil).Once()

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bosh.yaml", &mockCache, rootDir, true, ran, inProgress, logger); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if len(commands) != 1 {
			t.Fatalf("runCommand called %d times, want 1", len(commands))
		}

		assertCorrectCommand(stg, commands, t)

		expectedRan := map[string]bool{
			"bosh.yaml": true,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}

		wantLog := "running stage bosh.yaml (input out-of-date)\n"
		if diff := cmp.Diff(wantLog, infoLog.String()); diff != "" {
			t.Fatalf("log -want +got:\n%s", diff)
		}
	})

	t.Run("can disable recursion", func(t *testing.T) {
		resetTestHarness()
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		stgB := stage.Stage{
			Command: "echo 'run stage B'",
			Inputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		updateChecksum(&stgB, t)
		idx := Index{
			"foo.yaml": &stgA,
			"bar.yaml": &stgB,
		}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgB, &mockCache, rootDir, outOfDate(), true)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bar.yaml", &mockCache, rootDir, false, ran, inProgress, logger); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if len(commands) != 1 {
			t.Fatalf("runCommand called %d time(s), want 1", len(commands))
		}

		assertCorrectCommand(stgB, commands, t)

		expectedRan := map[string]bool{
			"bar.yaml": true,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}

		wantLog := "running stage bar.yaml (output out-of-date)\n"
		if diff := cmp.Diff(wantLog, infoLog.String()); diff != "" {
			t.Fatalf("log -want +got:\n%s", diff)
		}
	})

	t.Run("stage with out-of-date definition does run", func(t *testing.T) {
		resetTestHarness()
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		updateChecksum(&stgA, t)
		stgB := stage.Stage{
			WorkingDir: "b",
			Checksum:   "stale",
			Command:    "echo running stage B",
			Inputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		idx := Index{
			"foo.yaml": &stgA,
			"bar.yaml": &stgB,
		}

		mockCache := mocks.Cache{}

		expectStageStatusCalled(&stgA, &mockCache, rootDir, upToDate(), true)

		ran := make(map[string]bool)
		inProgress := make(map[string]bool)
		if err := idx.Run("bar.yaml", &mockCache, rootDir, true, ran, inProgress, logger); err != nil {
			t.Fatal(err)
		}

		mockCache.AssertExpectations(t)

		if len(commands) != 1 {
			t.Fatalf("runCommand called %d time(s), want 1", len(commands))
		}

		expectedRan := map[string]bool{
			"foo.yaml": false,
			"bar.yaml": true,
		}
		if diff := cmp.Diff(expectedRan, ran); diff != "" {
			t.Fatalf("committed -want +got:\n%s", diff)
		}

		wantLog := "nothing to do for stage foo.yaml (up-to-date)\n" +
			"running stage bar.yaml (definition modified)\n"
		if diff := cmp.Diff(wantLog, infoLog.String()); diff != "" {
			t.Fatalf("log -want +got:\n%s", diff)
		}
	})
}
