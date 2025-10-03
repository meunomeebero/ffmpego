package audio

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// SilenceConfig contains configuration for silence detection
type SilenceConfig struct {
	MinSilenceDuration int // Minimum silence duration in milliseconds
	SilenceThreshold   int // Silence threshold in dB (e.g., -30)
}

// DetectSilence detects silent segments in the audio and returns non-silent segments
func (a *Audio) DetectSilence(config SilenceConfig) ([]Segment, error) {
	// Convert ms to seconds for FFmpeg
	silenceLenSec := float64(config.MinSilenceDuration) / 1000.0

	// Use FFmpeg's silencedetect filter
	cmd := exec.Command("ffmpeg",
		"-i", a.path,
		"-af", fmt.Sprintf("silencedetect=noise=%ddB:d=%.3f", config.SilenceThreshold, silenceLenSec),
		"-f", "null", "-")

	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(output), "silencedetect") {
		return nil, fmt.Errorf("failed to detect silence: %w - %s", err, string(output))
	}

	// Extract silence_start and silence_end times
	silenceStarts, silenceEnds := parseSilenceOutput(string(output))

	// If no silence detected, return empty result
	if len(silenceStarts) == 0 || len(silenceEnds) == 0 {
		return []Segment{}, nil
	}

	// Get total audio duration
	info, err := a.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get audio duration: %w", err)
	}

	totalDuration := info.Duration

	// Create non-silent segments
	var segments []Segment

	// First segment - from start to first silence
	if silenceStarts[0] > 0 {
		segments = append(segments, Segment{
			StartTime: 0,
			EndTime:   silenceStarts[0],
			Duration:  silenceStarts[0],
		})
	}

	// Middle segments - between silences
	for i := 0; i < len(silenceStarts)-1; i++ {
		segmentStart := silenceEnds[i]
		segmentEnd := silenceStarts[i+1]

		// Skip very short segments
		if segmentEnd-segmentStart < 0.5 {
			continue
		}

		segments = append(segments, Segment{
			StartTime: segmentStart,
			EndTime:   segmentEnd,
			Duration:  segmentEnd - segmentStart,
		})
	}

	// Last segment - from last silence to end
	if len(silenceEnds) > 0 && silenceEnds[len(silenceEnds)-1] < totalDuration {
		segmentStart := silenceEnds[len(silenceEnds)-1]
		segmentEnd := totalDuration

		segments = append(segments, Segment{
			StartTime: segmentStart,
			EndTime:   segmentEnd,
			Duration:  segmentEnd - segmentStart,
		})
	}

	return segments, nil
}

// Helper functions

func parseSilenceOutput(output string) ([]float64, []float64) {
	startRegex := regexp.MustCompile(`silence_start: ([0-9.]+)`)
	startMatches := startRegex.FindAllStringSubmatch(output, -1)

	endRegex := regexp.MustCompile(`silence_end: ([0-9.]+)`)
	endMatches := endRegex.FindAllStringSubmatch(output, -1)

	var starts, ends []float64

	for _, match := range startMatches {
		if len(match) > 1 {
			time, err := strconv.ParseFloat(match[1], 64)
			if err == nil {
				starts = append(starts, time)
			}
		}
	}

	for _, match := range endMatches {
		if len(match) > 1 {
			time, err := strconv.ParseFloat(match[1], 64)
			if err == nil {
				ends = append(ends, time)
			}
		}
	}

	return starts, ends
}

