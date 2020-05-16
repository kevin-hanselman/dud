package fsutil

import (
	"github.com/go-yaml/yaml"
	"os"
)

// ToYamlFile saves a struct to a YAML file.
func ToYamlFile(path string, v interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return yaml.NewEncoder(file).Encode(v)
}

// FromYamlFile loads a struct from a YAML file.
func FromYamlFile(path string, v interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	return yaml.NewDecoder(file).Decode(v)
}
