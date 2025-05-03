package ffmpego

import "strings"

// IsCopyCompatibleCodec checks if a codec can be used with -c copy
// Note: This list might need refinement based on specific FFmpeg version and build.
func IsCopyCompatibleCodec(codec string) bool {
	// Common copy-compatible video/audio codecs
	copyCompatibleCodecs := []string{
		"h264", "h265", "hevc", "vp8", "vp9", "av1", "mjpeg", "mpeg4", "libx264", "libx265",
		"aac", "mp3", "opus", "vorbis", "flac", "pcm_s16le", "pcm_s24le", "pcm_f32le",
		"copy", // Explicit copy is always compatible
	}

	// Handle cases like 'libmp3lame'
	codecLower := strings.ToLower(codec)

	for _, c := range copyCompatibleCodecs {
		if strings.Contains(codecLower, c) {
			return true
		}
	}

	return false
}
