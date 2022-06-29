package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/checksum"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Commit calculates the checksum of the artifact, moves it to the cache, then
// performs a checkout.
func (ch LocalCache) Commit(
	workspaceDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
	logger *agglog.AggLogger,
) (err error) {
	if err := os.MkdirAll(ch.dir, 0o755); err != nil {
		return errors.Wrapf(err, "commit %s", art.Path)
	}
	// Try to move a dummy file between the workspace and the cache. If we can
	// move files (via rename syscall), we can avoid writing to disk
	// for file commits, dramatically improving performance.
	canRenameFile, err := canRenameFileBetweenDirs(workspaceDir, ch.dir)
	if err != nil {
		return errors.Wrapf(err, "commit %s", art.Path)
	}
	progress := newProgress(art.Path)
	progress.Start()
	defer progress.Finish()
	if art.IsDir {
		activeSharedWorkers := make(chan struct{}, maxSharedWorkers)
		err = commitDirArtifact(
			context.Background(),
			ch,
			workspaceDir,
			art,
			strat,
			activeSharedWorkers,
			progress,
			canRenameFile,
		)
	} else {
		err = commitFileArtifact(ch, workspaceDir, art, strat, progress, canRenameFile)
	}
	if err == nil && progress.Current() <= 0 {
		progress.SetTemplate(progressSkipCommitTemplate)
	}
	return errors.Wrapf(err, "commit %s", art.Path)
}

var canRenameFileBetweenDirs = func(srcDir, dstDir string) (bool, error) {
	// Touch a file in each directory.
	srcFile, err := os.CreateTemp(srcDir, "")
	if err != nil {
		return false, err
	}
	if err := srcFile.Close(); err != nil {
		return false, err
	}
	dstFile, err := os.CreateTemp(dstDir, "")
	if err != nil {
		return false, err
	}
	if err := dstFile.Close(); err != nil {
		return false, err
	}

	// Attempt the rename.
	renameErr := os.Rename(srcFile.Name(), dstFile.Name())

	// Cleanup the temp files.
	if err := os.Remove(srcFile.Name()); err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if err := os.Remove(dstFile.Name()); err != nil {
		return false, err
	}
	return renameErr == nil, nil
}

func commitFileArtifact(
	ch LocalCache,
	workspaceDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
	progress *pb.ProgressBar,
	canRenameFile bool,
) error {
	// Ignore cachePath because the artifact likely has a stale or empty checksum.
	status, _, workPath, err := quickStatus(ch, workspaceDir, *art)
	if err != nil {
		return err
	}
	if status.WorkspaceFileStatus == fsutil.StatusAbsent {
		return errors.Wrap(os.ErrNotExist, workPath)
	}
	if status.ContentsMatch {
		return nil
	}
	if status.WorkspaceFileStatus != fsutil.StatusRegularFile {
		return errors.New("not a regular file")
	}
	fileInfo, err := os.Stat(workPath)
	if err != nil {
		return err
	}
	progress.AddTotal(fileInfo.Size())
	srcFile, err := os.Open(workPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	srcReader := progress.NewProxyReader(srcFile)

	if art.SkipCache {
		cksum, err := checksum.Checksum(srcReader)
		if err != nil {
			return err
		}
		art.Checksum = cksum
		return nil
	}

	moveFile := ""
	if canRenameFile && strat == strategy.LinkStrategy {
		moveFile = workPath
	}

	cksum, err := ch.commitBytes(srcReader, moveFile)
	if err != nil {
		return err
	}

	art.Checksum = cksum
	// There's no need to call Checkout if using CopyStrategy; the original
	// file still exists.
	if strat == strategy.LinkStrategy {
		// If we can't rename the file then we copied it, and we need to remove
		// it before linking.
		if !canRenameFile {
			if err := os.Remove(workPath); err != nil {
				return err
			}
		}
		// Purposefully avoid cache.Checkout here as we don't need or want the
		// overhead of managing a progress bar.
		return checkoutFile(ch, workspaceDir, *art, strat, nil)
	}
	return nil
}

// commitBytes checksums the bytes from reader and results in said bytes being
// present in the cache. If moveFile is empty, commitBytes will copy from
// reader to the cache while checksumming. If moveFile is not empty, the file
// path it references is moved (i.e. renamed) to the cache after checksumming,
// thus eliminating unnecessary file IO.
func (ch LocalCache) commitBytes(reader io.Reader, moveFile string) (string, error) {
	// If there's no file we can move, we need to copy the bytes from reader to
	// the cache.
	if moveFile == "" {
		tempFile, err := os.CreateTemp(ch.dir, "")
		if err != nil {
			return "", err
		}
		defer tempFile.Close()
		reader = io.TeeReader(reader, tempFile)
		moveFile = tempFile.Name()
	}

	cksum, err := checksum.Checksum(reader)
	if err != nil {
		return "", err
	}
	cachePath, err := ch.PathForChecksum(cksum)
	if err != nil {
		return "", err
	}
	cachePath = filepath.Join(ch.dir, cachePath)
	dstDir := filepath.Dir(cachePath)
	if err = os.MkdirAll(dstDir, 0o755); err != nil {
		return "", err
	}
	// This rename may race others, but luckily we don't care who wins the
	// race. Everyone in the race is trying to put the same exact file in the
	// cache (because of content-addressed storage), so the outcome is the same
	// no matter who wins the race. At the OS level rename is atomic, so
	// there's no risk of corrupting the destination file with multiple
	// concurrent syscalls. (This is at least true for UNIX, but that's all we
	// support. See also: https://github.com/golang/go/issues/8914)
	if err = os.Rename(moveFile, cachePath); err != nil {
		return "", err
	}
	if err := os.Chmod(cachePath, cacheFilePerms); err != nil {
		return "", err
	}
	return cksum, nil
}

func commitDirManifest(ch LocalCache, manifest *directoryManifest) (string, error) {
	// TODO: Consider using an io.Pipe() instead of a buffer.
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(manifest); err != nil {
		return "", err
	}
	return ch.commitBytes(buf, "")
}

func commitDirArtifact(
	ctx context.Context,
	ch LocalCache,
	workspaceDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
	activeSharedWorkers chan struct{},
	progress *pb.ProgressBar,
	canRenameFile bool,
) error {
	status, cachePath, workPath, err := quickStatus(ch, workspaceDir, *art)
	if err != nil {
		return err
	}

	var oldManifest directoryManifest
	if status.ChecksumInCache {
		oldManifest, err = readDirManifest(filepath.Join(ch.dir, cachePath))
		if err != nil {
			return err
		}
	}

	entries, err := readDir(workPath, art.DisableRecursion)
	if err != nil {
		return err
	}

	// Start a goroutine to feed files/sub-directories to workers.
	errGroup, groupCtx := errgroup.WithContext(ctx)
	inputFiles := make(chan os.DirEntry)
	errGroup.Go(func() error {
		defer close(inputFiles)
		for _, entry := range entries {
			select {
			case inputFiles <- entry:
			case <-groupCtx.Done():
				return groupCtx.Err()
			}
		}
		return nil
	})

	childArtifacts := make(chan *artifact.Artifact)
	manifestReady := make(chan struct{})

	// Start a goroutine to build the directory manifest from committed
	// artifacts.
	newManifest := &directoryManifest{
		Path:     art.Path,
		Contents: make(map[string]*artifact.Artifact),
	}
	errGroup.Go(func() error {
		// There should be exactly len(entries) Artifacts returned in the
		// childArtifacts channel. This fact is critical for enabling the
		// dynamic worker scheduling below, because that logic needs to know
		// when to stop waiting for available worker tokens (via the
		// manifestReady channel).
		for i := 0; i < len(entries); i++ {
			select {
			case childArt := <-childArtifacts:
				newManifest.Contents[childArt.Path] = childArt
			case <-groupCtx.Done():
				return groupCtx.Err()
			}
		}
		close(manifestReady)
		return nil
	})

	startCommitWorkers(
		groupCtx,
		errGroup,
		ch,
		workPath,
		oldManifest,
		strat,
		len(entries),
		inputFiles,
		childArtifacts,
		manifestReady,
		activeSharedWorkers,
		progress,
		canRenameFile,
	)

	// Wait for all goroutines to exit and collect the group error.
	if err := errGroup.Wait(); err != nil {
		return err
	}

	close(childArtifacts)

	cksum, err := commitDirManifest(ch, newManifest)
	if err != nil {
		return err
	}
	art.Checksum = cksum
	return nil
}

// Start workers to commit artifacts. We spawn workers when there's free
// space in either of the "active worker" channels. We quit when we've
// either scheduled as many workers as files/sub-dirs, the manifest builder
// says the manifest is ready, or the group was cancelled.
func startCommitWorkers(
	ctx context.Context,
	errGroup *errgroup.Group,
	ch LocalCache,
	workPath string,
	oldManifest directoryManifest,
	strat strategy.CheckoutStrategy,
	totalWorkItems int,
	inputFiles <-chan os.DirEntry,
	outputArtifacts chan<- *artifact.Artifact,
	manifestReady chan struct{},
	activeSharedWorkers chan struct{},
	progress *pb.ProgressBar,
	canRenameFile bool,
) {
	activeDedicatedWorkers := make(chan struct{}, maxDedicatedWorkers)
	for i := 0; i < totalWorkItems; i++ {
		select {
		case <-ctx.Done():
			return
		case <-manifestReady:
			return
		case activeSharedWorkers <- struct{}{}:
			errGroup.Go(func() error {
				defer func() { <-activeSharedWorkers }()
				return commitWorker(
					ctx,
					ch,
					workPath,
					oldManifest,
					strat,
					inputFiles,
					outputArtifacts,
					activeSharedWorkers,
					progress,
					canRenameFile,
				)
			})
		case activeDedicatedWorkers <- struct{}{}:
			errGroup.Go(func() error {
				defer func() { <-activeDedicatedWorkers }()
				return commitWorker(
					ctx,
					ch,
					workPath,
					oldManifest,
					strat,
					inputFiles,
					outputArtifacts,
					activeSharedWorkers,
					progress,
					canRenameFile,
				)
			})
		}
	}
}

func commitWorker(
	ctx context.Context,
	ch LocalCache,
	workPath string,
	dirMan directoryManifest,
	strat strategy.CheckoutStrategy,
	inputFiles <-chan os.DirEntry,
	outputArtifacts chan<- *artifact.Artifact,
	activeSharedWorkers chan struct{},
	progress *pb.ProgressBar,
	canRenameFile bool,
) error {
	for entry := range inputFiles {
		path := entry.Name()
		var (
			childArt *artifact.Artifact
			err      error
		)
		// See if we can recover a child artifact from an existing directory
		// manifest. This enables skipping up-to-date artifacts.
		childArt, ok := dirMan.Contents[path]
		if !ok {
			childArt = &artifact.Artifact{
				Path:  path,
				IsDir: entry.IsDir(),
			}
		}
		if childArt.IsDir {
			err = commitDirArtifact(
				ctx,
				ch,
				workPath,
				childArt,
				strat,
				activeSharedWorkers,
				progress,
				canRenameFile,
			)
		} else {
			err = commitFileArtifact(
				ch,
				workPath,
				childArt,
				strat,
				progress,
				canRenameFile,
			)
		}
		if err != nil {
			return err
		}
		select {
		case outputArtifacts <- childArt:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func readDir(path string, excludeSubDirs bool) (out []os.DirEntry, err error) {
	dir, err := os.Open(path)
	if err != nil {
		return
	}
	defer dir.Close()
	out, err = dir.ReadDir(0)
	if err != nil {
		return
	}

	if excludeSubDirs {
		allOut := out // TODO: refactor to avoid this potential copy
		out = make([]os.DirEntry, 0, len(allOut))
		for _, entry := range allOut {
			if entry.IsDir() {
				continue
			}
			out = append(out, entry)
		}
	}
	return
}
