package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/checksum"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/strategy"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/exp/mmap"
	"golang.org/x/sync/errgroup"
)

const (
	cacheFilePerms = 0444
)

// These are somewhat arbitrary numbers. We need to profile more.
var (
	// The number of concurrent workers available to a top-level directory
	// artifact and all its child artifacts.
	maxSharedWorkers = 64
	// The number of concurrent workers available to each individual directory
	// artifact (and not its children). Dedicated workers are necessary because
	// without them, deadlocks can occur when maxSharedWorkers is less than the
	// directory depth.
	maxDedicatedWorkers = 1
)

// Commit calculates the checksum of the artifact, moves it to the cache, then
// performs a checkout.
func (ch *LocalCache) Commit(
	workspaceDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	pb := progressbar.DefaultBytes(-1, fmt.Sprintf("committing %s", art.Path))
	defer pb.Finish()
	if art.IsDir {
		activeSharedWorkers := make(chan struct{}, maxSharedWorkers)
		return commitDirArtifact(
			context.Background(),
			ch,
			workspaceDir,
			art,
			strat,
			activeSharedWorkers,
			pb,
		)
	}
	return commitFileArtifact(ch, workspaceDir, art, strat, pb)
}

func memoryMapOpen(path string) (reader io.Reader, closer io.Closer, err error) {
	readerAt, err := mmap.Open(path)
	if err != nil {
		return
	}
	reader = io.NewSectionReader(readerAt, 0, math.MaxInt64)
	closer = readerAt
	return
}

func commitFileArtifact(
	ch *LocalCache,
	workspaceDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
	pb *progressbar.ProgressBar,
) error {
	// ignore cachePath because the artifact likely has a stale or empty checksum
	status, _, workPath, err := quickStatus(ch, workspaceDir, *art)
	errorPrefix := fmt.Sprintf("commit file %s", workPath)
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	if status.WorkspaceFileStatus == fsutil.StatusAbsent {
		return errors.Wrap(errors.Wrap(os.ErrNotExist, workPath), errorPrefix)
	}
	if status.ContentsMatch {
		return nil
	}
	if status.WorkspaceFileStatus != fsutil.StatusRegularFile {
		return errors.Wrap(errors.New("not a regular file"), errorPrefix)
	}
	//srcReader, srcCloser, err := memoryMapOpen(workPath)
	srcFile, err := os.Open(workPath)
	var srcReader io.Reader = io.TeeReader(srcFile, pb)
	var srcCloser io.Closer = srcFile
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	defer srcCloser.Close()

	if art.SkipCache {
		cksum, err := checksum.Checksum(srcReader)
		if err != nil {
			return errors.Wrap(err, errorPrefix)
		}
		art.Checksum = cksum
		return nil
	}

	sameFs, err := fsutil.SameFilesystem(workPath, ch.Dir())
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	moveFile := ""
	if sameFs && strat == strategy.LinkStrategy {
		moveFile = workPath
	}

	cksum, err := ch.commitBytes(srcReader, moveFile)
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}

	art.Checksum = cksum
	// There's no need to call Checkout if using CopyStrategy; the original
	// file still exists.
	if strat == strategy.LinkStrategy {
		return ch.Checkout(workspaceDir, *art, strat)
	}
	return nil
}

// commitBytes checksums the bytes from reader and results in said bytes being
// present in the cache. If moveFile is empty, commitBytes will copy from
// reader to the cache while checksumming. If moveFile is not empty, the file
// path it references is moved (i.e. renamed) to the cache after checksumming,
// thus eliminating unnecessary file IO.
func (ch *LocalCache) commitBytes(reader io.Reader, moveFile string) (string, error) {
	// If there's no file we can move, we need to copy the bytes from reader to
	// the cache.
	if moveFile == "" {
		tempFile, err := ioutil.TempFile(ch.dir, "")
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
	if err = os.MkdirAll(dstDir, 0755); err != nil {
		return "", err
	}
	if err = os.Rename(moveFile, cachePath); err != nil {
		return "", err
	}
	if err := os.Chmod(cachePath, cacheFilePerms); err != nil {
		return "", err
	}
	return cksum, nil
}

func commitDirManifest(ch *LocalCache, manifest *directoryManifest) (string, error) {
	// TODO: Consider using an io.Pipe() instead of a buffer.
	// For large directories this is probably more important.
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(manifest); err != nil {
		return "", err
	}
	return ch.commitBytes(buf, "")
}

func commitDirArtifact(
	ctx context.Context,
	ch *LocalCache,
	workspaceDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
	activeSharedWorkers chan struct{},
	pb *progressbar.ProgressBar,
) error {
	// TODO: don't bother checking if regular files are up-to-date?
	status, oldManifest, err := dirArtifactStatus(ch, workspaceDir, *art)
	if err != nil {
		return err
	}
	if status.ContentsMatch {
		return nil
	}

	baseDir := filepath.Join(workspaceDir, art.Path)

	infos, err := readDir(baseDir, art.DisableRecursion)

	// Start a goroutine to feed files/sub-directories to workers.
	errGroup, groupCtx := errgroup.WithContext(ctx)
	inputFiles := make(chan os.FileInfo)
	errGroup.Go(func() error {
		defer close(inputFiles)
		for _, info := range infos {
			select {
			case inputFiles <- info:
			case <-groupCtx.Done():
				return groupCtx.Err()
			}
		}
		return nil
	})

	activeDedicatedWorkers := make(chan struct{}, maxDedicatedWorkers)
	childArtifacts := make(chan *artifact.Artifact)
	manifestReady := make(chan struct{})

	newManifest := &directoryManifest{Path: art.Path}
	newManifest.Contents = make(map[string]*artifact.Artifact)
	errGroup.Go(func() error {
		// There should be exactly len(infos) Artifacts returned in the
		// childArtifacts channel. This fact is critical for enabling the
		// dynamic worker scheduling below, because that logic needs to know
		// when to stop waiting for available worker tokens (via the
		// manifestReady channel).
		for i := 0; i < len(infos); i++ {
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

	// Start workers to commit child Artifacts. We spawn workers when there's
	// free space in either of the "active worker" channels. We quit when
	// either we've either scheduled as many workers as files/sub-dirs, the
	// manifest builder says the manifest is ready, or the group was cancelled.
	for i := 0; i < len(infos); i++ {
		select {
		case <-groupCtx.Done():
		case <-manifestReady:
			break
		case activeSharedWorkers <- struct{}{}:
			errGroup.Go(func() error {
				defer func() { <-activeSharedWorkers }()
				return dirWorker(
					groupCtx,
					ch,
					*art,
					baseDir,
					oldManifest,
					strat,
					inputFiles,
					childArtifacts,
					activeSharedWorkers,
					pb,
				)
			})
		case activeDedicatedWorkers <- struct{}{}:
			errGroup.Go(func() error {
				defer func() { <-activeDedicatedWorkers }()
				return dirWorker(
					groupCtx,
					ch,
					*art,
					baseDir,
					oldManifest,
					strat,
					inputFiles,
					childArtifacts,
					activeSharedWorkers,
					pb,
				)
			})
		}
	}

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

func dirWorker(
	ctx context.Context,
	ch *LocalCache,
	art artifact.Artifact,
	baseDir string,
	dirMan directoryManifest,
	strat strategy.CheckoutStrategy,
	inputFiles <-chan os.FileInfo,
	outputArtifacts chan<- *artifact.Artifact,
	activeSharedWorkers chan struct{},
	pb *progressbar.ProgressBar,
) error {
	for info := range inputFiles {
		path := info.Name()
		// See if we can recover a sub-artifact from an existing
		// dirManifest. This enables skipping up-to-date artifacts.
		var childArt *artifact.Artifact
		childArt, ok := dirMan.Contents[path]
		if !ok {
			childArt = &artifact.Artifact{Path: path}
		}
		if info.IsDir() {
			childArt.IsDir = true
			if err := commitDirArtifact(
				ctx,
				ch,
				baseDir,
				childArt,
				strat,
				activeSharedWorkers,
				pb,
			); err != nil {
				return err
			}
		} else {
			if err := commitFileArtifact(ch, baseDir, childArt, strat, pb); err != nil {
				return err
			}
		}
		select {
		case outputArtifacts <- childArt:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func readDir(path string, excludeDirs bool) (out []os.FileInfo, err error) {
	dir, err := os.Open(path)
	defer dir.Close()
	if err != nil {
		return
	}
	out, err = dir.Readdir(0)
	if err != nil {
		return
	}

	if excludeDirs {
		allOut := out
		out = make([]os.FileInfo, 0, len(allOut))
		for _, info := range allOut {
			if info.IsDir() {
				continue
			}
			out = append(out, info)
		}
	}
	return
}
