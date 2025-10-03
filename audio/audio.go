package audio

import (
	"fmt"
	"os"
)

// Audio represents an audio file with fluent API
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
