package cache

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/checksum"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Checkout finds the artifact in the cache and adds a copy of/link to said
// artifact in the working directory.
func (cache LocalCache) Checkout(
	workspaceDir string,
	art artifact.Artifact,
	strat strategy.CheckoutStrategy,
	progress *pb.ProgressBar,
) (err error) {
	if art.SkipCache {
		return
	}
	if progress == nil {
		progress = newProgress(art.Path)
	}
	if art.IsDir {
		activeSharedWorkers := make(chan struct{}, maxSharedWorkers)
		err = checkoutDir(
			context.Background(),
			cache,
			workspaceDir,
			art,
			strat,
			activeSharedWorkers,
			progress,
		)
	} else {
		err = checkoutFile(cache, workspaceDir, art, strat, progress)
	}
	progress.Finish()
	return errors.Wrapf(err, "checkout %s", art.Path)
}

func checkoutFile(
	ch LocalCache,
	workspaceDir string,
	art artifact.Artifact,
	strat strategy.CheckoutStrategy,
	progress *pb.ProgressBar,
) error {
	status, cachePath, workPath, err := quickStatus(ch, workspaceDir, art)
	if err != nil {
		return err
	}
	if !status.HasChecksum {
		return InvalidChecksumError{art.Checksum}
	}
	if !status.ChecksumInCache {
		return MissingFromCacheError{art.Checksum}
	}
	if err := os.MkdirAll(filepath.Dir(workPath), 0o755); err != nil {
		return err
	}
	cachePath = filepath.Join(ch.dir, cachePath)
	switch strat {
	case strategy.CopyStrategy:
		srcInfo, err := os.Lstat(cachePath)
		if err != nil {
			return err
		}
		progress.AddTotal(srcInfo.Size())
		progress.Start()

		srcFile, err := os.Open(cachePath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// ContentsMatch is set true in quickStatus only when the workspace
		// file is a link to the correct file in the cache. In this case, we
		// can safely remove the link to allow the copy checkout to proceed.
		// Otherwise, it's best to let os.OpenFile fail below to make the user
		// fix the issue.
		if status.ContentsMatch {
			if err := os.Remove(workPath); err != nil {
				return err
			}
		}

		dstFile, err := os.OpenFile(workPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// Might as well checksum the file while we copy to check data integrity.
		srcReader := io.TeeReader(progress.NewProxyReader(srcFile), dstFile)
		checksum, err := checksum.Checksum(srcReader)
		if err != nil {
			return err
		}
		if checksum != art.Checksum {
			return fmt.Errorf("found checksum %#v, expected %#v", checksum, art.Checksum)
		}
	case strategy.LinkStrategy:
		// Increment the count of files linked. If we're part of a directory,
		// checkoutDir() has already accounted for this file and started the
		// report. If we're a single file, this operation is essentially
		// instant, so we let this be a noop and don't bother starting
		// a report.
		defer progress.Increment()
		if status.ContentsMatch {
			return nil
		}
		// Make the symlink target relative to the parent directory of the
		// workspace file. For cache locations defined relative to the project
		// root (including the default location), this allows the project root
		// directory to move without invalidating the links to the cache.
		// TODO: For cache locations defined as absolute paths (e.g.
		// /mnt/my_shared_dud_cache), this change has the opposite effect;
		// moving the project may invalidate cache links. To completely
		// eliminate the link invalidation, we'd need to know if the cache is
		// a relative or absolute path and choose the linking strategy
		// accordingly. For now, always using relative link targets gives the
		// best user experience for the default cache location, so it is
		// preferred to absolute links.
		linkPath, err := filepath.Rel(filepath.Dir(workPath), cachePath)
		if err != nil {
			return err
		}
		if err := os.Symlink(linkPath, workPath); err != nil {
			return err
		}
	}
	return nil
}

func checkoutDir(
	ctx context.Context,
	ch LocalCache,
	workspaceDir string,
	art artifact.Artifact,
	strat strategy.CheckoutStrategy,
	activeSharedWorkers chan struct{},
	progress *pb.ProgressBar,
) error {
	status, cachePath, workPath, err := quickStatus(ch, workspaceDir, art)
	if err != nil {
		return err
	}
	cachePath = filepath.Join(ch.dir, cachePath)
	if !status.HasChecksum {
		return InvalidChecksumError{art.Checksum}
	}
	if !status.ChecksumInCache {
		return MissingFromCacheError{art.Checksum}
	}
	if !(status.WorkspaceFileStatus == fsutil.StatusAbsent ||
		status.WorkspaceFileStatus == fsutil.StatusDirectory) {
		return fmt.Errorf(
			"expected target to be empty or a directory, found %s",
			status.WorkspaceFileStatus,
		)
	}
	man, err := readDirManifest(cachePath)
	if err != nil {
		return err
	}

	// When linking, the progress report counts files linked. Add all of the
	// files we know about here to the total, and let checkoutFile handle
	// updating the report. (When copying, checkoutFile handles updating the
	// bytes transferred completely.)
	if strat == strategy.LinkStrategy {
		fileCount := 0
		for _, art := range man.Contents {
			if !art.IsDir {
				fileCount++
			}
		}
		progress.AddTotal(int64(fileCount))
		progress.Start()
	}

	// Start a goroutine to feed artifacts to workers.
	errGroup, groupCtx := errgroup.WithContext(ctx)
	childArtifacts := make(chan *artifact.Artifact)
	errGroup.Go(func() error {
		for _, childArt := range man.Contents {
			select {
			case childArtifacts <- childArt:
			case <-groupCtx.Done():
				return groupCtx.Err()
			}
		}
		close(childArtifacts)
		return nil
	})

	// Start workers to checkout artifacts.
	activeDedicatedWorkers := make(chan struct{}, maxDedicatedWorkers)
	for i := 0; i < len(man.Contents); i++ {
		select {
		case <-groupCtx.Done():
			break
		case activeSharedWorkers <- struct{}{}:
			errGroup.Go(func() error {
				defer func() { <-activeSharedWorkers }()
				return checkoutWorker(
					groupCtx,
					ch,
					workPath,
					childArtifacts,
					strat,
					activeSharedWorkers,
					progress,
				)
			})
		case activeDedicatedWorkers <- struct{}{}:
			errGroup.Go(func() error {
				defer func() { <-activeDedicatedWorkers }()
				return checkoutWorker(
					groupCtx,
					ch,
					workPath,
					childArtifacts,
					strat,
					activeSharedWorkers,
					progress,
				)
			})
		}
	}
	// Wait for all goroutines to exit and collect the group error.
	return errGroup.Wait()
}

func checkoutWorker(
	ctx context.Context,
	ch LocalCache,
	workPath string,
	input <-chan *artifact.Artifact,
	strat strategy.CheckoutStrategy,
	activeSharedWorkers chan struct{},
	progress *pb.ProgressBar,
) error {
	for {
		select {
		case childArt, ok := <-input:
			if !ok {
				return nil
			}
			if childArt.IsDir {
				if err := checkoutDir(
					ctx,
					ch,
					workPath,
					*childArt,
					strat,
					activeSharedWorkers,
					progress,
				); err != nil {
					return err
				}
			} else {
				if err := checkoutFile(ch, workPath, *childArt, strat, progress); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
