package cache

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/progress"
	"github.com/pkg/errors"
)

// Push uploads an Artifact from the local cache to a remote cache.
//
// This uses a map of Artifacts instead of a slice to ease both testing and
// calling code. Primarily, a Stage's outputs will be passed to this function,
// so it's convenient to pass stage.Outputs directly. This also eases testing,
// because transcribing the map into a slice would introduce non-determinism.
func (ch LocalCache) Push(
	remoteDst string,
	arts map[string]*artifact.Artifact,
	logger *agglog.AggLogger,
) error {
	logger.Info.Print("Gathering files to push")
	progress := progress.NewProgress(false, "  ")
	progress.Start()
	pushFiles := make(map[string]struct{})
	for _, art := range arts {
		if err := gatherFilesToPush(ch, *art, pushFiles, progress); err != nil {
			progress.Finish()
			return errors.Wrapf(err, "push %s", art.Path)
		}
	}
	progress.Finish()
	if len(pushFiles) > 0 {
		return errors.Wrap(remoteCopy(ch.dir, remoteDst, pushFiles, logger), "push")
	}
	return nil
}

func gatherFilesToPush(
	ch LocalCache,
	art artifact.Artifact,
	filesToPush map[string]struct{},
	progress progress.Progress,
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
			if err := gatherFilesToPush(ch, *childArt, filesToPush, progress); err != nil {
				return err
			}
		}
	}
	filesToPush[cachePath] = struct{}{}
	progress.DoneFile()
	return nil
}

var remoteCopy = func(
	src,
	dst string,
	fileSet map[string]struct{},
	logger *agglog.AggLogger,
) error {
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
	return setFilePerms(dst, fileSet, cacheFilePerms, logger)
}

func setFilePerms(
	commonDir string,
	fileSet map[string]struct{},
	mode fs.FileMode,
	logger *agglog.AggLogger,
) error {
	numFiles := len(fileSet)
	logger.Info.Print("Fixing permissions")
	progress := progress.NewProgress(false, "  ")
	progress.AddFiles(numFiles)
	progress.Start()
	defer progress.Finish()

	// If there's a small number of files don't bother with concurrency.
	if numFiles < maxSharedWorkers {
		var chmodErr error = nil
		for file := range fileSet {
			err := os.Chmod(filepath.Join(commonDir, file), mode)
			if err == nil || os.IsNotExist(err) {
				progress.DoneFile()
			} else {
				chmodErr = err
			}
		}
		return chmodErr
	}

	errs := make(chan error, numFiles)
	fileChan := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for file := range fileSet {
			fileChan <- file
		}
		close(fileChan)
	}()
	for i := 0; i < maxSharedWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				err := os.Chmod(filepath.Join(commonDir, file), mode)
				// TODO: Consider exiting early on "no such file" errors; this
				// likely means the remote is truly remote.
				if err == nil || os.IsNotExist(err) {
					progress.DoneFile()
				} else {
					errs <- err
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	// Return the first error reported and ignore the rest. If there were no
	// errors, because this is a buffered channel, we should receive the zero
	// value, nil.
	return <-errs
}
