package ffutil

import "fmt"

// DefaultFadeDurationSec is the duration in seconds of the audio fade applied at segment
// boundaries during silence removal.
//
// Why this exists: when audio is cut at arbitrary points (not zero-crossings), the waveform
// jumps discontinuously at each edit. After concatenation these jumps produce audible clicks,
// pops, or a "flickering" sound. A short fade (30ms) smooths the waveform to zero at each
// boundary, making the cuts inaudible while being too brief to affect perceived volume.
const DefaultFadeDurationSec = 0.03

// AudioFadeFilter returns an ffmpeg audio filter string that applies a fade-in at the start
// and a fade-out at the end of a segment. segmentDuration is the total duration of the segment
// in seconds. fadeDuration is the length of each fade in seconds (use DefaultFadeDurationSec
// for the recommended value).
//
// The returned string is meant to be used with ffmpeg's -af flag:
//
//	args = append(args, "-af", ffutil.AudioFadeFilter(duration, ffutil.DefaultFadeDurationSec))
func AudioFadeFilter(segmentDuration, fadeDuration float64) string {
	if fadeDuration <= 0 {
		fadeDuration = DefaultFadeDurationSec
	}
	fadeOutStart := segmentDuration - fadeDuration
	if fadeOutStart < 0 {
		fadeOutStart = 0
	}
	return fmt.Sprintf("afade=t=in:d=%.3f,afade=t=out:st=%.3f:d=%.3f",
		fadeDuration, fadeOutStart, fadeDuration)
}
