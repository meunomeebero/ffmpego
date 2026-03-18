package video

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	AspectRatio16x9 AspectRatio = "16:9"
	AspectRatio9x16 AspectRatio = "9:16"
	AspectRatio4x3  AspectRatio = "4:3"
	AspectRatio1x1  AspectRatio = "1:1"
	AspectRatio21x9 AspectRatio = "21:9"
	AspectRatioAuto AspectRatio = "auto"
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
	info, err := v.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	args := []string{"-i", v.path}
	args = append(args, buildConvertArgs(info, &config)...)
	args = append(args, "-y", outputPath)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg error: %w - %s", err, string(output))
	}
	return nil
}

func buildConvertArgs(info *Info, config *ConvertConfig) []string {
	var args []string

	if config.Resolution != "" {
		args = append(args, "-s", config.Resolution)
	} else if config.AspectRatio != "" && config.AspectRatio != AspectRatioAuto {
		resolution := calculateResolutionFromAspectRatio(info, config.AspectRatio)
		if resolution != "" {
			args = append(args, "-s", resolution)
		}
	}

	if config.FrameRate > 0 {
		args = append(args, "-r", fmt.Sprintf("%.3f", config.FrameRate))
	}

	videoCodec := config.VideoCodec
	if videoCodec == "" {
		videoCodec = CodecH264
	}
	args = append(args, "-c:v", videoCodec)

	audioCodec := config.AudioCodec
	if audioCodec == "" {
		audioCodec = CodecAAC
	}
	args = append(args, "-c:a", audioCodec)

	quality := config.Quality
	if quality == 0 {
		quality = 23
	}
	args = append(args, "-crf", strconv.Itoa(quality))

	preset := config.Preset
	if preset == "" {
		preset = PresetMedium
	}
	args = append(args, "-preset", preset)

	if config.PixelFormat != "" {
		args = append(args, "-pix_fmt", config.PixelFormat)
	}

	if config.Bitrate > 0 {
		args = append(args, "-b:v", fmt.Sprintf("%dk", config.Bitrate))
	}

	return args
}

// roundEven rounds n to the nearest even number (required by most video codecs).
func roundEven(n int) int {
	return (n + 1) &^ 1
}

func calculateResolutionFromAspectRatio(info *Info, aspectRatio AspectRatio) string {
	if info == nil || info.Height == 0 {
		return ""
	}

	switch aspectRatio {
	case AspectRatio16x9:
		width := roundEven((info.Height * 16) / 9)
		return fmt.Sprintf("%dx%d", width, roundEven(info.Height))
	case AspectRatio9x16:
		height := roundEven((info.Width * 16) / 9)
		return fmt.Sprintf("%dx%d", roundEven(info.Width), height)
	case AspectRatio4x3:
		width := roundEven((info.Height * 4) / 3)
		return fmt.Sprintf("%dx%d", width, roundEven(info.Height))
	case AspectRatio1x1:
		size := info.Width
		if info.Height < info.Width {
			size = info.Height
		}
		size = roundEven(size)
		return fmt.Sprintf("%dx%d", size, size)
	case AspectRatio21x9:
		width := roundEven((info.Height * 21) / 9)
		return fmt.Sprintf("%dx%d", width, roundEven(info.Height))
	default:
		return ""
	}
}

// decoderToEncoder maps ffprobe decoder names to ffmpeg encoder names.
var decoderToEncoder = map[string]string{
	"h264":       CodecH264,
	"hevc":       CodecH265,
	"h265":       CodecH265,
	"vp9":        CodecVP9,
	"av1":        CodecAV1,
	"prores":     CodecProRes,
	"aac":        CodecAAC,
	"mp3":        CodecMP3,
	"flac":       CodecFLAC,
	"opus":       CodecOpus,
	"vorbis":     CodecVorbis,
}

// encoderForDecoder returns the ffmpeg encoder name for an ffprobe decoder name.
func encoderForDecoder(decoder string) string {
	if enc, ok := decoderToEncoder[decoder]; ok {
		return enc
	}
	return decoder
}

func (c *ConvertConfig) needsReencoding(info *Info) bool {
	if c == nil {
		return false
	}
	if c.Resolution != "" && c.Resolution != fmt.Sprintf("%dx%d", info.Width, info.Height) {
		return true
	}
	if c.AspectRatio != "" && c.AspectRatio != AspectRatioAuto {
		return true
	}
	if c.FrameRate > 0 && c.FrameRate != info.FrameRate {
		return true
	}
	if c.VideoCodec != "" && c.VideoCodec != encoderForDecoder(info.VideoCodec) {
		return true
	}
	if c.AudioCodec != "" && c.AudioCodec != encoderForDecoder(info.AudioCodec) {
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
