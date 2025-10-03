package audio

import (
	"fmt"
	"os"
)

// Audio represents an audio file with fluent API.
//
// Thread-safety: Audio instances are not thread-safe. Each instance should be
// used by a single goroutine at a time. If you need to process the same audio
// file concurrently, create separate Audio instances for each goroutine.
// However, temporary files are uniquely named to prevent conflicts between
// concurrent operations on different Audio instances.
type Audio struct {
	path string
}

// New creates a new Audio instance from a file path
func New(path string) (*Audio, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("audio file does not exist: %s", path)
	}
	return &Audio{path: path}, nil
}

// Path returns the audio file path
func (a *Audio) Path() string {
	return a.path
}
