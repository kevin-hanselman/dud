package index

import (
	"fmt"
	"log"

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
	logger *log.Logger,
) error {
	if fetched[stagePath] {
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

	for artPath := range en.Stage.Dependencies {
		ownerPath, _, err := idx.findOwner(artPath)
		if err != nil {
			return err
		}
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
	logger.Printf("fetching stage %s\n", stagePath)
	for _, art := range en.Stage.Outputs {
		if err := ch.Fetch(rootDir, remote, *art); err != nil {
			return err
		}
	}
	fetched[stagePath] = true
	delete(inProgress, stagePath)
	return nil
}
