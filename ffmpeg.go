package ffmpego

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Logger interface for logging operations
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Section(name string)
	Step(format string, args ...interface{})
	Success(format string, args ...interface{})
}

// DefaultLogger implements the Logger interface using Go's standard log package
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger(out io.Writer) *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(out, "", log.LstdFlags),
	}
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	l.logger.Printf("[DEBUG] "+format, args...)
}

// Info logs an info message
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	l.logger.Printf("[INFO] "+format, args...)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(format string, args ...interface{}) {
	l.logger.Printf("[WARN] "+format, args...)
}

// Error logs an error message
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+format, args...)
}

// Section logs a section header
func (l *DefaultLogger) Section(name string) {
	l.logger.Printf("[SECTION] === %s ===", name)
}

// Step logs a step in a process
func (l *DefaultLogger) Step(format string, args ...interface{}) {
	l.logger.Printf("[STEP] "+format, args...)
}

// Success logs a success message
func (l *DefaultLogger) Success(format string, args ...interface{}) {
	l.logger.Printf("[SUCCESS] "+format, args...)
}

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

// ExtractFromVideo extracts audio from a video file
func (a *AudioProcessor) ExtractFromVideo(videoPath, outputPath string) error {
	return ExtractAudio(videoPath, outputPath)
}

// RemoveSilence processes an audio file by removing silent parts
func (a *AudioProcessor) RemoveSilence(audioPath, outputPath string, silenceConfig SilenceConfig, audioConfig *AudioConfig) error {
	return RemoveAudioSilence(audioPath, outputPath, silenceConfig.MinSilenceLen, silenceConfig.SilenceThresh,
		audioConfig, a.ffmpeg.logger)
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
