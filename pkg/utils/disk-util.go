package utils

import (
	"github.com/peterbourgon/diskv"
)

func SimpleTransform(key string) []string {
	// Simplest transform function: put all the data files into the base dir.
	return []string{}
}

func DiskvStore(basePath string) *diskv.Diskv {
	return diskv.New(diskv.Options{
		BasePath:  basePath,
		Transform: SimpleTransform,
	})
}

type DiskvRecorder struct {
	Store                     *diskv.Diskv
	RecordingBufferDataPoints int
	RecordingChan             chan bool
}

func CreateDiskvRecorder(basePath string, recordingBufferDataPoints int, recordingChan chan bool) DiskvRecorder {
	return DiskvRecorder{
		Store:                     DiskvStore(basePath),
		RecordingBufferDataPoints: recordingBufferDataPoints,
		RecordingChan:             recordingChan,
	}
}
