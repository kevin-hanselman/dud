package index

import (
	"fmt"
	"log"

	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/pkg/errors"
)

// Commit commits the given Stage's Outputs and recursive Dependencies.
func (idx Index) Commit(
	stagePath string,
	ch cache.Cache,
	rootDir string,
	strat strategy.CheckoutStrategy,
	committed map[string]bool,
	inProgress map[string]bool,
	logger *log.Logger,
) error {
	if committed[stagePath] {
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
		return fmt.Errorf("unknown stage %#v", stagePath)
	}

	for artPath, art := range stg.Dependencies {
		ownerPath, upstreamArt, err := idx.findOwner(artPath)
		if err != nil {
			return err
		}
		if ownerPath == "" {
			// Always skip the cache for dependencies. This is also enforced in
			// stage.FromFile, but most tests obviously don't use FromFile to
			// create Stages to test against. To be safe, it's best to leave
			// this here.
			art.SkipCache = true
			if err := ch.Commit(rootDir, art, strat); err != nil {
				return err
			}
		} else {
			if err := idx.Commit(
				ownerPath,
				ch,
				rootDir,
				strat,
				committed,
				inProgress,
				logger,
			); err != nil {
				return err
			}
			art.Checksum = upstreamArt.Checksum
		}
	}
	logger.Printf("committing stage %s\n", stagePath)
	for _, art := range stg.Outputs {
		if err := ch.Commit(rootDir, art, strat); err != nil {
			return err
		}
	}
	committed[stagePath] = true
	delete(inProgress, stagePath)
	return nil
}
