package cache

import (
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
func (cache *LocalCache) Commit(workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error {
	args := commitArgs{
		WorkingDir: workingDir,
		Cache:      cache,
		Artifact:   art,
		Strategy:   strat,
	}
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
	dstFile, err := ioutil.TempFile(args.Cache.Dir(), "")
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	defer dstFile.Close()

	// TODO: only copy if the cache is on a different filesystem (os.Rename if possible)
	// OR, if we're using CopyStrategy
	checksum, err := checksum.Checksum(io.TeeReader(srcFile, dstFile), 0)
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	cachePath, err := args.Cache.PathForChecksum(checksum)
	if err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	dstDir := filepath.Dir(cachePath)
	if err = os.MkdirAll(dstDir, 0755); err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	if err = os.Rename(dstFile.Name(), cachePath); err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	if err := os.Chmod(cachePath, 0444); err != nil {
		return errors.Wrap(err, errorPrefix)
	}
	args.Artifact.Checksum = checksum
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

var writeDirManifest = func(manPath string, manifest *directoryManifest) error {
	if err := os.MkdirAll(filepath.Dir(manPath), 0755); err != nil {
		return errors.Wrap(err, "writeDirManifest")
	}
	file, err := os.OpenFile(manPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		if os.IsExist(err) { // If the file already exists, trust the cache and carry on.
			return nil
		}
		return errors.Wrap(err, "writeDirManifest")
	}
	return json.NewEncoder(file).Encode(manifest)
}

func commitDirArtifact(args commitArgs) error {
	baseDir := filepath.Join(args.WorkingDir, args.Artifact.Path)
	entries, err := readDir(baseDir)
	if err != nil {
		return errors.Wrap(err, "commitDir")
	}
	manifest := directoryManifest{Path: baseDir}
	for _, entry := range entries {
		if !entry.Mode().IsDir() { // TODO: only proceed for reg files and links
			fileArt := artifact.Artifact{Path: entry.Name()}
			args := commitArgs{
				Cache:      args.Cache,
				WorkingDir: baseDir,
				Artifact:   &fileArt,
				Strategy:   args.Strategy,
			}
			if err := commitFileArtifact(args); err != nil {
				return errors.Wrap(err, "commitDir")
			}
			manifest.Contents = append(manifest.Contents, &fileArt)
		}
	}
	manChecksum, err := checksum.ChecksumObject(manifest)
	if err != nil {
		return errors.Wrap(err, "commitDir")
	}
	args.Artifact.Checksum = manChecksum
	path, err := args.Cache.PathForChecksum(manChecksum)
	if err != nil {
		return errors.Wrap(err, "commitDir")
	}
	if err := writeDirManifest(path, &manifest); err != nil {
		return errors.Wrap(err, "commitDir")
	}
	return nil
}
