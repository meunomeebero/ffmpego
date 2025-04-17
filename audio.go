package ffmpego

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/meunomeebero/ffmpego/internal/types"
)

/*
* Exported functions
 */

// GetAudioInfo retrieves information about an audio file
func GetAudioInfo(audioPath string) (*types.AudioInfo, error) {
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
	info := &types.AudioInfo{}
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
			// Handle 'N/A' case for bit rate
			if bitRateStr != "N/A" {
				info.BitRate, _ = strconv.Atoi(bitRateStr)
				info.BitRate /= 1000 // Convert bps to kbps
			}
		} else if strings.HasPrefix(line, "duration=") {
			durStr := strings.TrimPrefix(line, "duration=")
			// Handle 'N/A' case for duration
			if durStr != "N/A" {
				info.Duration, _ = strconv.ParseFloat(durStr, 64)
			}
		}
	}

	return info, nil
}

// RemoveAudioSilence processes an audio file by removing silent parts
func RemoveAudioSilence(audioPath, outputPath string, minSilenceLen int, silenceThresh int, config *types.AudioConfig, logger Logger) error {
	// Detect non-silent segments
	segments, err := DetectNonSilentSegments(audioPath, minSilenceLen, silenceThresh)
	if err != nil {
		return fmt.Errorf("failed to detect silence: %w", err)
	}

	// Handle case where no segments are found (entire file is silent)
	if len(segments) == 0 {
		logger.Info("No audible segments found, output will be empty")
		// Create an empty file or handle as appropriate
		f, createErr := os.Create(outputPath)
		if createErr != nil {
			return fmt.Errorf("failed to create empty output file: %w", createErr)
		}
		f.Close()
		return nil
	}

	// Create temporary directory for segments
	tempDir := filepath.Join(os.TempDir(), "audio_processor_"+time.Now().Format("20060102_150405"))
	defer os.RemoveAll(tempDir)

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Extract segments
	segmentPaths := make([]string, 0, len(segments))
	for i, segment := range segments {
		segmentPath := filepath.Join(tempDir, fmt.Sprintf("segment_%03d.mp3", i+1))

		// Get audio info for quality preservation
		audioInfo, infoErr := GetAudioInfo(audioPath)
		if infoErr != nil {
			return fmt.Errorf("failed to get audio info for segment %d: %w", i+1, infoErr)
		}

		err = ExtractAudioSegment(audioPath, segmentPath, segment.StartTime, segment.EndTime, audioInfo)
		if err != nil {
			return fmt.Errorf("failed to extract segment %d: %w", i+1, err)
		}
		segmentPaths = append(segmentPaths, segmentPath)
	}

	// Concatenate segments with specified configuration
	err = concatenateAudioSegmentsWithConfig(segmentPaths, outputPath, config)
	if err != nil {
		return fmt.Errorf("failed to concatenate segments: %w", err)
	}

	return nil
}

// ExtractAudioFromVideo extracts audio from a video file
func ExtractAudioFromVideo(videoPath, outputPath string) error {
	// Check if FFmpeg is available
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vn",                   // Disable video recording
		"-acodec", "libmp3lame", // Specify MP3 codec
		"-ab", "192k", // Set audio bitrate
		"-ar", "44100", // Set audio sample rate
		"-ac", "2", // Set audio channels to stereo
		"-y", // Overwrite output file if it exists
		outputPath)

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg error: %w - %s", err, string(output))
	}

	return nil
}

// DetectNonSilentSegments detects segments of audio that are not silent
func DetectNonSilentSegments(audioPath string, minSilenceLen int, silenceThresh int) ([]types.AudioSegment, error) {
	// Check if FFmpeg is available
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}

	// Build ffmpeg command for silence detection
	args := []string{
		"-i", audioPath,
		"-af", fmt.Sprintf("silencedetect=noise=%ddB:d=%.3f", silenceThresh, float64(minSilenceLen)/1000.0),
		"-f", "null",
		"-",
	}
	cmd := exec.Command("ffmpeg", args...)

	// Execute command and capture stderr
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()

	// FFmpeg often exits with status 1 when using null output, ignore specific error
	if err != nil && stderr.Len() == 0 { // Real error if stderr is empty
		return nil, fmt.Errorf("ffmpeg execution failed: %w", err)
	}

	// Parse stderr output for silence detection lines
	scanner := bufio.NewScanner(&stderr)
	var silenceIntervals [][2]float64

	for scanner.Scan() {
		line := scanner.Text()
		startMatch := silenceStartRegex.FindStringSubmatch(line)
		endMatch := silenceEndRegex.FindStringSubmatch(line)

		if len(startMatch) > 1 {
			start, _ := strconv.ParseFloat(startMatch[1], 64)
			silenceIntervals = append(silenceIntervals, [2]float64{start, -1.0}) // Mark end as unknown
		} else if len(endMatch) > 1 && len(silenceIntervals) > 0 {
			end, _ := strconv.ParseFloat(endMatch[1], 64)
			if silenceIntervals[len(silenceIntervals)-1][1] == -1.0 { // Find corresponding start
				silenceIntervals[len(silenceIntervals)-1][1] = end
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading ffmpeg output: %w", err)
	}

	// Get total duration of the audio
	audioInfo, err := GetAudioInfo(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio duration: %w", err)
	}
	totalDuration := audioInfo.Duration

	// Calculate non-silent segments based on silence intervals
	var nonSilentSegments []types.AudioSegment // Use types.AudioSegment
	lastEnd := 0.0

	for _, interval := range silenceIntervals {
		start := interval[0]
		end := interval[1]

		// Handle case where silence detection might be slightly off at the start
		if start < lastEnd {
			start = lastEnd
		}

		// Ensure end is valid
		if end < start {
			// If end is invalid or not found, assume silence continues to the end
			// This case should be rare with silencedetect but handled defensively
			if lastEnd < start {
				dur := start - lastEnd
				if dur > 0.01 { // Avoid tiny segments
					nonSilentSegments = append(nonSilentSegments, types.AudioSegment{
						StartTime: lastEnd,
						EndTime:   start,
						Duration:  dur,
					})
				}
			}
			lastEnd = totalDuration // Skip the rest
			break
		}

		// Add the non-silent segment before this silence interval
		if start > lastEnd {
			dur := start - lastEnd
			if dur > 0.01 { // Avoid tiny segments due to float precision
				nonSilentSegments = append(nonSilentSegments, types.AudioSegment{
					StartTime: lastEnd,
					EndTime:   start,
					Duration:  dur,
				})
			}
		}
		lastEnd = end
	}

	// Add the final non-silent segment after the last silence
	if lastEnd < totalDuration {
		dur := totalDuration - lastEnd
		if dur > 0.01 { // Avoid tiny segments
			nonSilentSegments = append(nonSilentSegments, types.AudioSegment{
				StartTime: lastEnd,
				EndTime:   totalDuration,
				Duration:  dur,
			})
		}
	}

	return nonSilentSegments, nil
}

// ExtractAudioSegment extracts a segment from an audio file
func ExtractAudioSegment(audioPath, outputPath string, startTime, endTime float64, audioInfo *types.AudioInfo) error {
	// Ensure output directory exists
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command
	args := []string{
		"-i", audioPath,
		"-ss", fmt.Sprintf("%.3f", startTime),
		"-to", fmt.Sprintf("%.3f", endTime),
		"-vn", // No video
	}

	// Preserve audio settings if info is provided
	if audioInfo != nil {
		if audioInfo.Codec != "" {
			args = append(args, "-c:a", audioInfo.Codec)
		}
		if audioInfo.SampleRate > 0 {
			args = append(args, "-ar", strconv.Itoa(audioInfo.SampleRate))
		}
		if audioInfo.Channels > 0 {
			args = append(args, "-ac", strconv.Itoa(audioInfo.Channels))
		}
		if audioInfo.BitRate > 0 {
			args = append(args, "-ab", fmt.Sprintf("%dk", audioInfo.BitRate))
		}
	} else {
		// Default to copy codec if no info provided
		args = append(args, "-c:a", "copy")
	}

	args = append(args, "-y", outputPath)

	// Execute FFmpeg command
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg error: %w - %s", err, string(output))
	}

	return nil
}

// ConcatenateAudioSegments concatenates multiple audio segments into a single file
func ConcatenateAudioSegments(segmentPaths []string, outputPath string, audioInfo *types.AudioInfo) error {
	// Check if segments exist
	if len(segmentPaths) == 0 {
		return fmt.Errorf("no segments to concatenate")
	}

	// Create temporary file list
	tempDir := filepath.Join(os.TempDir(), "audio_concat_"+time.Now().Format("20060102_150405"))
	defer os.RemoveAll(tempDir)

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	fileListPath := filepath.Join(tempDir, "segments_list.txt")

	fileList, err := os.Create(fileListPath)
	if err != nil {
		return fmt.Errorf("failed to create file list: %w", err)
	}
	defer fileList.Close()

	// Write segment paths to file list
	for _, segmentPath := range segmentPaths {
		if _, err := os.Stat(segmentPath); os.IsNotExist(err) {
			continue // Skip non-existent segments
		}
		// Use relative path if possible, otherwise absolute
		relPath, relErr := filepath.Rel(tempDir, segmentPath)
		if relErr != nil {
			relPath = segmentPath // Use absolute if relative fails
		}
		fileList.WriteString(fmt.Sprintf("file '%s'\n", relPath))
	}
	fileList.Close()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command for concatenation
	args := []string{
		"-f", "concat",
		"-safe", "0", // Allow unsafe file paths
		"-i", fileListPath,
		"-vn", // No video
	}

	// Add audio settings to preserve quality if possible
	if audioInfo != nil && audioInfo.Codec != "" && IsCopyCompatibleCodec(audioInfo.Codec) {
		args = append(args, "-c:a", "copy")
	} else {
		// Re-encode if codec is not copy-compatible or info is missing
		if audioInfo != nil {
			if audioInfo.Codec != "" {
				args = append(args, "-c:a", audioInfo.Codec)
			}
			if audioInfo.SampleRate > 0 {
				args = append(args, "-ar", strconv.Itoa(audioInfo.SampleRate))
			}
			if audioInfo.Channels > 0 {
				args = append(args, "-ac", strconv.Itoa(audioInfo.Channels))
			}
			if audioInfo.BitRate > 0 {
				args = append(args, "-ab", fmt.Sprintf("%dk", audioInfo.BitRate))
			}
		} else {
			// Sensible defaults if re-encoding without info
			args = append(args, "-c:a", "libmp3lame", "-ab", "192k")
		}
	}

	args = append(args, "-y", outputPath)

	// Execute FFmpeg command
	cmd := exec.Command("ffmpeg", args...)
	cmd.Dir = tempDir // Run command from tempDir to handle relative paths correctly
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg concat error: %w - %s", err, string(output))
	}

	return nil
}

// concatenateAudioSegmentsWithConfig concatenates multiple audio segments with specific configuration
func concatenateAudioSegmentsWithConfig(segmentPaths []string, outputPath string, config *types.AudioConfig) error {
	// Check if segments exist
	if len(segmentPaths) == 0 {
		return fmt.Errorf("no segments to concatenate")
	}

	// Create temporary file list
	tempDir := filepath.Join(os.TempDir(), "audio_concat_cfg_"+time.Now().Format("20060102_150405"))
	defer os.RemoveAll(tempDir)

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	fileListPath := filepath.Join(tempDir, "segments_list.txt")

	fileList, err := os.Create(fileListPath)
	if err != nil {
		return fmt.Errorf("failed to create file list: %w", err)
	}
	defer fileList.Close()

	// Write segment paths to file list
	for _, segmentPath := range segmentPaths {
		if _, err := os.Stat(segmentPath); os.IsNotExist(err) {
			continue // Skip non-existent segments
		}
		relPath, relErr := filepath.Rel(tempDir, segmentPath)
		if relErr != nil {
			relPath = segmentPath // Use absolute if relative fails
		}
		fileList.WriteString(fmt.Sprintf("file '%s'\n", relPath))
	}
	fileList.Close()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command for concatenation with specific config
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", fileListPath,
		"-vn", // No video
	}

	// Apply audio configuration
	if config != nil {
		if config.Codec != "" {
			args = append(args, "-c:a", config.Codec)
		} else {
			args = append(args, "-c:a", "libmp3lame") // Default codec
		}
		if config.SampleRate > 0 {
			args = append(args, "-ar", strconv.Itoa(config.SampleRate))
		}
		if config.Channels > 0 {
			args = append(args, "-ac", strconv.Itoa(config.Channels))
		}
		if config.BitRate > 0 {
			args = append(args, "-ab", fmt.Sprintf("%dk", config.BitRate))
		} else if config.Quality > 0 { // Use VBR quality if bitrate not set
			args = append(args, "-q:a", strconv.Itoa(config.Quality))
		} else {
			args = append(args, "-ab", "192k") // Default bitrate
		}
	} else {
		// Default settings if no config provided
		args = append(args, "-c:a", "libmp3lame", "-ab", "192k")
	}

	args = append(args, "-y", outputPath)

	// Execute FFmpeg command
	cmd := exec.Command("ffmpeg", args...)
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg concat error: %w - %s", err, string(output))
	}

	return nil
}

/*
* Helpers
 */

var silenceStartRegex = regexp.MustCompile(`silence_start: (\d+\.?\d*)`)
var silenceEndRegex = regexp.MustCompile(`silence_end: (\d+\.?\d*)`) // Adjusted regex
