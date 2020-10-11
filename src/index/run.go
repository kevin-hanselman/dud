package index

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/kevin-hanselman/duc/src/cache"
	"github.com/pkg/errors"
)

// for mocking
var runCommand = func(cmd *exec.Cmd) error {
	return cmd.Run()
}

// Run runs a Stage and all upstream Stages.
func (idx Index) Run(
	stagePath string,
	ch cache.Cache,
	ran map[string]bool,
	inProgress map[string]bool,
) error {
	if _, ok := ran[stagePath]; ok {
		return nil
	}

	// If we've visited this Stage but haven't recorded its status (the check
	// above), then we're in a cycle.
	if inProgress[stagePath] {
		return errors.New("cycle detected")
	}

	inProgress[stagePath] = true
	en, ok := idx[stagePath]
	if !ok {
		return fmt.Errorf("run: unknown stage %#v", stagePath)
	}
	hasCommand := en.Stage.Command != ""
	hasDeps := len(en.Stage.Dependencies) > 0
	doRun := hasCommand && !hasDeps
	for artPath, art := range en.Stage.Dependencies {
		ownerPath, _, err := idx.findOwner(filepath.Join(en.Stage.WorkingDir, artPath))
		if err != nil {
			errors.Wrap(err, "run")
		}
		if ownerPath == "" {
			artStatus, err := ch.Status(en.Stage.WorkingDir, *art)
			if err != nil {
				return errors.Wrap(err, "run")
			}
			doRun = doRun || !artStatus.ContentsMatch
		} else {
			if err := idx.Run(ownerPath, ch, ran, inProgress); err != nil {
				return err
			}
			doRun = doRun || ran[ownerPath]
		}
	}
	if !doRun {
		for _, art := range en.Stage.Outputs {
			artStatus, err := ch.Status(en.Stage.WorkingDir, *art)
			if err != nil {
				return errors.Wrap(err, "run")
			}
			if !artStatus.ContentsMatch {
				doRun = true
				break
			}
		}
	}
	if doRun && hasCommand {
		runCommand(en.Stage.CreateCommand())
	}
	ran[stagePath] = doRun
	delete(inProgress, stagePath)
	return nil
}
