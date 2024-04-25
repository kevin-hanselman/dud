package index

import (
	"os/exec"

	"github.com/kevin-hanselman/dud/src/agglog"
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
	rootDir string,
	recursive bool,
	ran map[string]bool,
	inProgress map[string]bool,
	logger *agglog.AggLogger,
) error {
	if _, ok := ran[stagePath]; ok {
		return nil
	}

	if inProgress[stagePath] {
		return errors.New("cycle detected")
	}
	inProgress[stagePath] = true

	stg, ok := idx[stagePath]
	if !ok {
		return unknownStageError{stagePath}
	}

	hasCommand := stg.Command != ""
	checksumUpToDate := false

	if stg.Checksum != "" {
		realChecksum, err := stg.CalculateChecksum()
		if err != nil {
			return err
		}
		checksumUpToDate = realChecksum == stg.Checksum
	}

	var doRun bool

	if hasCommand && (len(stg.Inputs) == 0) {
		// Run if we have a command and no inputs.
		doRun = true
		logger.Info.Printf("running stage %s because there is a command and no inputs\n", stagePath)
	} else if !checksumUpToDate {
		// Run if our checksum is stale.
		doRun = true
		logger.Info.Printf("running stage %s because the checksum is not up to date\n", stagePath)
	}

	for artPath, art := range stg.Inputs {
		ownerPath, _ := idx.findOwner(artPath)
		if ownerPath == "" {
			artStatus, err := ch.Status(rootDir, *art, true)
			if err != nil {
				return err
			}
			doRun = doRun || !artStatus.ContentsMatch
		} else if recursive {
			if err := idx.Run(ownerPath, ch, rootDir, recursive, ran, inProgress, logger); err != nil {
				return err
			}
			doRun = doRun || ran[ownerPath]
		}
	}
	if !doRun {
		for _, art := range stg.Outputs {
			artStatus, err := ch.Status(rootDir, *art, true)
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
		logger.Info.Printf("running stage %s\n", stagePath)
		if err := runCommand(stg.CreateCommand()); err != nil {
			return err
		}
	} else {
		logger.Info.Printf("nothing to do for stage %s\n", stagePath)
	}
	ran[stagePath] = doRun
	delete(inProgress, stagePath)
	return nil
}
