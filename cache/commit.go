package cache

import (
	"fmt"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/strategy"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
)

var readDir = ioutil.ReadDir

type commitFileArtifactArgs struct {
	Cache      *LocalCache
	WorkingDir string
	Artifact   *artifact.Artifact
	Strategy   strategy.CheckoutStrategy
}

// defined as a private var to enable mocking
var commitFileArtifact = func(args commitFileArtifactArgs) error {
	srcPath := path.Join(args.WorkingDir, args.Artifact.Path)
	isRegFile, err := fsutil.IsRegularFile(srcPath)
	if err != nil {
		return err
	}
	if !isRegFile {
		return fmt.Errorf("file %#v is not a regular file", srcPath)
	}
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := ioutil.TempFile(args.Cache.Dir, "")
	if err != nil {
		return errors.Wrapf(err, "creating tempfile in %#v failed", args.Cache.Dir)
	}
	defer dstFile.Close()

	// TODO: only copy if the cache is on a different filesystem (os.Rename if possible)
	// OR, if we're using CopyStrategy
	checksum, err := fsutil.ChecksumAndCopy(srcFile, dstFile)
	if err != nil {
		return errors.Wrapf(err, "checksum of %#v failed", srcPath)
	}
	dstDir := path.Join(args.Cache.Dir, checksum[:2])
	if err = os.MkdirAll(dstDir, 0755); err != nil {
		return errors.Wrapf(err, "mkdirs %#v failed", dstDir)
	}
	cachePath := path.Join(dstDir, checksum[2:])
	if err = os.Rename(dstFile.Name(), cachePath); err != nil {
		return errors.Wrapf(err, "mv %#v failed", dstFile)
	}
	if err := os.Chmod(cachePath, 0444); err != nil {
		return errors.Wrapf(err, "chmod %#v failed", cachePath)
	}
	args.Artifact.Checksum = checksum
	// There's no need to call Checkout if using CopyStrategy; the original file still exists.
	if args.Strategy == strategy.LinkStrategy {
		// TODO: add rm to checkout as "force" option
		if err := os.Remove(srcPath); err != nil {
			return errors.Wrapf(err, "rm %#v failed", srcPath)
		}
		return args.Cache.Checkout(args.WorkingDir, args.Artifact, args.Strategy)
	}
	return nil
}

// Commit calculates the checksum of the artifact, moves it to the cache, then performs a checkout.
func (cache *LocalCache) Commit(workingDir string, art *artifact.Artifact, strat strategy.CheckoutStrategy) error {
	if art.IsDir {
		newWorkingDir := path.Join(workingDir, art.Path)
		entries, err := readDir(newWorkingDir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !entry.Mode().IsDir() {
				fileArt := artifact.Artifact{Path: entry.Name()}
				args := commitFileArtifactArgs{
					Cache:      cache,
					WorkingDir: newWorkingDir,
					Artifact:   &fileArt,
					Strategy:   strat,
				}
				if err := commitFileArtifact(args); err != nil {
					return err
				}
			}
		}
	} else {
		args := commitFileArtifactArgs{
			Cache:      cache,
			WorkingDir: workingDir,
			Artifact:   art,
			Strategy:   strat,
		}
		return commitFileArtifact(args)
	}
	return nil
}
