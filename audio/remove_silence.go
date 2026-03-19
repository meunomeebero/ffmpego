package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/meunomeebero/ffmpego/internal/ffutil"
)

// RemoveSilence detects and removes silent parts from the audio, producing a new
// audio file with only the non-silent segments concatenated together.
// Uses parallel extraction. Each segment gets a short audio fade at its boundaries
// to prevent clicks and pops when concatenated.
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
				errs[i] = a.extractSegmentWithAudioFade(path, seg.StartTime, seg.EndTime)
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

// extractSegmentWithAudioFade extracts an audio segment with a short fade-in/fade-out
// at the boundaries.
//
// This is used exclusively by RemoveSilence. When audio is cut at arbitrary points,
// the waveform rarely lands on a zero-crossing, which produces audible clicks after
// concatenation. The fade smooths these transitions without perceptibly affecting volume.
func (a *Audio) extractSegmentWithAudioFade(outputPath string, startTime, endTime float64) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	duration := endTime - startTime
	fadeFilter := ffutil.AudioFadeFilter(duration, ffutil.DefaultFadeDurationSec)

	args := []string{
		"-ss", fmt.Sprintf("%.3f", startTime),
		"-i", a.path,
		"-t", fmt.Sprintf("%.3f", duration),
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
