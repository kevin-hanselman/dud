package track

import (
	"github.com/kevlar1818/duc/stage"
	"os"
)

var fileExists = func(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return err
	}
	return nil
}

// Track creates a stage for tracking the given paths.
func Track(paths ...string) (s stage.Stage, err error) {
	outputs := make([]stage.Artifact, len(paths))
	for i, path := range paths {
		if err = fileExists(path); err != nil {
			return
		}
		outputs[i] = stage.Artifact{
			Checksum: "",
			Path:     path,
		}
	}
	s = stage.Stage{
		Checksum: "",
		Outputs:  outputs,
	}
	return
}
