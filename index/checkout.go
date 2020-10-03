package index

import (
	"fmt"

	"github.com/kevin-hanselman/duc/cache"
	"github.com/kevin-hanselman/duc/strategy"
	"github.com/pkg/errors"
)

// Checkout checks out a Stage and all upstream Stages.
func (idx Index) Checkout(
	stagePath string,
	ch cache.Cache,
	strat strategy.CheckoutStrategy,
	checkedOut map[string]bool,
) error {
	if checkedOut[stagePath] {
		return nil
	}
	errorPrefix := "stage checkout"
	en, ok := idx[stagePath]
	if !ok {
		return fmt.Errorf("status: unknown stage %#v", stagePath)
	}
	for artPath := range en.Stage.Dependencies {
		ownerPath, _, err := idx.findOwner(artPath)
		if err != nil {
			return errors.Wrap(err, errorPrefix)
		}
		if ownerPath == "" {
			continue
		} else {
			if err := idx.Checkout(ownerPath, ch, strat, checkedOut); err != nil {
				return errors.Wrap(err, errorPrefix)
			}
		}
	}
	for _, art := range en.Stage.Outputs {
		if err := ch.Checkout(en.Stage.WorkingDir, art, strat); err != nil {
			return errors.Wrap(err, errorPrefix)
		}
	}
	checkedOut[stagePath] = true
	return nil
}
