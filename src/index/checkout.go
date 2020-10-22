package index

import (
	"fmt"
	"log"

	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/pkg/errors"
)

// Checkout checks out a Stage and all upstream Stages.
func (idx Index) Checkout(
	stagePath string,
	ch cache.Cache,
	rootDir string,
	strat strategy.CheckoutStrategy,
	recursive bool,
	checkedOut map[string]bool,
	inProgress map[string]bool,
	logger *log.Logger,
) error {
	if checkedOut[stagePath] {
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
			if err := idx.Checkout(
				ownerPath,
				ch,
				rootDir,
				strat,
				recursive,
				checkedOut,
				inProgress,
				logger,
			); err != nil {
				return err
			}
		}
	}
	logger.Printf("checking out stage %s\n", stagePath)
	for _, art := range en.Stage.Outputs {
		if err := ch.Checkout(rootDir, art, strat); err != nil {
			return err
		}
	}
	checkedOut[stagePath] = true
	delete(inProgress, stagePath)
	return nil
}
