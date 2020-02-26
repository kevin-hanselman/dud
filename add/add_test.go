package add

import (
	"bytes"
	"testing"
)

func TestAdd(t *testing.T) {
	inputString := "Hello, World!"
	inputReader := bytes.NewBufferString(inputString)
	want := "0a0a9f2a6772942557ab5355d76af442f8f65e01"
	output, err := Add(inputReader)
	if err != nil {
		t.Error(err)
	}
	if output != want {
		t.Errorf("Add(%v) = %s, want %s", inputString, output, want)
	}
}
