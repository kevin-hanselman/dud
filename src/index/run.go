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

	doRun := false
	var runReason string

	// Run if we have a command and no inputs.
	if hasCommand && (len(stg.Inputs) == 0) {
		doRun = true
		runReason = "has command and no inputs"
	}

	// Run if our checksum is stale.
	if !checksumUpToDate {
		doRun = true
		runReason = "definition modified"
	}

	// Always check all upstream stages.
	for artPath, art := range stg.Inputs {
		ownerPath, _ := idx.findOwner(artPath)
		if ownerPath == "" {
			artStatus, err := ch.Status(rootDir, *art, true)
			if err != nil {
				return err
			}
			if !artStatus.ContentsMatch {
				doRun = true
				runReason = "input out-of-date"
			}
		} else if recursive {
			if err := idx.Run(ownerPath, ch, rootDir, recursive, ran, inProgress, logger); err != nil {
				return err
			}
			if ran[ownerPath] {
				doRun = true
				runReason = "upstream stage out-of-date"
			}
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
				runReason = "output out-of-date"
				break
			}
		}
	}
	if doRun {
		if hasCommand {
			logger.Info.Printf("Running stage %s (%s)", stagePath, runReason)
			cmd := stg.CreateCommand()
			// Avoid cmd.Command here because it will include "sh -c ...".
			logger.Debug.Printf("(in %s) %s\n", cmd.Dir, stg.Command)
			if err := runCommand(cmd); err != nil {
				return err
			}
		} else {
			logger.Info.Printf("nothing to do for stage %s (%s, but no command)\n", stagePath, runReason)
		}
	} else {
		logger.Info.Printf("nothing to do for stage %s (up-to-date)\n", stagePath)
	}
	ran[stagePath] = doRun
	delete(inProgress, stagePath)
	return nil
}
