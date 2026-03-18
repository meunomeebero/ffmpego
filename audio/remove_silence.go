package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// RemoveSilence detects and removes silent parts from the audio, producing a new
// audio file with only the non-silent segments concatenated together.
// Uses parallel extraction and stream copy to preserve original quality.
func (a *Audio) RemoveSilence(outputPath string, config SilenceConfig) error {
	segments, err := a.GetNonSilentSegments(config)
	if err != nil {
		return fmt.Errorf("failed to detect segments: %w", err)
	}

	if len(segments) == 0 {
		return fmt.Errorf("no audible content found above the configured threshold")
	}

	tempDir, err := os.MkdirTemp("", "ffmpego_silence_*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	ext := filepath.Ext(a.path)
	if ext == "" {
		ext = ".mp3"
	}

	segmentPaths := make([]string, len(segments))
	errs := make([]error, len(segments))

	maxWorkers := 4
	if len(segments) < maxWorkers {
		maxWorkers = len(segments)
	}

	var wg sync.WaitGroup
	jobs := make(chan int, len(segments))

	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				seg := segments[i]
				path := filepath.Join(tempDir, fmt.Sprintf("seg_%03d%s", i, ext))
				segmentPaths[i] = path
				errs[i] = a.ExtractSegment(path, seg.StartTime, seg.EndTime, nil)
			}
		}()
	}

	for i := range segments {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			return fmt.Errorf("failed to extract segment %d: %w", i+1, err)
		}
	}

	return ConcatenateSegments(segmentPaths, outputPath, nil)
}
