package ffutil

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
)

var depsOnce sync.Once
var depsErr error

// CheckDependencies verifies that ffmpeg and ffprobe are available in PATH.
// The check runs only once per process; subsequent calls return the cached result.
func CheckDependencies() error {
	depsOnce.Do(func() {
		if _, err := exec.LookPath("ffmpeg"); err != nil {
			depsErr = fmt.Errorf("ffmpeg not found in PATH: %w", err)
			return
		}
		if _, err := exec.LookPath("ffprobe"); err != nil {
			depsErr = fmt.Errorf("ffprobe not found in PATH: %w", err)
		}
	})
	return depsErr
}

var (
	silenceStartRegex = regexp.MustCompile(`silence_start: ([0-9.]+)`)
	silenceEndRegex   = regexp.MustCompile(`silence_end: ([0-9.]+)`)
)

// ParseSilenceOutput parses ffmpeg silencedetect filter output and returns
// silence start and end times.
func ParseSilenceOutput(output string) (starts, ends []float64) {
	for _, match := range silenceStartRegex.FindAllStringSubmatch(output, -1) {
		if len(match) > 1 {
			if t, err := strconv.ParseFloat(match[1], 64); err == nil {
				starts = append(starts, t)
			}
		}
	}

	for _, match := range silenceEndRegex.FindAllStringSubmatch(output, -1) {
		if len(match) > 1 {
			if t, err := strconv.ParseFloat(match[1], 64); err == nil {
				ends = append(ends, t)
			}
		}
	}

	return starts, ends
}

// BuildNonSilentSegments converts silence boundaries into non-silent segments.
// totalDuration is the total duration of the media file.
// minSegmentDuration is the minimum duration for a segment to be included (typically 0.5s).
func BuildNonSilentSegments(starts, ends []float64, totalDuration, minSegmentDuration float64) []Segment {
	if len(starts) == 0 {
		return []Segment{{
			StartTime: 0,
			EndTime:   totalDuration,
			Duration:  totalDuration,
		}}
	}

	var segments []Segment
	pos := 0.0

	for i := 0; i < len(starts); i++ {
		if starts[i] > pos {
			seg := Segment{
				StartTime: pos,
				EndTime:   starts[i],
				Duration:  starts[i] - pos,
			}
			if seg.Duration >= minSegmentDuration {
				segments = append(segments, seg)
			}
		}
		if i < len(ends) {
			pos = ends[i]
		} else {
			pos = totalDuration
		}
	}

	if pos < totalDuration {
		seg := Segment{
			StartTime: pos,
			EndTime:   totalDuration,
			Duration:  totalDuration - pos,
		}
		if seg.Duration >= minSegmentDuration {
			segments = append(segments, seg)
		}
	}

	return segments
}

// Segment represents a time-based segment of media.
type Segment struct {
	StartTime float64
	EndTime   float64
	Duration  float64
}
