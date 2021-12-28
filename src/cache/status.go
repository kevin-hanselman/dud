package cache

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/checksum"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Status reports the status of an Artifact in the Cache. If shortCircuit is
// true, Status will exit as soon as the overall state of the Artifact is known
// -- saving time, but potentially leaving artifact.Status.ChildrenStatus
// incomplete. If shortCircuit is false, Status will fully populate
// artifact.Status.ChildrenStatus.
func (ch LocalCache) Status(workspaceDir string, art artifact.Artifact, shortCircuit bool) (
	status artifact.Status,
	err error,
) {
	if art.IsDir {
		activeSharedWorkers := make(chan struct{}, maxSharedWorkers)
		status, err = dirArtifactStatus(
			context.Background(),
			ch,
			workspaceDir,
			art,
			shortCircuit,
			activeSharedWorkers,
		)
	} else {
		status, err = fileArtifactStatus(ch, workspaceDir, art)
	}
	err = errors.Wrapf(err, "status %s", art.Path)
	return
}

// checksumStatus populates the HasChecksum and ChecksumInCache fields of
// artifact.Status and returns any relevant cache file information.
func checksumStatus(ch LocalCache, art artifact.Artifact) (
	status artifact.Status,
	cachePath string,
	cacheFileInfo fs.FileInfo,
	err error,
) {
	cachePath, err = ch.PathForChecksum(art.Checksum)
	if _, ok := err.(InvalidChecksumError); ok {
		err = nil
		status.HasChecksum = false
		return
	}
	if err != nil {
		return
	}
	status.HasChecksum = true
	cacheFileInfo, err = os.Stat(filepath.Join(ch.dir, cachePath))
	if err == nil {
		status.ChecksumInCache = true
	} else if os.IsNotExist(err) {
		err = nil
		status.ChecksumInCache = false
	}
	return
}

// quickStatus populates all artifact.Status fields except for ContentsMatch
// and ChildrenStatus. However, this function will set ContentsMatch if the
// Artifact is a file, the workspace file is a link, and the other status
// booleans are true. Checking to see if a link points to the cache is, as this
// function suggests, quick.
var quickStatus = func(
	ch LocalCache,
	workspaceDir string,
	art artifact.Artifact,
) (status artifact.Status, cachePath, workPath string, err error) {
	// These FileInfos are used to verify a committed file is correctly linked
	// to the cache.
	var cacheFileInfo, workFileInfo fs.FileInfo

	status, cachePath, cacheFileInfo, err = checksumStatus(ch, art)
	if err != nil {
		return
	}
	status.Artifact = art

	workPath = filepath.Join(workspaceDir, art.Path)
	status.WorkspaceFileStatus, err = fsutil.FileStatusFromPath(workPath)
	if err != nil {
		return
	}
	if status.HasChecksum &&
		status.ChecksumInCache &&
		status.WorkspaceFileStatus == fsutil.StatusLink {
		workFileInfo, err = os.Stat(workPath)
		// A NotExist error here means the link is dead. Leave ContentsMatch as
		// false and let the caller handle the invalid link.
		if os.IsNotExist(err) {
			err = nil
			return
		} else if err != nil {
			return
		}
		status.ContentsMatch = os.SameFile(cacheFileInfo, workFileInfo)
	}
	return
}

func fileArtifactStatus(
	ch LocalCache,
	workspaceDir string,
	art artifact.Artifact,
) (artifact.Status, error) {
	status, cachePath, workPath, err := quickStatus(ch, workspaceDir, art)
	if err != nil {
		return status, err
	}
	cachePath = filepath.Join(ch.dir, cachePath)

	if status.WorkspaceFileStatus != fsutil.StatusRegularFile {
		return status, nil
	}

	if art.SkipCache {
		if !status.HasChecksum {
			return status, nil
		}
		fileReader, err := os.Open(workPath)
		if err != nil {
			return status, err
		}
		defer fileReader.Close()
		workspaceFileChecksum, err := checksum.Checksum(fileReader)
		if err != nil {
			return status, err
		}
		status.ContentsMatch = workspaceFileChecksum == art.Checksum
	} else {
		if !status.ChecksumInCache {
			return status, nil
		}
		status.ContentsMatch, err = fsutil.SameContents(workPath, cachePath)
		if err != nil {
			return status, err
		}
	}
	return status, nil
}

func dirArtifactStatus(
	ctx context.Context,
	ch LocalCache,
	workspaceDir string,
	art artifact.Artifact,
	shortCircuit bool,
	activeSharedWorkers chan struct{},
) (artifact.Status, error) {
	status, cachePath, workPath, err := quickStatus(ch, workspaceDir, art)
	if err != nil {
		return status, err
	}
	cachePath = filepath.Join(ch.dir, cachePath)

	if !(status.HasChecksum && status.ChecksumInCache) && shortCircuit {
		return status, nil
	}

	if status.WorkspaceFileStatus != fsutil.StatusDirectory {
		return status, nil
	}

	// Any child Artifact that's out-of-date or not committed will flip this
	// value.
	status.ContentsMatch = true
	status.ChildrenStatus = make(map[string]*artifact.Status)

	var manifest directoryManifest
	// First, ensure all artifacts in the directoryManifest are up-to-date.
	if status.ChecksumInCache {
		manifest, err = readDirManifest(cachePath)
		if err != nil {
			return status, err
		}

		children := make([]*artifact.Artifact, len(manifest.Contents))
		i := 0
		for _, art := range manifest.Contents {
			children[i] = art
			i++
		}

		err = concurrentStatus(
			ctx,
			ch,
			workPath,
			children,
			shortCircuit,
			activeSharedWorkers,
			&status,
		)
		if err != nil {
			return status, err
		}
		if shortCircuit && !status.ContentsMatch {
			return status, nil
		}
	}

	// Second, get a directory listing and check for untracked files.
	entries, err := readDir(workPath, art.DisableRecursion)
	if err != nil {
		return status, err
	}
	children := make([]*artifact.Artifact, 0, len(entries))

	for _, entry := range entries {
		newArt := artifact.Artifact{Path: entry.Name(), IsDir: entry.IsDir()}
		// Ignore all entries in the manifest; we've already checked them
		// above. (While assigning to a nil map panics, accessing a nil map is
		// safe.)
		if _, ok := manifest.Contents[newArt.Path]; ok {
			continue
		}
		children = append(children, &newArt)
	}
	if len(children) == 0 {
		return status, nil
	}
	// After the directory listing is filtered above, the presence of any
	// untracked files or directories means this Artifact is out-of-date.
	// Therefore we can exit early if needed.
	status.ContentsMatch = false
	if shortCircuit {
		return status, nil
	}
	err = concurrentStatus(
		ctx,
		ch,
		workPath,
		children,
		shortCircuit, // This will always be false due to the check above.
		activeSharedWorkers,
		&status,
	)
	return status, err
}

type shortCircuited struct{}

func (c shortCircuited) Error() string {
	return "short-circuited"
}

func concurrentStatus(
	ctx context.Context,
	ch LocalCache,
	workspaceDir string,
	children []*artifact.Artifact,
	shortCircuit bool,
	activeSharedWorkers chan struct{},
	status *artifact.Status,
) error {
	work := make(chan *artifact.Artifact)
	results := make(chan *artifact.Status)
	statusReady := make(chan struct{})

	// Start a goroutine to feed workers.
	errGroup, groupCtx := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		defer close(work)
		for _, child := range children {
			select {
			case work <- child:
			case <-groupCtx.Done():
				return groupCtx.Err()
			}
		}
		return nil
	})

	// Start a goroutine to collect statuses of children.
	errGroup.Go(func() error {
		for i := 0; i < len(children); i++ {
			select {
			case childStatus := <-results:
				status.ChildrenStatus[childStatus.Path] = childStatus
				status.ContentsMatch = status.ContentsMatch && childStatus.ContentsMatch
				if shortCircuit && !status.ContentsMatch {
					return shortCircuited{}
				}
			case <-groupCtx.Done():
				return groupCtx.Err()
			}
		}
		close(statusReady)
		return nil
	})

	startStatusWorkers(
		groupCtx,
		errGroup,
		ch,
		workspaceDir,
		shortCircuit,
		len(children),
		activeSharedWorkers,
		statusReady,
		work,
		results,
	)

	// Wait for all goroutines to exit and collect the group error.
	err := errGroup.Wait()
	if _, ok := err.(shortCircuited); ok {
		err = nil
	}
	return err
}

// Start workers when there's free space in either of the "active worker"
// channels. We quit when we've either scheduled as many workers as
// child artifacts, the status builder says the parent status is ready, or
// the group was cancelled.
func startStatusWorkers(
	ctx context.Context,
	errGroup *errgroup.Group,
	ch LocalCache,
	workspaceDir string,
	shortCircuit bool,
	totalWorkItems int,
	activeSharedWorkers chan struct{},
	statusReady chan struct{},
	work <-chan *artifact.Artifact,
	out chan<- *artifact.Status,
) {
	activeDedicatedWorkers := make(chan struct{}, maxDedicatedWorkers)
	for i := 0; i < totalWorkItems; i++ {
		select {
		case <-ctx.Done():
			return
		case <-statusReady:
			return
		case activeSharedWorkers <- struct{}{}:
			errGroup.Go(func() error {
				defer func() { <-activeSharedWorkers }()
				return statusWorker(
					ctx,
					ch,
					workspaceDir,
					shortCircuit,
					activeSharedWorkers,
					work,
					out,
				)
			})
		case activeDedicatedWorkers <- struct{}{}:
			errGroup.Go(func() error {
				defer func() { <-activeDedicatedWorkers }()
				return statusWorker(
					ctx,
					ch,
					workspaceDir,
					shortCircuit,
					activeSharedWorkers,
					work,
					out,
				)
			})
		}
	}
}

func statusWorker(
	ctx context.Context,
	ch LocalCache,
	workspaceDir string,
	shortCircuit bool,
	activeSharedWorkers chan struct{},
	work <-chan *artifact.Artifact,
	out chan<- *artifact.Status,
) (err error) {
	for art := range work {
		// It's important to declare this var inside the loop so we get a fresh
		// object each iteration. If it's declared outside the loop, we'll
		// continually overwrite the same struct and cause races with the
		// consumers of the channel. Alternatively, we could make the output
		// channel carry actual status structs, not pointers.
		var st artifact.Status
		if art.IsDir {
			st, err = dirArtifactStatus(
				ctx,
				ch,
				workspaceDir,
				*art,
				shortCircuit,
				activeSharedWorkers,
			)
		} else {
			st, err = fileArtifactStatus(ch, workspaceDir, *art)
		}
		if err != nil {
			return err
		}
		select {
		case out <- &st:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
