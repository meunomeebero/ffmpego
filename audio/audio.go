package audio

import (
	"fmt"
	"os"

	"github.com/meunomeebero/ffmpego/internal/ffutil"
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
	info *Info
}

// New creates a new Audio instance from a file path.
// Returns an error if ffmpeg/ffprobe are not installed or the file does not exist.
func New(path string) (*Audio, error) {
	if err := ffutil.CheckDependencies(); err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("audio file not accessible: %s: %w", path, err)
	}
	return &Audio{path: path}, nil
}

// Path returns the audio file path
func (a *Audio) Path() string {
	return a.path
}
