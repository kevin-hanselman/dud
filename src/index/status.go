package index

import (
	"fmt"

	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/stage"
	"github.com/pkg/errors"
)

// Status is a map of Stage paths to Stage Statuses
type Status map[string]stage.Status

// Status returns the status for the given Stage and all upstream Stages.
func (idx Index) Status(
	stagePath string,
	ch cache.Cache,
	rootDir string,
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

	stg, ok := idx[stagePath]
	if !ok {
		return fmt.Errorf("status: unknown stage %#v", stagePath)
	}

	stageStatus := stage.NewStatus()
	if stg.Checksum != "" {
		stageStatus.HasChecksum = true
		realChecksum, err := stg.CalculateChecksum()
		if err != nil {
			return err
		}
		stageStatus.ChecksumMatches = realChecksum == stg.Checksum
	}

	for artPath, art := range stg.Inputs {
		var err error
		ownerPath, _ := idx.findOwner(artPath)
		if ownerPath == "" {
			stageStatus.ArtifactStatus[artPath], err = ch.Status(rootDir, *art, false)
			if err != nil {
				return err
			}
		} else {
			if err := idx.Status(ownerPath, ch, rootDir, out, inProgress); err != nil {
				return err
			}
		}
	}

	for artPath, art := range stg.Outputs {
		var err error
		stageStatus.ArtifactStatus[artPath], err = ch.Status(rootDir, *art, false)
		if err != nil {
			return errors.Wrapf(err, "status: %s", art.Path)
		}
	}
	// Record status and mark the Stage as complete.
	out[stagePath] = stageStatus
	delete(inProgress, stagePath)
	return nil
}
