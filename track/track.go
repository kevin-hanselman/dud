package track

import (
	"github.com/kevlar1818/duc/fsutil"
	"github.com/kevlar1818/duc/stage"
)

var fileExists func(string) (bool, error) = fsutil.Exists

// Track creates a stage for tracking the given paths.
func Track(paths ...string) (s stage.Stage, err error) {
	outputs := make([]stage.Artifact, len(paths))
	for i, path := range paths {
		var exists bool
		exists, err = fileExists(path)
		if (!exists) || err != nil {
			return
		}
		outputs[i] = stage.Artifact{
			Checksum: "",
			Path:     path,
		}
	}
	s = stage.Stage{
		Outputs: outputs,
	}
	return
}
