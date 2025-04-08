package ffmpego

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
* Exported functions
 */

// GetAudioInfo retrieves information about an audio file
func GetAudioInfo(audioPath string) (*AudioInfo, error) {
	// Check if FFprobe is available
	_, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, fmt.Errorf("ffprobe not found in PATH: %w", err)
	}

	// Get audio stream information
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=sample_rate,channels,codec_name,bit_rate",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1",
		audioPath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get audio info: %w", err)
	}

	// Parse output
	info := &AudioInfo{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "sample_rate=") {
			sampleRateStr := strings.TrimPrefix(line, "sample_rate=")
			info.SampleRate, _ = strconv.Atoi(sampleRateStr)
		} else if strings.HasPrefix(line, "channels=") {
			channelsStr := strings.TrimPrefix(line, "channels=")
			info.Channels, _ = strconv.Atoi(channelsStr)
		} else if strings.HasPrefix(line, "codec_name=") {
			info.Codec = strings.TrimPrefix(line, "codec_name=")
		} else if strings.HasPrefix(line, "bit_rate=") {
			bitRateStr := strings.TrimPrefix(line, "bit_rate=")
			info.BitRate, _ = strconv.Atoi(bitRateStr)
		} else if strings.HasPrefix(line, "duration=") {
			durStr := strings.TrimPrefix(line, "duration=")
			info.Duration, _ = strconv.ParseFloat(durStr, 64)
		}
	}

	return info, nil
}

// RemoveAudioSilence processes an audio file by removing silent parts
func RemoveAudioSilence(audioPath, outputPath string, minSilenceLen int, silenceThresh int, config *AudioConfig, logger Logger) error {
	// Create temporary directories
	tempDir := filepath.Join(os.TempDir(), "audio_processor_"+time.Now().Format("20060102_150405"))
	segmentsDir := filepath.Join(tempDir, "segments")

	defer os.RemoveAll(tempDir) // Cleanup when done

	// Create directories
	if err := os.MkdirAll(segmentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create segments directory: %w", err)
	}

	// Step 1: Get audio information for preserving quality
	logger.Section("Analyzing Audio")
	logger.Step("Getting audio information")

	audioInfo, err := GetAudioInfo(audioPath)
	if err != nil {
		return fmt.Errorf("failed to get audio info: %w", err)
	}

	logger.Success("Audio info: %d Hz, %d channels, codec: %s",
		audioInfo.SampleRate, audioInfo.Channels, audioInfo.Codec)

	// Step 2: Detect silence in audio
	logger.Step("Detecting silence in audio")

	audioSegments, err := DetectNonSilentSegments(audioPath, minSilenceLen, silenceThresh)
	if err != nil {
		return fmt.Errorf("failed to detect silence: %w", err)
	}

	// Handle case when no silence is detected
	if len(audioSegments) == 0 {
		logger.Info("No silent segments detected in the audio")
		// Simply copy the input to output since no processing is needed
		err = copyAudioFile(audioPath, outputPath)
		if err != nil {
			return fmt.Errorf("failed to copy audio: %w", err)
		}
		logger.Success("Original audio copied to output path")
		return nil
	}

	logger.Success("Detected %d non-silent segments", len(audioSegments))

	// Step 3: Extract each audio segment
	logger.Section("Processing Audio Segments")
	logger.Info("Extracting %d audio segments", len(audioSegments))

	segmentPaths := make([]string, len(audioSegments))
	fileListPath := filepath.Join(tempDir, "segments.txt")
	fileList, err := os.Create(fileListPath)
	if err != nil {
		return fmt.Errorf("failed to create file list: %w", err)
	}
	defer fileList.Close()

	for i, segment := range audioSegments {
		segmentPath := filepath.Join(segmentsDir, fmt.Sprintf("segment_%03d.mp3", i+1))

		// Extract audio segment
		logger.Debug("Extracting segment %d (%.2fs to %.2fs)", i+1, segment.StartTime, segment.EndTime)

		args := []string{
			"-i", audioPath,
			"-ss", fmt.Sprintf("%.3f", segment.StartTime),
			"-to", fmt.Sprintf("%.3f", segment.EndTime),
		}

		// Apply audio configuration if provided
		if config != nil {
			// Use specified codec or original
			if config.Codec != "" {
				args = append(args, "-c:a", config.Codec)
			} else {
				args = append(args, "-c:a", audioInfo.Codec)
			}

			// Apply sample rate if specified
			if config.SampleRate > 0 {
				args = append(args, "-ar", strconv.Itoa(config.SampleRate))
			} else {
				args = append(args, "-ar", strconv.Itoa(audioInfo.SampleRate))
			}

			// Apply channels if specified
			if config.Channels > 0 {
				args = append(args, "-ac", strconv.Itoa(config.Channels))
			} else {
				args = append(args, "-ac", strconv.Itoa(audioInfo.Channels))
			}

			// Apply quality if specified
			if config.Quality > 0 {
				args = append(args, "-q:a", strconv.Itoa(config.Quality))
			}

			// Apply bitrate if specified
			if config.BitRate > 0 {
				args = append(args, "-b:a", fmt.Sprintf("%dk", config.BitRate))
			}
		} else {
			// Use default settings to preserve quality
			args = append(args,
				"-c:a", audioInfo.Codec,
				"-ar", strconv.Itoa(audioInfo.SampleRate),
				"-ac", strconv.Itoa(audioInfo.Channels))
		}

		// Add output path
		args = append(args, "-y", segmentPath)

		cmd := exec.Command("ffmpeg", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.Error("Failed to extract segment %d: %s - %s", i+1, err, string(output))
			continue
		}

		segmentPaths[i] = segmentPath
		fileList.WriteString(fmt.Sprintf("file '%s'\n", segmentPath))
		logger.Success("Segment %d/%d processed successfully", i+1, len(audioSegments))
	}

	fileList.Close()

	// Filter out any failed segments
	validSegmentCount := 0
	for _, path := range segmentPaths {
		if path != "" {
			validSegmentCount++
		}
	}

	if validSegmentCount == 0 {
		return fmt.Errorf("all segments failed to process")
	}

	// Step 4: Concatenate segments
	logger.Section("Creating Final Audio")
	logger.Step("Concatenating %d audio segments", validSegmentCount)

	// Create directory for output if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use FFmpeg's concat demuxer to combine the segments
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", fileListPath,
	}

	// Apply audio configuration if provided
	if config != nil {
		// Use specified codec or original
		if config.Codec != "" {
			args = append(args, "-c:a", config.Codec)
		} else {
			args = append(args, "-c:a", audioInfo.Codec)
		}

		// Apply sample rate if specified
		if config.SampleRate > 0 {
			args = append(args, "-ar", strconv.Itoa(config.SampleRate))
		} else {
			args = append(args, "-ar", strconv.Itoa(audioInfo.SampleRate))
		}

		// Apply channels if specified
		if config.Channels > 0 {
			args = append(args, "-ac", strconv.Itoa(config.Channels))
		} else {
			args = append(args, "-ac", strconv.Itoa(audioInfo.Channels))
		}

		// Apply quality if specified
		if config.Quality > 0 {
			args = append(args, "-q:a", strconv.Itoa(config.Quality))
		}

		// Apply bitrate if specified
		if config.BitRate > 0 {
			args = append(args, "-b:a", fmt.Sprintf("%dk", config.BitRate))
		}
	} else {
		// Use default settings to preserve quality
		args = append(args,
			"-c:a", audioInfo.Codec,
			"-ar", strconv.Itoa(audioInfo.SampleRate),
			"-ac", strconv.Itoa(audioInfo.Channels))
	}

	// Add output path
	args = append(args, "-y", outputPath)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to concatenate segments: %w - %s", err, string(output))
	}

	logger.Success("Audio processed successfully")
	logger.Info("Output saved to: %s", outputPath)

	return nil
}

// Extract audio from a video file and save it to a file
func ExtractAudioFromVideo(videoPath, outputPath string) error {
	// Check if FFmpeg is available
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}

	// Create output directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Extract audio from video
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-q:a", "2", "-y", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to extract audio: %w - %s", err, string(output))
	}

	return nil
}

// DetectNonSilentSegments detects silence in an audio file and returns the non-silent segments
func DetectNonSilentSegments(audioPath string, minSilenceLen int, silenceThresh int) ([]AudioSegment, error) {
	// Convert ms to seconds for FFmpeg
	silenceLenSec := float64(minSilenceLen) / 1000.0

	// Use FFmpeg's silencedetect filter to find silence periods
	cmd := exec.Command("ffmpeg",
		"-i", audioPath,
		"-af", fmt.Sprintf("silencedetect=noise=%ddB:d=%.3f", silenceThresh, silenceLenSec),
		"-f", "null", "-")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to detect silence: %w - %s", err, string(output))
	}

	// Extract silence_start and silence_end times
	silenceStarts, silenceEnds := parseSilenceOutput(string(output))

	// If no silence detected, return empty result
	if len(silenceStarts) == 0 || len(silenceEnds) == 0 {
		return []AudioSegment{}, nil
	}

	// Get total audio duration
	audioInfo, err := GetAudioInfo(audioPath)

	if err != nil {
		return nil, fmt.Errorf("failed to get audio duration: %w", err)
	}

	totalDuration := audioInfo.Duration

	// Create non-silent segments
	var nonSilentSegments []AudioSegment

	// First segment - from start to first silence
	if silenceStarts[0] > 0 {
		nonSilentSegments = append(nonSilentSegments, AudioSegment{
			StartTime: 0,
			EndTime:   silenceStarts[0],
			Duration:  silenceStarts[0],
		})
	}

	// Middle segments - between silences
	for i := 0; i < len(silenceStarts)-1; i++ {
		segmentStart := silenceEnds[i]
		segmentEnd := silenceStarts[i+1]

		// Skip very short segments
		if segmentEnd-segmentStart < 0.5 {
			continue
		}

		nonSilentSegments = append(nonSilentSegments, AudioSegment{
			StartTime: segmentStart,
			EndTime:   segmentEnd,
			Duration:  segmentEnd - segmentStart,
		})
	}

	// Last segment - from last silence to end
	if len(silenceEnds) > 0 && silenceEnds[len(silenceEnds)-1] < totalDuration {
		segmentStart := silenceEnds[len(silenceEnds)-1]
		segmentEnd := totalDuration

		nonSilentSegments = append(nonSilentSegments, AudioSegment{
			StartTime: segmentStart,
			EndTime:   segmentEnd,
			Duration:  segmentEnd - segmentStart,
		})
	}

	return nonSilentSegments, nil
}

// ExtractAudioSegment extracts a segment from an audio file
func ExtractAudioSegment(inputPath, outputPath string, startTime, endTime float64, audioInfo *AudioInfo) error {
	// Ensure output directory exists
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command with quality preservation
	args := []string{
		"-i", inputPath,
		"-ss", fmt.Sprintf("%.3f", startTime),
		"-to", fmt.Sprintf("%.3f", endTime),
	}

	// Add audio settings to preserve quality
	if audioInfo != nil {
		// Use the same audio codec
		if audioInfo.Codec != "" {
			args = append(args, "-c:a", audioInfo.Codec)
		}

		// Preserve sample rate
		if audioInfo.SampleRate > 0 {
			args = append(args, "-ar", strconv.Itoa(audioInfo.SampleRate))
		}

		// Preserve channels
		if audioInfo.Channels > 0 {
			args = append(args, "-ac", strconv.Itoa(audioInfo.Channels))
		}
	} else {
		// Use default high quality settings
		args = append(args, "-c:a", "libmp3lame", "-q:a", "2")
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

// ConcatenateAudioSegments concatenates multiple audio segments into a single audio file
func ConcatenateAudioSegments(segments []string, outputPath string, audioInfo *AudioInfo) error {
	// Check if segments exist
	if len(segments) == 0 {
		return fmt.Errorf("no segments to concatenate")
	}

	// Create temporary file list
	tempDir := os.TempDir()
	fileListPath := filepath.Join(tempDir, "audio_segments_list.txt")

	fileList, err := os.Create(fileListPath)
	if err != nil {
		return fmt.Errorf("failed to create file list: %w", err)
	}
	defer fileList.Close()
	defer os.Remove(fileListPath)

	// Write segment paths to file list
	for _, segmentPath := range segments {
		if _, err := os.Stat(segmentPath); os.IsNotExist(err) {
			continue // Skip non-existent segments
		}
		fileList.WriteString(fmt.Sprintf("file '%s'\n", segmentPath))
	}
	fileList.Close()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command for concatenation
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", fileListPath,
	}

	// Add audio settings to preserve quality
	if audioInfo != nil {
		// Use the same audio codec if available
		if audioInfo.Codec != "" {
			args = append(args, "-c:a", audioInfo.Codec)
		} else {
			args = append(args, "-c:a", "libmp3lame") // Default to MP3
		}

		// Preserve sample rate if available
		if audioInfo.SampleRate > 0 {
			args = append(args, "-ar", strconv.Itoa(audioInfo.SampleRate))
		}

		// Preserve channels if available
		if audioInfo.Channels > 0 {
			args = append(args, "-ac", strconv.Itoa(audioInfo.Channels))
		}

		// Set quality if using MP3
		if audioInfo.Codec == "libmp3lame" {
			args = append(args, "-q:a", "2") // High quality
		}
	} else {
		// Default high quality settings
		args = append(args, "-c:a", "libmp3lame", "-q:a", "2")
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

/*
* Helpers
 */

// parseSilenceOutput parses FFmpeg silence detection output
func parseSilenceOutput(output string) ([]float64, []float64) {
	// Extract silence_start times
	startRegex := regexp.MustCompile(`silence_start: ([0-9.]+)`)
	startMatches := startRegex.FindAllStringSubmatch(output, -1)

	// Extract silence_end times
	endRegex := regexp.MustCompile(`silence_end: ([0-9.]+)`)
	endMatches := endRegex.FindAllStringSubmatch(output, -1)

	// Convert to float arrays
	var starts, ends []float64

	for _, match := range startMatches {
		if len(match) > 1 {
			time, err := strconv.ParseFloat(match[1], 64)
			if err == nil {
				starts = append(starts, time)
			}
		}
	}

	for _, match := range endMatches {
		if len(match) > 1 {
			time, err := strconv.ParseFloat(match[1], 64)
			if err == nil {
				ends = append(ends, time)
			}
		}
	}

	return starts, ends
}

// copyAudioFile copies a file from src to dst
func copyAudioFile(src, dst string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use FFmpeg for copying to handle audio files properly
	cmd := exec.Command("ffmpeg", "-i", src, "-c", "copy", "-y", dst)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy file: %w - %s", err, string(output))
	}

	return nil
}
