package video

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ConvertConfig contains configuration for video conversion
type ConvertConfig struct {
	// Resolution in format "WIDTHxHEIGHT" (e.g., "1920x1080")
	Resolution string

	// Aspect ratio - automatically adjusts resolution
	AspectRatio AspectRatio

	// Frame rate (e.g., 30, 60)
	FrameRate float64

	// Video codec (e.g., "libx264", "libx265", "libvpx-vp9")
	VideoCodec string

	// Audio codec (e.g., "aac", "libmp3lame")
	AudioCodec string

	// Quality - CRF value (0-51, lower is better quality)
	// Default: 23 (good quality)
	Quality int

	// Encoding preset (ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow)
	// Default: medium
	Preset string

	// Pixel format (e.g., "yuv420p")
	PixelFormat string

	// Bitrate in kbps (e.g., 5000 for 5 Mbps)
	Bitrate int
}

// AspectRatio represents common aspect ratios
type AspectRatio string

const (
	AspectRatio16x9  AspectRatio = "16:9"
	AspectRatio9x16  AspectRatio = "9:16"
	AspectRatio4x3   AspectRatio = "4:3"
	AspectRatio1x1   AspectRatio = "1:1"
	AspectRatio21x9  AspectRatio = "21:9"
	AspectRatioAuto  AspectRatio = "auto"
)

// Common video codecs
const (
	CodecH264   = "libx264"
	CodecH265   = "libx265"
	CodecVP9    = "libvpx-vp9"
	CodecAV1    = "libaom-av1"
	CodecProRes = "prores"
)

// Common audio codecs
const (
	CodecAAC    = "aac"
	CodecMP3    = "libmp3lame"
	CodecFLAC   = "flac"
	CodecOpus   = "libopus"
	CodecVorbis = "libvorbis"
)

// Common presets
const (
	PresetUltrafast = "ultrafast"
	PresetSuperfast = "superfast"
	PresetVeryfast  = "veryfast"
	PresetFaster    = "faster"
	PresetFast      = "fast"
	PresetMedium    = "medium"
	PresetSlow      = "slow"
	PresetSlower    = "slower"
	PresetVeryslow  = "veryslow"
)

// Convert converts the video according to the configuration
func (v *Video) Convert(outputPath string, config ConvertConfig) error {
	// Get video info
	info, err := v.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command
	args := []string{"-i", v.path}

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

	// Handle resolution and aspect ratio
	if config.Resolution != "" {
		args = append(args, "-s", config.Resolution)
	} else if config.AspectRatio != "" && config.AspectRatio != AspectRatioAuto {
		// Calculate resolution based on aspect ratio
		resolution := calculateResolutionFromAspectRatio(info, config.AspectRatio)
		if resolution != "" {
			args = append(args, "-s", resolution)
		}
	}

	// Frame rate
	if config.FrameRate > 0 {
		args = append(args, "-r", fmt.Sprintf("%.3f", config.FrameRate))
	}

	// Video codec
	videoCodec := config.VideoCodec
	if videoCodec == "" {
		videoCodec = CodecH264 // Default
	}
	args = append(args, "-c:v", videoCodec)

	// Audio codec
	audioCodec := config.AudioCodec
	if audioCodec == "" {
		audioCodec = CodecAAC // Default
	}
	args = append(args, "-c:a", audioCodec)

	// Quality (CRF)
	quality := config.Quality
	if quality == 0 {
		quality = 23 // Default medium quality
	}
	args = append(args, "-crf", strconv.Itoa(quality))

	// Preset
	preset := config.Preset
	if preset == "" {
		preset = PresetMedium // Default
	}
	args = append(args, "-preset", preset)

	// Pixel format
	if config.PixelFormat != "" {
		args = append(args, "-pix_fmt", config.PixelFormat)
	}

	// Bitrate
	if config.Bitrate > 0 {
		args = append(args, "-b:v", fmt.Sprintf("%dk", config.Bitrate))
	}

	return args
}

func buildDefaultArgs(info *Info) []string {
	var args []string

	// Preserve codecs
	if info.VideoCodec != "" {
		args = append(args, "-c:v", info.VideoCodec)
	}
	if info.AudioCodec != "" {
		args = append(args, "-c:a", info.AudioCodec)
	}

	// Preserve frame rate
	if info.FrameRate > 0 {
		args = append(args, "-r", fmt.Sprintf("%.3f", info.FrameRate))
	}

	// Preserve resolution
	if info.Width > 0 && info.Height > 0 {
		args = append(args, "-s", fmt.Sprintf("%dx%d", info.Width, info.Height))
	}

	// Preserve pixel format
	if info.PixelFormat != "" {
		args = append(args, "-pix_fmt", info.PixelFormat)
	}

	// High quality settings
	args = append(args, "-crf", "18", "-preset", "medium")

	return args
}

func calculateResolutionFromAspectRatio(info *Info, aspectRatio AspectRatio) string {
	if info == nil || info.Height == 0 {
		return ""
	}

	switch aspectRatio {
	case AspectRatio16x9:
		// Keep height, calculate width for 16:9
		width := (info.Height * 16) / 9
		return fmt.Sprintf("%dx%d", width, info.Height)

	case AspectRatio9x16:
		// Keep width, calculate height for 9:16 (vertical)
		height := (info.Width * 16) / 9
		return fmt.Sprintf("%dx%d", info.Width, height)

	case AspectRatio4x3:
		// Keep height, calculate width for 4:3
		width := (info.Height * 4) / 3
		return fmt.Sprintf("%dx%d", width, info.Height)

	case AspectRatio1x1:
		// Square - use the smaller dimension
		size := info.Width
		if info.Height < info.Width {
			size = info.Height
		}
		return fmt.Sprintf("%dx%d", size, size)

	case AspectRatio21x9:
		// Ultra-wide - keep height, calculate width for 21:9
		width := (info.Height * 21) / 9
		return fmt.Sprintf("%dx%d", width, info.Height)

	default:
		return ""
	}
}

func (c *ConvertConfig) needsReencoding(info *Info) bool {
	if c == nil {
		return false
	}

	// Check if any parameter requires re-encoding
	if c.Resolution != "" && c.Resolution != fmt.Sprintf("%dx%d", info.Width, info.Height) {
		return true
	}

	if c.AspectRatio != "" && c.AspectRatio != AspectRatioAuto {
		return true
	}

	if c.FrameRate > 0 && c.FrameRate != info.FrameRate {
		return true
	}

	if c.VideoCodec != "" && c.VideoCodec != info.VideoCodec {
		return true
	}

	if c.AudioCodec != "" && c.AudioCodec != info.AudioCodec {
		return true
	}

	if c.PixelFormat != "" && c.PixelFormat != info.PixelFormat {
		return true
	}

	if c.Bitrate > 0 {
		return true
	}

	return false
}

func parseResolution(resolution string) string {
	parts := strings.Split(strings.ToLower(resolution), "x")
	if len(parts) != 2 {
		return resolution
	}

	width, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	height, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

	if err1 != nil || err2 != nil {
		return resolution
	}

	return fmt.Sprintf("%dx%d", width, height)
}
