package video

import (
	"fmt"
	"os"

	"github.com/meunomeebero/ffmpego/internal/ffutil"
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
	info *Info
}

// New creates a new Video instance from a file path.
// Returns an error if ffmpeg/ffprobe are not installed or the file does not exist.
func New(path string) (*Video, error) {
	if err := ffutil.CheckDependencies(); err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("video file not accessible: %s: %w", path, err)
	}
	return &Video{path: path}, nil
}

// Path returns the video file path
func (v *Video) Path() string {
	return v.path
}
