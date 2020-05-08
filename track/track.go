package track

import (
	"fmt"
	"github.com/kevlar1818/duc/artifact"
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/stage"
)

var fileStatusFromPath = fsutil.FileStatusFromPath

// Track creates a stage for tracking the given paths.
func Track(paths ...string) (stage.Stage, error) {
	outputs := make([]artifact.Artifact, len(paths))
	var stg stage.Stage
	for i, path := range paths {
		status, err := fileStatusFromPath(path)
		if err != nil {
			return stg, err
		}
		if status == artifact.Absent {
			return stg, fmt.Errorf("path %v does not exist", path)
		}
		if status == artifact.Other {
			return stg, fmt.Errorf("unsupported file type for path %v", path)
		}
		outputs[i] = artifact.Artifact{
			Checksum: "",
			Path:     path,
			IsDir:    status == artifact.Directory,
		}
	}
	stg = stage.Stage{
		Outputs: outputs,
	}
	return stg, nil
}
