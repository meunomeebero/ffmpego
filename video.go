package ffmpego

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GetVideoInfo retrieves information about a video file
func GetVideoInfo(videoPath string) (*VideoInfo, error) {
	// Check if FFprobe is available
	_, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, fmt.Errorf("ffprobe not found in PATH: %w", err)
	}

	// Get video stream information
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,r_frame_rate,codec_name,pix_fmt",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1",
		videoPath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// Parse output
	info := &VideoInfo{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "width=") {
			widthStr := strings.TrimPrefix(line, "width=")
			info.Width, _ = strconv.Atoi(widthStr)
		} else if strings.HasPrefix(line, "height=") {
			heightStr := strings.TrimPrefix(line, "height=")
			info.Height, _ = strconv.Atoi(heightStr)
		} else if strings.HasPrefix(line, "r_frame_rate=") {
			frStr := strings.TrimPrefix(line, "r_frame_rate=")
			frParts := strings.Split(frStr, "/")
			if len(frParts) == 2 {
				num, _ := strconv.ParseFloat(frParts[0], 64)
				den, _ := strconv.ParseFloat(frParts[1], 64)
				if den > 0 {
					info.FrameRate = num / den
				}
			}
		} else if strings.HasPrefix(line, "codec_name=") {
			info.VideoCodec = strings.TrimPrefix(line, "codec_name=")
		} else if strings.HasPrefix(line, "duration=") {
			durStr := strings.TrimPrefix(line, "duration=")
			info.Duration, _ = strconv.ParseFloat(durStr, 64)
		} else if strings.HasPrefix(line, "pix_fmt=") {
			info.PixelFormat = strings.TrimPrefix(line, "pix_fmt=")
		}
	}

	// Get audio codec info
	cmd = exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1",
		videoPath)

	output, err = cmd.Output()
	if err == nil {
		lines = strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "codec_name=") {
				info.AudioCodec = strings.TrimPrefix(line, "codec_name=")
				break
			}
		}
	}

	return info, nil
}

// RemoveVideoSilence processes a video file by removing silent parts
func RemoveVideoSilence(videoPath, outputPath string, minSilenceLen int, silenceThresh int, config *VideoConfig, logger Logger) error {
	// Create temporary directories
	tempDir := filepath.Join(os.TempDir(), "video_processor_"+time.Now().Format("20060102_150405"))
	audioDir := filepath.Join(tempDir, "audio")
	segmentsDir := filepath.Join(tempDir, "segments")

	defer os.RemoveAll(tempDir) // Cleanup when done

	// Create directories
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		return fmt.Errorf("failed to create audio directory: %w", err)
	}

	if err := os.MkdirAll(segmentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create segments directory: %w", err)
	}

	// Step 1: Get video information for preserving quality
	logger.Section("Analyzing Video")
	logger.Step("Getting video information")

	videoInfo, err := GetVideoInfo(videoPath)
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	logger.Success("Video info: %dx%d, %.2f fps, codec: %s",
		videoInfo.Width, videoInfo.Height, videoInfo.FrameRate, videoInfo.VideoCodec)

	// Step 2: Extract audio from video
	logger.Section("Extracting Audio")
	logger.Step("Extracting audio from video")

	audioPath := filepath.Join(audioDir, "audio.mp3")
	err = ExtractAudioFromVideo(videoPath, audioPath)
	if err != nil {
		return fmt.Errorf("failed to extract audio: %w", err)
	}

	logger.Success("Audio extracted successfully")

	// Step 3: Detect silence in audio
	logger.Section("Analyzing Audio")
	logger.Step("Detecting silence in audio")

	audioSegments, err := DetectNonSilentSegments(audioPath, minSilenceLen, silenceThresh)

	if err != nil {
		return fmt.Errorf("failed to detect silence: %w", err)
	}

	// Handle case when no silence is detected
	if len(audioSegments) == 0 {
		logger.Info("No silent segments detected in the audio")
		// Simply copy the input to output since no processing is needed
		err = copyVideoFile(videoPath, outputPath)
		if err != nil {
			return fmt.Errorf("failed to copy video: %w", err)
		}
		logger.Success("Original video copied to output path")
		return nil
	}

	logger.Success("Detected %d non-silent segments", len(audioSegments))

	// Step 4: Process each segment using goroutines
	logger.Section("Processing Video Segments")
	logger.Info("Extracting %d video segments", len(audioSegments))

	// Determine the number of workers based on CPU cores
	numWorkers := min(min(runtime.NumCPU(), 8), len(audioSegments))

	logger.Info("Using %d parallel workers", numWorkers)

	// Create job and result channels
	type job struct {
		index   int
		segment AudioSegment
	}

	type result struct {
		index       int
		segmentPath string
		err         error
	}

	jobs := make(chan job, len(audioSegments))
	results := make(chan result, len(audioSegments))
	segmentPaths := make([]string, 0, len(audioSegments))

	// Start worker pool
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for w := 1; w <= numWorkers; w++ {
		go func(workerID int) {
			defer wg.Done()

			for j := range jobs {
				logger.Debug("Worker %d processing segment %d", workerID, j.index+1)

				// Create segment path
				segmentPath := filepath.Join(segmentsDir, fmt.Sprintf("segment_%03d%s", j.index+1, filepath.Ext(videoPath)))

				// Extract video segment with quality preservation
				err := extractVideoSegmentWithConfig(
					videoPath,
					segmentPath,
					j.segment.StartTime,
					j.segment.EndTime,
					videoInfo,
					config,
				)

				if err != nil {
					logger.Error("Worker %d failed to extract segment %d: %s",
						workerID, j.index+1, err)
					results <- result{index: j.index, err: err}
					continue
				}

				results <- result{index: j.index, segmentPath: segmentPath, err: nil}
			}
		}(w)
	}

	// Send jobs to workers
	for i, segment := range audioSegments {
		jobs <- job{index: i, segment: segment}
	}
	close(jobs)

	// Wait for all segments to be processed and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect segment paths in order
	validSegments := make([]string, len(audioSegments))
	var errCount int

	for r := range results {
		if r.err != nil {
			errCount++
			continue
		}

		logger.Success("Segment %d/%d processed successfully", r.index+1, len(audioSegments))
		validSegments[r.index] = r.segmentPath
	}

	// Filter out empty segments
	for _, path := range validSegments {
		if path != "" {
			segmentPaths = append(segmentPaths, path)
		}
	}

	if len(segmentPaths) == 0 {
		return fmt.Errorf("all segments failed to process")
	}

	// Step 5: Concatenate segments
	logger.Section("Creating Final Video")
	logger.Step("Concatenating %d video segments", len(segmentPaths))

	// Create directory for output if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Concatenate segments with quality preservation
	err = concatenateVideoSegmentsWithConfig(segmentPaths, outputPath, videoInfo, config)
	if err != nil {
		return fmt.Errorf("failed to concatenate segments: %w", err)
	}

	logger.Success("Video processed successfully")
	logger.Info("Output saved to: %s", outputPath)

	return nil
}

// ExtractVideoSegment extracts a segment from a video file
func ExtractVideoSegment(videoPath, outputPath string, startTime, endTime float64, videoInfo *VideoInfo) error {
	// Ensure output directory exists
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command with precise quality preservation
	args := []string{
		"-i", videoPath,
		"-ss", fmt.Sprintf("%.3f", startTime),
		"-to", fmt.Sprintf("%.3f", endTime),
	}

	// Add video settings to preserve quality and framerate
	if videoInfo != nil {
		// Use the same video codec
		if videoInfo.VideoCodec != "" {
			args = append(args, "-c:v", videoInfo.VideoCodec)
		}

		// Preserve framerate
		if videoInfo.FrameRate > 0 {
			args = append(args, "-r", fmt.Sprintf("%.3f", videoInfo.FrameRate))
		}

		// Preserve resolution
		if videoInfo.Width > 0 && videoInfo.Height > 0 {
			args = append(args, "-s", fmt.Sprintf("%dx%d", videoInfo.Width, videoInfo.Height))
		}

		// Preserve pixel format if available
		if videoInfo.PixelFormat != "" {
			args = append(args, "-pix_fmt", videoInfo.PixelFormat)
		}

		// Preserve audio codec
		if videoInfo.AudioCodec != "" {
			args = append(args, "-c:a", videoInfo.AudioCodec)
		}
	}

	// Add high quality settings and output path
	args = append(args,
		"-crf", "18", // Lower CRF value means higher quality
		"-preset", "slow", // Slower preset means better compression
		"-y", // Overwrite output file if it exists
		outputPath)

	// Execute FFmpeg command
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg error: %w - %s", err, string(output))
	}

	return nil
}

// ConcatenateVideoSegments concatenates multiple video segments into a single video
func ConcatenateVideoSegments(segmentPaths []string, outputPath string, videoInfo *VideoInfo) error {
	// Check if segments exist
	if len(segmentPaths) == 0 {
		return fmt.Errorf("no segments to concatenate")
	}

	// Create temporary file list
	tempDir := filepath.Dir(segmentPaths[0])
	fileListPath := filepath.Join(tempDir, "segments_list.txt")

	fileList, err := os.Create(fileListPath)
	if err != nil {
		return fmt.Errorf("failed to create file list: %w", err)
	}
	defer fileList.Close()
	defer os.Remove(fileListPath)

	// Write segment paths to file list
	for _, segmentPath := range segmentPaths {
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

	// Add video settings to preserve quality
	if videoInfo != nil {
		// Use specified video codec if available
		if videoInfo.VideoCodec != "" {
			// Check if the codec is a copy-compatible codec
			if isCopyCompatibleCodec(videoInfo.VideoCodec) {
				args = append(args, "-c:v", "copy")
			} else {
				args = append(args, "-c:v", videoInfo.VideoCodec)
			}
		} else {
			args = append(args, "-c:v", "libx264") // Default to H.264
		}

		// Preserve framerate
		if videoInfo.FrameRate > 0 {
			args = append(args, "-r", fmt.Sprintf("%.3f", videoInfo.FrameRate))
		}

		// Preserve resolution
		if videoInfo.Width > 0 && videoInfo.Height > 0 {
			args = append(args, "-s", fmt.Sprintf("%dx%d", videoInfo.Width, videoInfo.Height))
		}

		// Preserve pixel format if available
		if videoInfo.PixelFormat != "" {
			args = append(args, "-pix_fmt", videoInfo.PixelFormat)
		}

		// Preserve audio codec
		if videoInfo.AudioCodec != "" {
			// Check if the codec is a copy-compatible codec
			if isCopyCompatibleCodec(videoInfo.AudioCodec) {
				args = append(args, "-c:a", "copy")
			} else {
				args = append(args, "-c:a", videoInfo.AudioCodec)
			}
		} else {
			args = append(args, "-c:a", "aac") // Default to AAC
		}
	} else {
		// If no video info, use copy codec
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

// ResizeVideo resizes a video according to the specified configuration
func ConvertVideo(inputPath, outputPath string, videoInfo *VideoInfo, config *VideoConfig) error {
	// Ensure output directory exists
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command
	args := []string{
		"-i", inputPath,
	}

	// Apply configuration
	if config != nil {
		// Set resolution
		if config.TargetResolution != "" {
			args = append(args, "-s", config.TargetResolution)
		}

		// Set frame rate
		if config.FrameRate > 0 {
			args = append(args, "-r", fmt.Sprintf("%.3f", config.FrameRate))
		}

		// Set video codec
		if config.VideoCodec != "" {
			args = append(args, "-c:v", config.VideoCodec)
		} else {
			// Default to libx264 for compatibility
			args = append(args, "-c:v", "libx264")
		}

		// Set audio codec
		if config.AudioCodec != "" {
			args = append(args, "-c:a", config.AudioCodec)
		} else if config.PreserveCodecs && videoInfo.AudioCodec != "" {
			args = append(args, "-c:a", "copy")
		} else {
			args = append(args, "-c:a", "aac")
		}

		// Set pixel format
		if config.PixelFormat != "" {
			args = append(args, "-pix_fmt", config.PixelFormat)
		}

		// Set quality (CRF)
		if config.CRF > 0 {
			args = append(args, "-crf", strconv.Itoa(config.CRF))
		} else {
			args = append(args, "-crf", "23") // Default medium quality
		}

		// Set encoding preset
		if config.Preset != "" {
			args = append(args, "-preset", config.Preset)
		} else {
			args = append(args, "-preset", "medium")
		}
	} else {
		// Default resize settings for good quality
		args = append(args,
			"-c:v", "libx264",
			"-c:a", "aac",
			"-crf", "23",
			"-preset", "medium")
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

/*
* Helpers
 */

// concatenateVideoSegmentsWithConfig concatenates multiple video segments into a single video with configuration
func concatenateVideoSegmentsWithConfig(segmentPaths []string, outputPath string, videoInfo *VideoInfo, config *VideoConfig) error {
	// Check if segments exist
	if len(segmentPaths) == 0 {
		return fmt.Errorf("no segments to concatenate")
	}

	// Create temporary file list
	tempDir := filepath.Dir(segmentPaths[0])
	fileListPath := filepath.Join(tempDir, "segments_list.txt")

	fileList, err := os.Create(fileListPath)
	if err != nil {
		return fmt.Errorf("failed to create file list: %w", err)
	}
	defer fileList.Close()
	defer os.Remove(fileListPath)

	// Write segment paths to file list
	for _, segmentPath := range segmentPaths {
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

	// Handle config
	if config != nil {
		// Check if we need to change any parameters that would require re-encoding
		needsReencoding := (config.TargetResolution != "" &&
			(parseResolution(config.TargetResolution) != fmt.Sprintf("%dx%d", videoInfo.Width, videoInfo.Height))) ||
			(config.FrameRate > 0 && config.FrameRate != videoInfo.FrameRate) ||
			(config.PixelFormat != "" && config.PixelFormat != videoInfo.PixelFormat) ||
			(config.VideoCodec != "" && config.VideoCodec != videoInfo.VideoCodec) ||
			(config.AudioCodec != "" && config.AudioCodec != videoInfo.AudioCodec)

		if needsReencoding {
			// Video codec
			if config.VideoCodec != "" {
				args = append(args, "-c:v", config.VideoCodec)
			} else {
				args = append(args, "-c:v", "libx264") // Default to H.264
			}

			// Audio codec
			if config.AudioCodec != "" {
				args = append(args, "-c:a", config.AudioCodec)
			} else {
				args = append(args, "-c:a", "aac") // Default to AAC
			}

			// Frame rate
			if config.FrameRate > 0 {
				args = append(args, "-r", fmt.Sprintf("%.3f", config.FrameRate))
			}

			// Resolution
			if config.TargetResolution != "" {
				args = append(args, "-s", config.TargetResolution)
			}

			// Pixel format
			if config.PixelFormat != "" {
				args = append(args, "-pix_fmt", config.PixelFormat)
			}

			// Quality (CRF)
			if config.CRF > 0 {
				args = append(args, "-crf", strconv.Itoa(config.CRF))
			} else {
				args = append(args, "-crf", "18") // Default high quality
			}

			// Encoding preset
			if config.Preset != "" {
				args = append(args, "-preset", config.Preset)
			} else {
				args = append(args, "-preset", "medium") // Default preset
			}
		} else if config.PreserveCodecs {
			// Just copy streams without re-encoding
			args = append(args, "-c", "copy")
		} else {
			// Use copy for both streams unless specific codecs are requested
			args = append(args, "-c:v", "copy", "-c:a", "copy")
		}
	} else {
		// If no config, use copy codec for speed and quality
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

// parseResolution parses a resolution string like "1920x1080" and returns normalized format
func parseResolution(resolution string) string {
	// Split by 'x' or 'X'
	parts := strings.Split(strings.ToLower(resolution), "x")
	if len(parts) != 2 {
		return resolution // Return original if invalid
	}

	// Try to parse width and height
	width, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	height, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

	if err1 != nil || err2 != nil {
		return resolution // Return original if invalid
	}

	// Return normalized format
	return fmt.Sprintf("%dx%d", width, height)
}

// isCopyCompatibleCodec checks if a codec can be used with -c copy
func isCopyCompatibleCodec(codec string) bool {
	// Common copy-compatible video codecs
	copyCompatibleCodecs := []string{
		"h264", "h265", "hevc", "vp8", "vp9", "av1", "mjpeg", "mpeg4", "libx264", "libx265",
		"aac", "mp3", "opus", "vorbis", "flac", "pcm_s16le", "pcm_s24le", "pcm_f32le",
	}

	for _, c := range copyCompatibleCodecs {
		if strings.Contains(codec, c) {
			return true
		}
	}

	return false
}

// copyVideoFile copies a file from src to dst
func copyVideoFile(src, dst string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use FFmpeg for copying to handle video files properly
	cmd := exec.Command("ffmpeg", "-i", src, "-c", "copy", "-y", dst)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy file: %w - %s", err, string(output))
	}

	return nil
}

// extractVideoSegmentWithConfig extracts a segment from a video file with configuration options
func extractVideoSegmentWithConfig(videoPath, outputPath string, startTime, endTime float64, videoInfo *VideoInfo, config *VideoConfig) error {
	// Ensure output directory exists
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command with precise quality preservation
	args := []string{
		"-i", videoPath,
		"-ss", fmt.Sprintf("%.3f", startTime),
		"-to", fmt.Sprintf("%.3f", endTime),
	}

	// Apply video configuration if provided
	if config != nil {
		// Video codec
		if config.VideoCodec != "" {
			args = append(args, "-c:v", config.VideoCodec)
		} else if !config.PreserveCodecs && videoInfo.VideoCodec != "" {
			args = append(args, "-c:v", videoInfo.VideoCodec)
		}

		// Audio codec
		if config.AudioCodec != "" {
			args = append(args, "-c:a", config.AudioCodec)
		} else if !config.PreserveCodecs && videoInfo.AudioCodec != "" {
			args = append(args, "-c:a", videoInfo.AudioCodec)
		}

		// Frame rate
		if config.FrameRate > 0 {
			args = append(args, "-r", fmt.Sprintf("%.3f", config.FrameRate))
		} else if videoInfo.FrameRate > 0 {
			args = append(args, "-r", fmt.Sprintf("%.3f", videoInfo.FrameRate))
		}

		// Resolution
		if config.TargetResolution != "" {
			args = append(args, "-s", config.TargetResolution)
		} else if videoInfo.Width > 0 && videoInfo.Height > 0 {
			args = append(args, "-s", fmt.Sprintf("%dx%d", videoInfo.Width, videoInfo.Height))
		}

		// Pixel format
		if config.PixelFormat != "" {
			args = append(args, "-pix_fmt", config.PixelFormat)
		} else if videoInfo.PixelFormat != "" {
			args = append(args, "-pix_fmt", videoInfo.PixelFormat)
		}

		// Quality (CRF)
		if config.CRF > 0 {
			args = append(args, "-crf", strconv.Itoa(config.CRF))
		} else {
			args = append(args, "-crf", "18") // Default high quality
		}

		// Encoding preset
		if config.Preset != "" {
			args = append(args, "-preset", config.Preset)
		} else {
			args = append(args, "-preset", "medium") // Default preset
		}
	} else {
		// Use default settings from original video
		if videoInfo.VideoCodec != "" {
			args = append(args, "-c:v", videoInfo.VideoCodec)
		}

		if videoInfo.AudioCodec != "" {
			args = append(args, "-c:a", videoInfo.AudioCodec)
		}

		if videoInfo.FrameRate > 0 {
			args = append(args, "-r", fmt.Sprintf("%.3f", videoInfo.FrameRate))
		}

		if videoInfo.Width > 0 && videoInfo.Height > 0 {
			args = append(args, "-s", fmt.Sprintf("%dx%d", videoInfo.Width, videoInfo.Height))
		}

		if videoInfo.PixelFormat != "" {
			args = append(args, "-pix_fmt", videoInfo.PixelFormat)
		}

		// Set default high quality
		args = append(args,
			"-crf", "18",
			"-preset", "medium")
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
