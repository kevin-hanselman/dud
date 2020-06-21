package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/checksum"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Commit calculates the checksum of the artifact, moves it to the cache, then performs a checkout.
func (ch *LocalCache) Commit(
	workingDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	// TODO: improve error reporting? avoid recursive wrapping
	if art.IsDir {
		return commitDirArtifact(context.Background(), ch, workingDir, art, strat)
	}
	return commitFileArtifact(ch, workingDir, art, strat)
}

var readDir = ioutil.ReadDir

var commitFileArtifact = func(
	ch *LocalCache,
	workingDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	// ignore cachePath because the artifact likely has a stale or empty checksum
	status, _, workPath, err := quickStatus(ch, workingDir, *art)
	errorPrefix := fmt.Sprintf("commit %#v", workPath)
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	if status.WorkspaceFileStatus == fsutil.Absent {
		return errors.Wrapf(os.ErrNotExist, "commitFile: %#v does not exist", workPath)
	}
	if status.WorkspaceFileStatus != fsutil.RegularFile {
		return fmt.Errorf("%s: not a regular file", errorPrefix)
	}
	srcFile, err := os.Open(workPath)
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	defer srcFile.Close()

	sameFs, err := fsutil.SameFilesystem(workPath, ch.Dir())
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	moveFile := ""
	if sameFs && strat == strategy.LinkStrategy {
		moveFile = workPath
	}

	cksum, err := ch.commitBytes(srcFile, moveFile)
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}

	art.Checksum = cksum
	// There's no need to call Checkout if using CopyStrategy; the original file still exists.
	if strat == strategy.LinkStrategy {
		// TODO: add "force" option to cache.Checkout to replace this
		exists, err := fsutil.Exists(workPath, false)
		if err != nil {
			return errors.Wrap(err, errorPrefix)
		}
		if exists {
			if err := os.Remove(workPath); err != nil {
				return errors.Wrap(err, errorPrefix)
			}
		}
		return ch.Checkout(workingDir, art, strat)
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

	cksum, err := checksum.Checksum(reader, 0)
	if err != nil {
		return "", err
	}
	cachePath, err := ch.PathForChecksum(cksum)
	if err != nil {
		return "", err
	}
	dstDir := filepath.Dir(cachePath)
	if err = os.MkdirAll(dstDir, 0755); err != nil {
		return "", err
	}
	if err = os.Rename(moveFile, cachePath); err != nil {
		return "", err
	}
	if err := os.Chmod(cachePath, 0444); err != nil {
		return "", err
	}
	return cksum, nil
}

var commitDirManifest = func(ch *LocalCache, manifest *directoryManifest) (string, error) {
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
	workingDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
) error {
	baseDir := filepath.Join(workingDir, art.Path)
	entries, err := readDir(baseDir)
	if err != nil {
		return err
	}
	errGroup, childCtx := errgroup.WithContext(ctx)
	childArtChan := make(chan *artifact.Artifact, len(entries))
	manifest := directoryManifest{Path: baseDir}
	for i := range entries {
		// This verbose declaration of entry is necessary to avoid capturing
		// loop variables in the closure below.
		// See: https://eli.thegreenplace.net/2019/go-internals-capturing-loop-variables-in-closures/
		entry := entries[i]
		errGroup.Go(func() error {
			childArt := artifact.Artifact{Path: entry.Name()}
			if entry.IsDir() {
				if !art.IsRecursive {
					return nil
				}
				childArt.IsDir = true
				childArt.IsRecursive = true
				if err := commitDirArtifact(childCtx, ch, baseDir, &childArt, strat); err != nil {
					return err
				}
			} else { // TODO: ensure regular file or symlink
				if err := commitFileArtifact(ch, baseDir, &childArt, strat); err != nil {
					return err
				}
			}
			childArtChan <- &childArt
			return nil
		})
	}
	if err := errGroup.Wait(); err != nil {
		return err
	}
	close(childArtChan)
	for childArt := range childArtChan {
		manifest.Contents = append(manifest.Contents, childArt)
	}
	cksum, err := commitDirManifest(ch, &manifest)
	if err != nil {
		return err
	}
	art.Checksum = cksum
	return nil
}
