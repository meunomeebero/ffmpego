package video

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Segment represents a time-based segment of video
type Segment struct {
	StartTime float64
	EndTime   float64
	Duration  float64
}

// ExtractSegment extracts a segment from the video file
func (v *Video) ExtractSegment(outputPath string, startTime, endTime float64, config *ConvertConfig) error {
	// Ensure output directory exists
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get video info for quality preservation
	info, err := v.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	// Build FFmpeg command
	args := []string{
		"-i", v.path,
		"-ss", fmt.Sprintf("%.3f", startTime),
		"-to", fmt.Sprintf("%.3f", endTime),
	}

	// Apply configuration or use defaults
	if config != nil {
		args = append(args, buildConvertArgs(info, config)...)
	} else {
		// Use defaults to preserve quality
		args = append(args, buildDefaultArgs(info)...)
	}

	// Add output path
	args = append(args, "-y", outputPath)

	// Execute FFmpeg command
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg error: %w - %s", err, string(output))
	}

	return nil
}

// ConcatenateSegments concatenates multiple video segment files into a single video
func ConcatenateSegments(segmentPaths []string, outputPath string, config *ConvertConfig) error {
	if len(segmentPaths) == 0 {
		return fmt.Errorf("no segments to concatenate")
	}

	// Get info from first segment for quality preservation
	firstVideo, err := New(segmentPaths[0])
	if err != nil {
		return fmt.Errorf("failed to open first segment: %w", err)
	}

	info, err := firstVideo.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	// Create temporary file list
	tempDir := filepath.Dir(segmentPaths[0])
	fileListPath := filepath.Join(tempDir, "segments_list.txt")

	fileList, err := os.Create(fileListPath)
	if err != nil {
		return fmt.Errorf("failed to create file list: %w", err)
	}
	defer os.Remove(fileListPath)

	// Write segment paths to file list
	for _, segmentPath := range segmentPaths {
		absSegmentPath, err := filepath.Abs(segmentPath)
		if err != nil {
			fmt.Printf("Warning: could not get absolute path for %s: %v\n", segmentPath, err)
			continue
		}

		if _, err := os.Stat(absSegmentPath); os.IsNotExist(err) {
			fmt.Printf("Warning: segment file does not exist: %s\n", absSegmentPath)
			continue
		}

		fileList.WriteString(fmt.Sprintf("file '%s'\n", absSegmentPath))
	}
	fileList.Close()

	// Check if the file list is empty
	fileInfo, err := os.Stat(fileListPath)
	if err != nil {
		return fmt.Errorf("failed to stat file list '%s': %w", fileListPath, err)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("no valid segments found to concatenate")
	}

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command for concatenation
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", fileListPath,
	}

	// Apply configuration or use copy for speed
	if config != nil && config.needsReencoding(info) {
		args = append(args, buildConvertArgs(info, config)...)
	} else {
		// Just copy streams without re-encoding
		args = append(args, "-c", "copy")
	}

	// Add output path
	args = append(args, "-y", outputPath)

	// Execute FFmpeg command
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to concatenate segments: %w - %s", err, string(output))
	}

	return nil
}
