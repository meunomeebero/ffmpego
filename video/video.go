package video

import (
	"fmt"
	"os"
)

// Video represents a video file with fluent API.
//
// Thread-safety: Video instances are not thread-safe. Each instance should be
// used by a single goroutine at a time. If you need to process the same video
// file concurrently, create separate Video instances for each goroutine.
// However, temporary files are uniquely named to prevent conflicts between
// concurrent operations on different Video instances.
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
