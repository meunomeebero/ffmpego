package types

// Constants for video quality (CRF)
const (
	// CRF (Constant Rate Factor) - lower value = better quality
	VideoQualityLossless = 0  // No perceptible visual loss
	VideoQualityHighest  = 18 // Very high quality, almost no visible loss
	VideoQualityHigher   = 20 // Very good quality for most cases
	VideoQualityHigh     = 23 // Default ffmpeg quality, good quality
	VideoQualityMedium   = 26 // Medium quality, good balance between size and quality
	VideoQualityLow      = 29 // Lower quality, smaller file
	VideoQualityLower    = 32 // Low quality, much smaller file
	VideoQualityLowest   = 35 // Very low quality, significant visual loss
)

// Constants for encoding presets
const (
	// Encoding presets - faster = less efficient compression
	PresetUltrafast = "ultrafast" // Fastest, larger file
	PresetSuperfast = "superfast"
	PresetVeryfast  = "veryfast"
	PresetFaster    = "faster"
	PresetFast      = "fast"
	PresetMedium    = "medium" // Default ffmpeg preset, good balance
	PresetSlow      = "slow"
	PresetSlower    = "slower"
	PresetVeryslow  = "veryslow" // Slowest, smaller file
)

// Constants for audio quality
const (
	// Audio quality - lower value = better quality
	AudioQualityHighest = 0 // Best possible quality
	AudioQualityHigher  = 2 // Very high quality
	AudioQualityHigh    = 3 // High quality
	AudioQualityMedium  = 5 // Medium quality
	AudioQualityLow     = 7 // Low quality
	AudioQualityLowest  = 9 // Lowest acceptable quality
)

// Constants for common codecs
const (
	// Video codecs
	VideoCodecH264   = "libx264"    // Most compatible
	VideoCodecH265   = "libx265"    // Better compression, less compatible
	VideoCodecVP9    = "libvpx-vp9" // Good quality and size, for web
	VideoCodecAV1    = "libaom-av1" // Next-gen, excellent quality/size
	VideoCodecProRes = "prores"     // For professional editing

	// Audio codecs
	AudioCodecAAC    = "aac"        // Good quality, widely compatible
	AudioCodecMP3    = "libmp3lame" // Very compatible
	AudioCodecFLAC   = "flac"       // Lossless, larger file
	AudioCodecOpus   = "libopus"    // Excellent quality/size for voice
	AudioCodecVorbis = "libvorbis"  // Good quality, for web
)

// Constants for common resolutions
const (
	Resolution4K     = "3840x2160"
	ResolutionQHD    = "2560x1440"
	ResolutionFullHD = "1920x1080"
	ResolutionHD     = "1280x720"
	ResolutionSD     = "854x480"
	ResolutionLowSD  = "640x360"
)

// Constants for common pixel formats
const (
	// YUV formats
	PixelFormatYuv420p = "yuv420p" // Most compatible, 4:2:0 subsampling
	PixelFormatYuv422p = "yuv422p" // Better quality, 4:2:2 subsampling
	PixelFormatYuv444p = "yuv444p" // Best quality, 4:4:4 subsampling

	// RGB formats
	PixelFormatRgb24 = "rgb24" // Standard RGB, 8 bits per channel
	PixelFormatRgb48 = "rgb48" // High-depth RGB, 16 bits per channel

	// For HDR content
	PixelFormatYuv420p10le = "yuv420p10le" // 10-bit YUV with 4:2:0 subsampling
	PixelFormatYuv422p10le = "yuv422p10le" // 10-bit YUV with 4:2:2 subsampling
)

// Constants for silence thresholds
const (
	SilenceThresholdStrict  = -20 // Detects only deeper silence (less silence detected)
	SilenceThresholdDefault = -30 // Good for most cases
	SilenceThresholdRelaxed = -40 // Detects more silence (more sensitive)
)

// Constants for silence duration
const (
	SilenceDurationShort  = 300  // 300ms - detects short pauses
	SilenceDurationMedium = 700  // 700ms - good for most cases
	SilenceDurationLong   = 1500 // 1.5s - only long pauses
)

// VideoInfo contains information about a video file
type VideoInfo struct {
	Width         int
	Height        int
	Duration      float64
	FrameRate     float64
	VideoCodec    string
	AudioCodec    string
	PixelFormat   string
	FileSizeBytes int64
}

// AudioInfo contains information about an audio file
type AudioInfo struct {
	Duration      float64
	SampleRate    int
	Channels      int
	Codec         string
	BitRate       int
	FileSizeBytes int64
}

// MediaSegment represents a segment of media (audio or video)
type MediaSegment struct {
	StartTime float64
	EndTime   float64
	Duration  float64
	Path      string // Optional, used when representing file segments
}

// AudioSegment represents a segment of audio (compatibility type)
type AudioSegment struct {
	StartTime float64
	EndTime   float64
	Duration  float64
}

// ProcessingOptions contains options for processing media files
type ProcessingOptions struct {
	MinSilenceLen  int     // Silence detection minimum length in ms
	SilenceThresh  int     // Silence detection threshold in dB
	Quality        int     // Output quality (1-31 for video, 1-9 for audio)
	PreserveCodecs bool    // Whether to preserve original codecs
	VideoWidth     int     // Output video width (0 = preserve original)
	VideoHeight    int     // Output video height (0 = preserve original)
	FrameRate      float64 // Output frame rate (0 = preserve original)
	Channels       int     // Output audio channels (0 = preserve original)
	SampleRate     int     // Output audio sample rate (0 = preserve original)
}

// VideoConfig contains configuration options for video processing
type VideoConfig struct {
	TargetResolution string  // Format: "WIDTHxHEIGHT" (e.g., "1920x1080") or empty for original
	FrameRate        float64 // Target frame rate or 0 for original
	Quality          int     // Output quality (1-31 for video, lower is better) or 0 for default
	VideoCodec       string  // Output video codec or empty for default/original
	AudioCodec       string  // Output audio codec or empty for default/original
	PreserveCodecs   bool    // Whether to preserve original codecs (overrides codec settings)
	CRF              int     // Constant Rate Factor (0-51, lower is better quality) or 0 for default
	Preset           string  // Encoding preset (ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow) or empty for default
	PixelFormat      string  // Output pixel format or empty for default/original
}

// AudioConfig contains configuration options for audio processing
type AudioConfig struct {
	SampleRate int    // Target sample rate in Hz or 0 for original
	Channels   int    // Number of audio channels or 0 for original
	Quality    int    // Output quality (0-9 for audio, lower is better) or 0 for default
	Codec      string // Output audio codec or empty for default/original
	BitRate    int    // Output bit rate in kbps or 0 for default/variable
}

// SilenceConfig contains configuration for silence detection and removal
type SilenceConfig struct {
	MinSilenceLen int // Minimum silence length in milliseconds
	SilenceThresh int // Silence threshold in dB
}
