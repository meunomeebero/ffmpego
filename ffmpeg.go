package ffmpego

import (
	"fmt"
	"os"
	"path/filepath"
)

// FFmpeg is the main struct that provides access to all FFmpeg functionality
type FFmpeg struct {
	logger Logger
	Audio  *AudioProcessor
	Video  *VideoProcessor
}

// New creates a new instance of the FFmpeg library with a default logger
func New() *FFmpeg {
	return NewWithLogger(NewDefaultLogger(os.Stdout))
}

// NewWithLogger creates a new instance of the FFmpeg library with a custom logger
func NewWithLogger(logger Logger) *FFmpeg {
	ffmpeg := &FFmpeg{
		logger: logger,
	}

	// Initialize the processors and set their parent references
	ffmpeg.Audio = &AudioProcessor{ffmpeg: ffmpeg}
	ffmpeg.Video = &VideoProcessor{ffmpeg: ffmpeg}

	return ffmpeg
}

// AudioProcessor provides audio-specific functionality
type AudioProcessor struct {
	ffmpeg *FFmpeg
}

// GetInfo retrieves information about an audio file
func (a *AudioProcessor) GetInfo(audioPath string) (*AudioInfo, error) {
	info, err := GetAudioInfo(audioPath)
	if err != nil {
		return nil, err
	}

	// Add file size information
	fileInfo, err := os.Stat(audioPath)
	if err == nil {
		info.FileSizeBytes = fileInfo.Size()
	}

	return info, nil
}

// RemoveSilence processes an audio file by removing silent parts
func (a *AudioProcessor) RemoveSilence(audioPath, outputPath string, silenceConfig SilenceConfig, audioConfig *AudioConfig) error {
	return RemoveAudioSilence(audioPath, outputPath, silenceConfig.MinSilenceLen, silenceConfig.SilenceThresh,
		audioConfig, a.ffmpeg.logger)
}

// ExtractFromVideo extracts audio from a video file
func (a *AudioProcessor) ExtractFromVideo(videoPath, outputPath string) error {
	return ExtractAudioFromVideo(videoPath, outputPath)
}

// ExtractSegment extracts a segment from an audio file
func (a *AudioProcessor) ExtractSegment(audioPath, outputPath string, startTime, endTime float64, audioInfo *AudioInfo) error {
	return ExtractAudioSegment(audioPath, outputPath, startTime, endTime, audioInfo)
}

// ConcatenateSegments concatenates multiple audio segments into a single audio
func (a *AudioProcessor) ConcatenateSegments(segments []string, outputPath string, audioInfo *AudioInfo) error {
	return ConcatenateAudioSegments(segments, outputPath, audioInfo)
}

// VideoProcessor provides video-specific functionality
type VideoProcessor struct {
	ffmpeg *FFmpeg
}

// GetInfo retrieves information about a video file
func (v *VideoProcessor) GetInfo(videoPath string) (*VideoInfo, error) {
	info, err := GetVideoInfo(videoPath)
	if err != nil {
		return nil, err
	}

	// Add file size information
	fileInfo, err := os.Stat(videoPath)
	if err == nil {
		info.FileSizeBytes = fileInfo.Size()
	}

	return info, nil
}

// RemoveSilence processes a video file by removing silent parts
func (v *VideoProcessor) RemoveSilence(videoPath, outputPath string, silenceConfig SilenceConfig, videoConfig *VideoConfig) error {
	return RemoveVideoSilence(videoPath, outputPath, silenceConfig.MinSilenceLen, silenceConfig.SilenceThresh,
		videoConfig, v.ffmpeg.logger)
}

// ExtractSegment extracts a segment from a video file
func (v *VideoProcessor) ExtractSegment(videoPath, outputPath string, startTime, endTime float64, videoInfo *VideoInfo) error {
	return ExtractVideoSegment(videoPath, outputPath, startTime, endTime, videoInfo)
}

// ConcatenateSegments concatenates multiple video segments into a single video
func (v *VideoProcessor) ConcatenateSegments(segments []string, outputPath string, videoInfo *VideoInfo) error {
	return ConcatenateVideoSegments(segments, outputPath, videoInfo)
}

// Resize resizes a video file according to the specified configuration
func (v *VideoProcessor) Resize(inputPath, outputPath string, config *VideoConfig) error {
	// Get video info to preserve aspects that aren't changing
	videoInfo, err := v.GetInfo(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return ResizeVideo(inputPath, outputPath, videoInfo, config)
}
