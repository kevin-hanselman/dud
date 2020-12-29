package cache

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/artifact"
)

// Push uploads an Artifact from the local cache to a remote cache.
// TODO: Consider removing the workspaceDir argument. Technically Push and
// Fetch shouldn't care about the workspace at all; they purely interact with
// the local cache.
func (ch *LocalCache) Push(workspaceDir, remoteDst string, art artifact.Artifact) error {
	fetchFiles := make(map[string]struct{})
	if err := gatherFilesToPush(ch, workspaceDir, art, fetchFiles); err != nil {
		return err
	}
	if len(fetchFiles) > 0 {
		return remoteCopy(ch.dir, remoteDst, fetchFiles)
	}
	return nil
}

func gatherFilesToPush(
	ch *LocalCache,
	workspaceDir string,
	art artifact.Artifact,
	filesToPush map[string]struct{},
) error {
	if art.SkipCache {
		return nil
	}
	status, cachePath, _, err := quickStatus(ch, workspaceDir, art)
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
		childWorkspaceDir := filepath.Join(workspaceDir, art.Path)
		for _, childArt := range man.Contents {
			if err := gatherFilesToPush(ch, childWorkspaceDir, *childArt, filesToPush); err != nil {
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
		"--immutable", // TODO: Assess this flag further.
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

	// Ensure any local files that were created end up as read-only.
	// Try to chmod all files, ignoring "no such file" errors which are
	// probably due to the destination being remote.
	// TODO: Consider making this a standalone function and only call it for
	// fetch, not push. You could argue that we still want to do this on push
	// because someone may use a local directory as their "remote".
	var chmodErr error
	for file := range fileSet {
		err := os.Chmod(filepath.Join(dst, file), cacheFilePerms)
		if err != nil && !os.IsNotExist(err) {
			chmodErr = err
		}
	}
	return chmodErr
}
