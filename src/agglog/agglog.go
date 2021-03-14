package agglog

import (
	"io/ioutil"
	"log"
)

// AggLogger is an aggregate logger composed of multiple levels.
type AggLogger struct {
	Error *log.Logger
	Info  *log.Logger
	Debug *log.Logger
}

// NewNullLogger returns an AggLogger that discards all logging requests.
func NewNullLogger() *AggLogger {
	return &AggLogger{
		Error: log.New(ioutil.Discard, "", 0),
		Info:  log.New(ioutil.Discard, "", 0),
		Debug: log.New(ioutil.Discard, "", 0),
	}
}
