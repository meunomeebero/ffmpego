package audio

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/meunomeebero/ffmpego/internal/ffutil"
)

// Common silence thresholds in dB (lower = more sensitive)
const (
	SilenceThresholdVeryStrict  = -50 // Detects even very quiet sounds as non-silence
	SilenceThresholdStrict      = -40 // Detects most quiet sounds (recommended for audio)
	SilenceThresholdModerate    = -30 // Balanced - good for most audio
	SilenceThresholdRelaxed     = -20 // Only loud parts are considered non-silence
	SilenceThresholdVeryRelaxed = -10 // Only very loud parts are considered non-silence
)

// Common minimum silence durations in milliseconds
const (
	SilenceDurationVeryShort = 200  // 0.2 seconds - very sensitive
	SilenceDurationShort     = 500  // 0.5 seconds - sensitive (recommended for audio)
	SilenceDurationMedium    = 700  // 0.7 seconds - balanced
	SilenceDurationLong      = 1000 // 1 second - less sensitive
	SilenceDurationVeryLong  = 2000 // 2 seconds - very conservative
)

// SilenceConfig contains configuration for silence detection
type SilenceConfig struct {
	MinSilenceDuration int // Minimum silence duration in milliseconds (use SilenceDuration constants)
	SilenceThreshold   int // Silence threshold in dB (use SilenceThreshold constants)
}

// GetNonSilentSegments detects silent segments in the audio and returns non-silent segments.
// If no silence is detected, returns the entire file as a single segment.
func (a *Audio) GetNonSilentSegments(config SilenceConfig) ([]Segment, error) {
	if config.SilenceThreshold == 0 {
		config.SilenceThreshold = SilenceThresholdModerate
	}
	if config.MinSilenceDuration == 0 {
		config.MinSilenceDuration = SilenceDurationMedium
	}

	info, err := a.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get audio info: %w", err)
	}

	silenceLenSec := float64(config.MinSilenceDuration) / 1000.0

	cmd := exec.Command("ffmpeg",
		"-i", a.path,
		"-af", fmt.Sprintf("silencedetect=noise=%ddB:d=%.3f", config.SilenceThreshold, silenceLenSec),
		"-f", "null", "-")

	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	if err != nil {
		// silencedetect always exits non-zero with -f null; only fail if no silence data was produced
		if !strings.Contains(outputStr, "silence_start") && !strings.Contains(outputStr, "silence_end") {
			return nil, fmt.Errorf("failed to detect silence: %w - %s", err, outputStr)
		}
	}

	starts, ends := ffutil.ParseSilenceOutput(outputStr)
	return ffutil.BuildNonSilentSegments(starts, ends, info.Duration, 0.5), nil
}
