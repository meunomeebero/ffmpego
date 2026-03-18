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
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	args := []string{"-i", a.path}
	args = append(args, buildConvertArgs(&config)...)
	args = append(args, "-y", outputPath)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg error: %w - %s", err, string(output))
	}
	return nil
}

func buildConvertArgs(config *ConvertConfig) []string {
	var args []string

	codec := config.Codec
	if codec == "" {
		codec = CodecAAC
	}
	args = append(args, "-c:a", codec)

	if config.SampleRate > 0 {
		args = append(args, "-ar", strconv.Itoa(config.SampleRate))
	}

	if config.Channels > 0 {
		args = append(args, "-ac", strconv.Itoa(config.Channels))
	}

	if config.Quality > 0 {
		args = append(args, "-q:a", strconv.Itoa(config.Quality))
	}

	if config.Bitrate > 0 {
		args = append(args, "-b:a", fmt.Sprintf("%dk", config.Bitrate))
	}

	return args
}
