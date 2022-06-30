package cache

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/pkg/errors"
)

// Push uploads an Artifact from the local cache to a remote cache.
//
// This uses a map of Artifacts instead of a slice to ease both testing and
// calling code. Primarily, a Stage's outputs will be passed to this function,
// so it's convenient to pass stage.Outputs directly. This also eases testing,
// because transcribing the map into a slice would introduce non-determinism.
func (ch LocalCache) Push(remoteDst string, arts map[string]*artifact.Artifact) error {
	pushFiles := make(map[string]struct{})
	// TODO: add a progress bar while gathering files
	for _, art := range arts {
		if err := gatherFilesToPush(ch, *art, pushFiles); err != nil {
			return errors.Wrapf(err, "push %s", art.Path)
		}
	}
	if len(pushFiles) > 0 {
		return errors.Wrap(remoteCopy(ch.dir, remoteDst, pushFiles), "push")
	}
	return nil
}

func gatherFilesToPush(
	ch LocalCache,
	art artifact.Artifact,
	filesToPush map[string]struct{},
) error {
	if art.SkipCache {
		return nil
	}
	status, cachePath, _, err := checksumStatus(ch, art)
	if err != nil {
		return err
	}
	if !status.HasChecksum {
		return InvalidChecksumError{art.Checksum}
	}
	if !status.ChecksumInCache {
		return MissingFromCacheError{art.Checksum}
	}
	if art.IsDir {
		man, err := readDirManifest(filepath.Join(ch.dir, cachePath))
		if err != nil {
			return err
		}
		for _, childArt := range man.Contents {
			if err := gatherFilesToPush(ch, *childArt, filesToPush); err != nil {
				return err
			}
		}
	}
	filesToPush[cachePath] = struct{}{}
	return nil
}

var remoteCopy = func(src, dst string, fileSet map[string]struct{}) error {
	cmd := exec.Command(
		"rclone",
		"--config",
		".dud/rclone.conf",
		// Ideally these sorts of flags could be added to the rclone config,
		// but I haven't found a way to add them.
		// See: https://github.com/rclone/rclone/issues/2697
		"--progress",
		"--immutable",
		// If file modification times change locally, without "--size-only",
		// rclone will error-out because of the "--immutable" flag above.
		"--size-only",
		"copy",
		// "--files-from -" means to get the list of files to copy from STDIN.
		"--files-from",
		"-",
		src,
		dst,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		for file := range fileSet {
			// We can ignore errors here because cmd.Wait() will return an
			// error on any I/O failures.
			fmt.Fprintln(stdin, file)
		}
	}()

	if err := cmd.Wait(); err != nil {
		return err
	}

	// Ensure any local files that were created end up as read-only. Try to
	// chmod all files, ignoring "no such file" errors which are probably due
	// to the destination being remote. This is important even for push,
	// because the "remote" might be a local directory.
	// TODO: make this concurrent and add a progress bar
	var chmodErr error
	for file := range fileSet {
		// Try correcting all the files and return the last error seen.
		err := os.Chmod(filepath.Join(dst, file), cacheFilePerms)
		if err != nil && !os.IsNotExist(err) {
			chmodErr = err
		}
	}
	return chmodErr
}
