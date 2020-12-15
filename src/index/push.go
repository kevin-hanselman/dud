package index

import (
	"fmt"
	"log"

	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/pkg/errors"
)

// Push uploads a Stage's Outputs and the Outputs of any upstream Stages.
func (idx Index) Push(
	stagePath string,
	ch cache.Cache,
	rootDir string,
	recursive bool,
	remote string,
	pushed map[string]bool,
	inProgress map[string]bool,
	logger *log.Logger,
) error {
	if pushed[stagePath] {
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

	for artPath := range stg.Dependencies {
		ownerPath, _, err := idx.findOwner(artPath)
		if err != nil {
			return err
		}
		if ownerPath == "" {
			continue
		} else if recursive {
			if err := idx.Push(
				ownerPath,
				ch,
				rootDir,
				recursive,
				remote,
				pushed,
				inProgress,
				logger,
			); err != nil {
				return err
			}
		}
	}
	logger.Printf("pushing stage %s\n", stagePath)
	for _, art := range stg.Outputs {
		if err := ch.Push(rootDir, remote, *art); err != nil {
			return err
		}
	}
	pushed[stagePath] = true
	delete(inProgress, stagePath)
	return nil
}
