package video

import (
	"fmt"
	"os"
)

// Video represents a video file with fluent API
type Video struct {
	path string
}

// New creates a new Video instance from a file path
func New(path string) (*Video, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("video file does not exist: %s", path)
	}
	return &Video{path: path}, nil
}

// Path returns the video file path
func (v *Video) Path() string {
	return v.path
}
