package index

import (
	"fmt"
	"path/filepath"

	"github.com/kevin-hanselman/duc/cache"
	"github.com/kevin-hanselman/duc/strategy"
	"github.com/pkg/errors"
)

// Commit commits the given Stage's Outputs and recursive Dependencies.
func (idx Index) Commit(
	stagePath string,
	ch cache.Cache,
	strat strategy.CheckoutStrategy,
	committed map[string]bool,
) error {
	if committed[stagePath] {
		return nil
	}
	errorPrefix := "stage commit"
	en, ok := idx[stagePath]
	if !ok {
		return fmt.Errorf("status: unknown stage %#v", stagePath)
	}
	for artPath, art := range en.Stage.Dependencies {
		ownerPath, upstreamArt, err := idx.findOwner(filepath.Join(en.Stage.WorkingDir, artPath))
		if err != nil {
			return errors.Wrap(err, errorPrefix)
		}
		if ownerPath == "" {
			art.SkipCache = true // always skip the cache for dependencies
			if err := ch.Commit(en.Stage.WorkingDir, art, strat); err != nil {
				return errors.Wrap(err, errorPrefix)
			}
		} else {
			if err := idx.Commit(ownerPath, ch, strat, committed); err != nil {
				return errors.Wrap(err, errorPrefix)
			}
			art.Checksum = upstreamArt.Checksum
		}
	}
	for _, art := range en.Stage.Outputs {
		if err := ch.Commit(en.Stage.WorkingDir, art, strat); err != nil {
			return errors.Wrap(err, errorPrefix)
		}
	}
	committed[stagePath] = true
	return nil
}
