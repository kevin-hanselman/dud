package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/checksum"
	"github.com/kevlar1818/duc/fsutil"
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
	srcPath := path.Join(args.WorkingDir, args.Artifact.Path)
	isRegFile, err := fsutil.IsRegularFile(srcPath)
	if err != nil {
		return errors.Wrap(err, "commitFile")
	}
	if !isRegFile {
		return fmt.Errorf("commitFile: path %v is not a regular file", srcPath)
	}
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return errors.Wrap(err, "commitFile")
	}
	defer srcFile.Close()
	dstFile, err := ioutil.TempFile(args.Cache.Dir, "")
	if err != nil {
		return errors.Wrap(err, "commitFile")
	}
	defer dstFile.Close()

	// TODO: only copy if the cache is on a different filesystem (os.Rename if possible)
	// OR, if we're using CopyStrategy
	checksum, err := checksum.Checksum(io.TeeReader(srcFile, dstFile), 0)
	if err != nil {
		return errors.Wrap(err, "commitFile")
	}
	// TODO: remove usage of checksum slices -- leave this logic to PathForChecksum
	dstDir := path.Join(args.Cache.Dir, checksum[:2])
	if err = os.MkdirAll(dstDir, 0755); err != nil {
		return errors.Wrap(err, "commitFile")
	}
	cachePath := path.Join(dstDir, checksum[2:])
	if err = os.Rename(dstFile.Name(), cachePath); err != nil {
		return errors.Wrap(err, "commitFile")
	}
	if err := os.Chmod(cachePath, 0444); err != nil {
		return errors.Wrap(err, "commitFile")
	}
	args.Artifact.Checksum = checksum
	// There's no need to call Checkout if using CopyStrategy; the original file still exists.
	if args.Strategy == strategy.LinkStrategy {
		// TODO: add rm to checkout as "force" option
		if err := os.Remove(srcPath); err != nil {
			return errors.Wrap(err, "commitFile")
		}
		return args.Cache.Checkout(args.WorkingDir, args.Artifact, args.Strategy)
	}
	return nil
}

var writeDirManifest = func(manPath string, manifest *directoryManifest) error {
	if err := os.MkdirAll(path.Dir(manPath), 0755); err != nil {
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
	baseDir := path.Join(args.WorkingDir, args.Artifact.Path)
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
	// TODO: Dir manifests don't need to store the checksum internally. The
	// checksum should be stored in the artifact.
	if err := checksum.Update(&manifest); err != nil {
		return errors.Wrap(err, "commitDir")
	}
	args.Artifact.Checksum = manifest.Checksum
	path, err := args.Cache.PathForChecksum(manifest.Checksum)
	if err != nil {
		return errors.Wrap(err, "commitDir")
	}
	if err := writeDirManifest(path, &manifest); err != nil {
		return errors.Wrap(err, "commitDir")
	}
	return nil
}
