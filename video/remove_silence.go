package video

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/meunomeebero/ffmpego/internal/ffutil"
)

// RemoveSilence detects and removes silent parts from the video, producing a new
// video with only the non-silent segments concatenated together.
// Uses parallel extraction for speed. Video streams are copied without re-encoding;
// audio gets a short fade at each cut point to prevent clicks and pops.
func (v *Video) RemoveSilence(outputPath string, config SilenceConfig) error {
	segments, err := v.GetNonSilentSegments(config)
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

	ext := filepath.Ext(v.path)
	if ext == "" {
		ext = ".mp4"
	}

	// Extract segments in parallel
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
				errs[i] = v.extractSegmentWithAudioFade(path, seg.StartTime, seg.EndTime)
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

// extractSegmentWithAudioFade extracts a video segment keeping the video stream as-is
// (stream copy) while re-encoding audio with a short fade-in/fade-out at the boundaries.
//
// This is used exclusively by RemoveSilence. When segments are cut and later concatenated,
// the audio waveform at each cut point is unlikely to be at a zero-crossing, which produces
// audible clicks. The fade eliminates these artifacts without affecting video quality or
// significantly increasing processing time (only the audio track is re-encoded).
func (v *Video) extractSegmentWithAudioFade(outputPath string, startTime, endTime float64) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	duration := endTime - startTime
	fadeFilter := ffutil.AudioFadeFilter(duration, ffutil.DefaultFadeDurationSec)

	args := []string{
		"-ss", fmt.Sprintf("%.3f", startTime),
		"-i", v.path,
		"-t", fmt.Sprintf("%.3f", duration),
		"-c:v", "copy",
		"-af", fadeFilter,
		"-y", outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg error: %w - %s", err, string(output))
	}
	return nil
}
