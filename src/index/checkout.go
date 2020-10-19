package index

import (
	"fmt"
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/pkg/errors"
)

// Checkout checks out a Stage and all upstream Stages.
func (idx Index) Checkout(
	stagePath string,
	ch cache.Cache,
	strat strategy.CheckoutStrategy,
	checkedOut map[string]bool,
	inProgress map[string]bool,
) error {
	if checkedOut[stagePath] {
		return nil
	}

	// If we've visited this Stage but haven't recorded its status (the check
	// above), then we're in a cycle.
	if inProgress[stagePath] {
		return errors.New("cycle detected")
	}
	inProgress[stagePath] = true

	errorPrefix := "stage checkout"
	en, ok := idx[stagePath]
	if !ok {
		return fmt.Errorf("status: unknown stage %#v", stagePath)
	}
	for artPath := range en.Stage.Dependencies {
		ownerPath, _, err := idx.findOwner(filepath.Join(en.Stage.WorkingDir, artPath))
		if err != nil {
			return errors.Wrap(err, errorPrefix)
		}
		if ownerPath == "" {
			continue
		} else {
			if err := idx.Checkout(ownerPath, ch, strat, checkedOut, inProgress); err != nil {
				return err
			}
		}
	}
	for _, art := range en.Stage.Outputs {
		if err := ch.Checkout(en.Stage.WorkingDir, art, strat); err != nil {
			return errors.Wrap(err, errorPrefix)
		}
	}
	checkedOut[stagePath] = true
	delete(inProgress, stagePath)
	return nil
}
