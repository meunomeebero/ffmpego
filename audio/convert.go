package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// ConvertConfig contains configuration for audio conversion
type ConvertConfig struct {
	// Sample rate in Hz (e.g., 44100, 48000)
	SampleRate int

	// Number of audio channels (1 = mono, 2 = stereo)
	Channels int

	// Audio codec (e.g., "aac", "libmp3lame", "flac")
	Codec string

	// Quality (0-9 for VBR, lower is better)
	// For MP3: 0-9 (0 = best quality, 9 = worst)
	// For AAC: 0.1-2 (typically)
	Quality int

	// Bitrate in kbps (e.g., 128, 192, 320)
	Bitrate int
}

// Common audio codecs
const (
	CodecAAC    = "aac"
	CodecMP3    = "libmp3lame"
	CodecFLAC   = "flac"
	CodecOpus   = "libopus"
	CodecVorbis = "libvorbis"
)

// Common sample rates
const (
	SampleRate8000  = 8000
	SampleRate16000 = 16000
	SampleRate22050 = 22050
	SampleRate44100 = 44100
	SampleRate48000 = 48000
	SampleRate96000 = 96000
)

// Convert converts the audio according to the configuration
func (a *Audio) Convert(outputPath string, config ConvertConfig) error {
	// Get audio info
	info, err := a.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get audio info: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command
	args := []string{"-i", a.path}

	// Add conversion arguments
	args = append(args, buildConvertArgs(info, &config)...)

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

// Helper functions

func buildConvertArgs(info *Info, config *ConvertConfig) []string {
	var args []string

	// Audio codec
	codec := config.Codec
	if codec == "" {
		codec = CodecAAC // Default
	}
	args = append(args, "-c:a", codec)

	// Sample rate
	if config.SampleRate > 0 {
		args = append(args, "-ar", strconv.Itoa(config.SampleRate))
	}

	// Channels
	if config.Channels > 0 {
		args = append(args, "-ac", strconv.Itoa(config.Channels))
	}

	// Quality
	if config.Quality > 0 {
		args = append(args, "-q:a", strconv.Itoa(config.Quality))
	}

	// Bitrate
	if config.Bitrate > 0 {
		args = append(args, "-b:a", fmt.Sprintf("%dk", config.Bitrate))
	}

	return args
}

func buildDefaultArgs(info *Info) []string {
	var args []string

	// Preserve codec
	if info.Codec != "" {
		args = append(args, "-c:a", info.Codec)
	} else {
		args = append(args, "-c:a", CodecMP3)
	}

	// Preserve sample rate
	if info.SampleRate > 0 {
		args = append(args, "-ar", strconv.Itoa(info.SampleRate))
	}

	// Preserve channels
	if info.Channels > 0 {
		args = append(args, "-ac", strconv.Itoa(info.Channels))
	}

	// High quality
	args = append(args, "-q:a", "2")

	return args
}
