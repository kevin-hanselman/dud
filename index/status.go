package index

import (
	"fmt"
	"path/filepath"

	"github.com/kevin-hanselman/duc/cache"
	"github.com/kevin-hanselman/duc/stage"
	"github.com/pkg/errors"
)

// Status is a map of Stage paths to Stage Statuses
type Status map[string]stage.Status

// Status returns the status for the given Stage and all upstream Stages.
func (idx Index) Status(
	stagePath string,
	ch cache.Cache,
	out Status,
	inProgress map[string]bool,
) error {
	// Exit early if we've already recorded this Stage's status.
	if _, ok := out[stagePath]; ok {
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
		return fmt.Errorf("status: unknown stage %#v", stagePath)
	}
	stageStatus := make(stage.Status)
	for artPath, art := range en.Stage.Dependencies {
		ownerPath, _, err := idx.findOwner(filepath.Join(en.Stage.WorkingDir, artPath))
		if err != nil {
			errors.Wrap(err, "status")
		}
		if ownerPath == "" {
			stageStatus[artPath], err = ch.Status(en.Stage.WorkingDir, *art)
			if err != nil {
				return errors.Wrapf(err, "status: %s", art.Path)
			}
		} else {
			if err := idx.Status(ownerPath, ch, out, inProgress); err != nil {
				return err
			}
		}
	}

	for artPath, art := range en.Stage.Outputs {
		var err error
		stageStatus[artPath], err = ch.Status(en.Stage.WorkingDir, *art)
		if err != nil {
			return errors.Wrapf(err, "status: %s", art.Path)
		}
	}
	// Record status and mark the Stage as complete.
	out[stagePath] = stageStatus
	delete(inProgress, stagePath)
	return nil
}
