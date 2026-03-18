package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/meunomeebero/ffmpego/internal/ffutil"
)

// Segment represents a time-based segment of audio
type Segment = ffutil.Segment

// ExtractSegment extracts a segment from the audio file.
// Pass nil for config to use stream copy (fastest, no quality loss).
func (a *Audio) ExtractSegment(outputPath string, startTime, endTime float64, config *ConvertConfig) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// -ss before -i enables fast input seeking (keyframe-level)
	args := []string{
		"-ss", fmt.Sprintf("%.3f", startTime),
		"-i", a.path,
		"-t", fmt.Sprintf("%.3f", endTime-startTime),
	}

	if config != nil {
		args = append(args, buildConvertArgs(config)...)
	} else {
		args = append(args, "-c", "copy")
	}

	args = append(args, "-y", outputPath)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg error: %w - %s", err, string(output))
	}
	return nil
}

// ConcatenateSegments concatenates multiple audio segment files into a single audio.
// Pass nil for config to use stream copy (fastest, no quality loss).
func ConcatenateSegments(segmentPaths []string, outputPath string, config *ConvertConfig) error {
	if len(segmentPaths) == 0 {
		return fmt.Errorf("no segments to concatenate")
	}

	fileList, err := os.CreateTemp("", "audio_segments_list_*.txt")
	if err != nil {
		return fmt.Errorf("failed to create file list: %w", err)
	}
	fileListPath := fileList.Name()
	defer fileList.Close()
	defer os.Remove(fileListPath)

	for _, segmentPath := range segmentPaths {
		absPath, err := filepath.Abs(segmentPath)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", segmentPath, err)
		}
		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("segment file not accessible: %s: %w", absPath, err)
		}
		if strings.ContainsAny(absPath, "\n\r") {
			return fmt.Errorf("segment path contains invalid characters: %s", segmentPath)
		}
		escaped := strings.ReplaceAll(absPath, "'", "'\\''")
		if _, err := fmt.Fprintf(fileList, "file '%s'\n", escaped); err != nil {
			return fmt.Errorf("failed to write file list: %w", err)
		}
	}

	if err := fileList.Close(); err != nil {
		return fmt.Errorf("failed to close file list: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", fileListPath,
	}

	if config != nil {
		args = append(args, buildConvertArgs(config)...)
	} else {
		args = append(args, "-c", "copy")
	}

	args = append(args, "-y", outputPath)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to concatenate segments: %w - %s", err, string(output))
	}
	return nil
}
