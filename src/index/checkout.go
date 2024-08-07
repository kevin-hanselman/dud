package index

import (
	"github.com/kevin-hanselman/dud/src/agglog"
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
	logger *agglog.AggLogger,
) error {
	if checkedOut[stagePath] {
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

	for artPath := range stg.Inputs {
		ownerPath, _ := idx.findOwner(artPath)
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
	logger.Info.Printf("Checking out stage %s", stagePath)
	for _, art := range stg.Outputs {
		if err := ch.Checkout(rootDir, *art, strat, nil, logger); err != nil {
			return err
		}
	}
	checkedOut[stagePath] = true
	delete(inProgress, stagePath)
	return nil
}
