package index

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/kevin-hanselman/dud/src/cache"
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
	recursive bool,
	ran map[string]bool,
	inProgress map[string]bool,
	logger *log.Logger,
) error {
	if _, ok := ran[stagePath]; ok {
		return nil
	}

	if inProgress[stagePath] {
		return errors.New("cycle detected")
	}
	inProgress[stagePath] = true

	en, ok := idx[stagePath]
	if !ok {
		return fmt.Errorf("unknown stage %#v", stagePath)
	}

	hasCommand := en.Stage.Command != ""
	hasDeps := len(en.Stage.Dependencies) > 0
	doRun := hasCommand && !hasDeps
	for artPath, art := range en.Stage.Dependencies {
		ownerPath, _, err := idx.findOwner(artPath)
		if err != nil {
			return err
		}
		if ownerPath == "" {
			artStatus, err := ch.Status(en.Stage.WorkingDir, *art)
			if err != nil {
				return err
			}
			doRun = doRun || !artStatus.ContentsMatch
		} else if recursive {
			if err := idx.Run(ownerPath, ch, recursive, ran, inProgress, logger); err != nil {
				return err
			}
			doRun = doRun || ran[ownerPath]
		}
	}
	if !doRun {
		for _, art := range en.Stage.Outputs {
			artStatus, err := ch.Status(en.Stage.WorkingDir, *art)
			if err != nil {
				return err
			}
			if !artStatus.ContentsMatch {
				doRun = true
				break
			}
		}
	}
	if doRun && hasCommand {
		logger.Printf("running stage %s\n", stagePath)
		runCommand(en.Stage.CreateCommand())
	} else {
		logger.Printf("nothing to do for stage %s\n", stagePath)
	}
	ran[stagePath] = doRun
	delete(inProgress, stagePath)
	return nil
}
