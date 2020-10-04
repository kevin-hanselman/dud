package index

import (
	"fmt"
	"os/exec"

	"github.com/kevin-hanselman/duc/cache"
	"github.com/kevin-hanselman/duc/stage"
	"github.com/pkg/errors"
)

// for mocking
var runCommand = func(cmd *exec.Cmd) error {
	return cmd.Run()
}

func isUpToDate(status stage.Status) bool {
	for _, artStatus := range status {
		if !artStatus.ContentsMatch {
			return false
		}
	}
	return true
}

// Run runs a Stage and all upstream Stages.
func (idx Index) Run(stagePath string, ch cache.Cache, ran map[string]bool) error {
	if _, ok := ran[stagePath]; ok {
		return nil
	}
	en, ok := idx[stagePath]
	if !ok {
		return fmt.Errorf("run: unknown stage %#v", stagePath)
	}
	hasCommand := en.Stage.Command != ""
	hasDeps := len(en.Stage.Dependencies) > 0
	doRun := hasCommand && !hasDeps
	for artPath, art := range en.Stage.Dependencies {
		ownerPath, _, err := idx.findOwner(artPath)
		if err != nil {
			errors.Wrap(err, "run")
		}
		if ownerPath == "" {
			artStatus, err := ch.Status(en.Stage.WorkingDir, *art)
			if err != nil {
				return errors.Wrap(err, "run")
			}
			if !artStatus.ContentsMatch {
				doRun = true
				break
			}
		} else {
			if err := idx.Run(ownerPath, ch, ran); err != nil {
				return errors.Wrap(err, "run")
			}
			if ran[ownerPath] {
				doRun = true
				break
			}
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
	return nil
}
