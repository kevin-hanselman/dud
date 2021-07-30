package index

import (
	"fmt"

	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/pkg/errors"
)

// Fetch downloads a Stage's Outputs and the Outputs of any upstream Stages.
func (idx Index) Fetch(
	stagePath string,
	ch cache.Cache,
	rootDir string,
	recursive bool,
	remote string,
	fetched map[string]bool,
	inProgress map[string]bool,
	logger *agglog.AggLogger,
) error {
	if fetched[stagePath] {
		return nil
	}

	if inProgress[stagePath] {
		return errors.New("cycle detected")
	}
	inProgress[stagePath] = true

	stg, ok := idx[stagePath]
	if !ok {
		return fmt.Errorf("unknown stage %#v", stagePath)
	}

	for artPath := range stg.Inputs {
		ownerPath, _ := idx.findOwner(artPath)
		if ownerPath == "" {
			continue
		} else if recursive {
			if err := idx.Fetch(
				ownerPath,
				ch,
				rootDir,
				recursive,
				remote,
				fetched,
				inProgress,
				logger,
			); err != nil {
				return err
			}
		}
	}
	logger.Info.Printf("fetching stage %s\n", stagePath)
	// Call Fetch on all Outputs at once to minimize the number of rclone calls.
	arts := make([]artifact.Artifact, len(stg.Outputs))
	i := 0
	for _, art := range stg.Outputs {
		arts[i] = *art
		i++
	}
	if err := ch.Fetch(remote, arts...); err != nil {
		return err
	}
	fetched[stagePath] = true
	delete(inProgress, stagePath)
	return nil
}
