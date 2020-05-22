package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/checksum"
	"github.com/kevlar1818/duc/strategy"
	"github.com/pkg/errors"
)

// Commit calculates the checksum of the artifact, moves it to the cache, then performs a checkout.
func (cache *LocalCache) Commit(
	workingDir string,
	art *artifact.Artifact,
	strat strategy.CheckoutStrategy,
	recursive bool,
) error {
	args := commitArgs{
		WorkingDir: workingDir,
		Cache:      cache,
		Artifact:   art,
		Strategy:   strat,
		Recursive:  recursive,
	}
	// TODO: improve error reporting? avoid recursive wrapping
	if art.IsDir {
		return commitDirArtifact(args)
	}
	return commitFileArtifact(args)
}

var readDir = ioutil.ReadDir

type commitArgs struct {
	Cache      *LocalCache
	WorkingDir string
	Artifact   *artifact.Artifact
	Strategy   strategy.CheckoutStrategy
	Recursive  bool
}

var commitFileArtifact = func(args commitArgs) error {
	// ignore cachePath because the artifact likely has a stale or empty checksum
	status, _, workPath, err := quickStatus(args.Cache, args.WorkingDir, *args.Artifact)
	errorPrefix := fmt.Sprintf("commit %#v", workPath)
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	if status.WorkspaceFileStatus == artifact.Absent {
		return errors.Wrapf(os.ErrNotExist, "commitFile: %#v does not exist", workPath)
	}
	if status.WorkspaceFileStatus != artifact.RegularFile {
		return fmt.Errorf("%s: not a regular file", errorPrefix)
	}
	srcFile, err := os.Open(workPath)
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	defer srcFile.Close()

	cksum, err := args.Cache.commitBytes(srcFile)

	args.Artifact.Checksum = cksum
	// There's no need to call Checkout if using CopyStrategy; the original file still exists.
	if args.Strategy == strategy.LinkStrategy {
		// TODO: add rm to checkout as "force" option
		if err := os.Remove(workPath); err != nil {
			return errors.Wrap(err, errorPrefix)
		}
		return args.Cache.Checkout(args.WorkingDir, args.Artifact, args.Strategy)
	}
	return nil
}

// commitBytes checksums and writes the bytes from reader to the cache.
func (cache *LocalCache) commitBytes(reader io.Reader) (string, error) {
	dstFile, err := ioutil.TempFile(cache.dir, "")
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	// TODO: only copy if the cache is on a different filesystem (os.Rename if possible)
	// OR, if we're using CopyStrategy
	cksum, err := checksum.Checksum(io.TeeReader(reader, dstFile), 0)
	if err != nil {
		return "", err
	}
	cachePath, err := cache.PathForChecksum(cksum)
	if err != nil {
		return "", err
	}
	dstDir := filepath.Dir(cachePath)
	if err = os.MkdirAll(dstDir, 0755); err != nil {
		return "", err
	}
	if err = os.Rename(dstFile.Name(), cachePath); err != nil {
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
	return ch.commitBytes(buf)
}

func commitDirArtifact(args commitArgs) error {
	baseDir := filepath.Join(args.WorkingDir, args.Artifact.Path)
	entries, err := readDir(baseDir)
	if err != nil {
		return err
	}
	manifest := directoryManifest{Path: baseDir}
	for _, entry := range entries {
		childArt := artifact.Artifact{Path: entry.Name()}
		args := commitArgs{
			Cache:      args.Cache,
			WorkingDir: baseDir,
			Strategy:   args.Strategy,
			Recursive:  args.Recursive,
			Artifact:   &childArt,
		}
		if entry.IsDir() {
			if !args.Recursive {
				continue
			}
			childArt.IsDir = true
			if err := commitDirArtifact(args); err != nil {
				return err
			}
		} else { // TODO: ensure regular file or symlink
			if err := commitFileArtifact(args); err != nil {
				return err
			}
		}
		manifest.Contents = append(manifest.Contents, &childArt)
	}
	cksum, err := commitDirManifest(args.Cache, &manifest)
	if err != nil {
		return err
	}
	args.Artifact.Checksum = cksum
	return nil
}
