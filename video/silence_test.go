package video

import (
	"testing"
)

func TestGetNonSilentSegments_SilenceAtStart(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("silence-start.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}

	segments, err := v.GetNonSilentSegments(config)
	if err != nil {
		t.Fatalf("GetNonSilentSegments: %v", err)
	}

	if len(segments) < 1 {
		t.Fatalf("expected at least 1 segment, got %d", len(segments))
	}

	// The first segment should start after ~2s of silence.
	const silenceDuration = 2.0
	const tolerance = 0.7
	if segments[0].StartTime < silenceDuration-tolerance {
		t.Errorf("first segment starts at %.3f, expected to start after ~%.1fs of silence",
			segments[0].StartTime, silenceDuration)
	}
}

func TestGetNonSilentSegments_SilenceMiddle(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("silence-middle.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}

	segments, err := v.GetNonSilentSegments(config)
	if err != nil {
		t.Fatalf("GetNonSilentSegments: %v", err)
	}

	if len(segments) < 2 {
		t.Fatalf("expected at least 2 segments (tone-silence-tone), got %d", len(segments))
	}
}

func TestGetNonSilentSegments_SilenceAtEnd(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("silence-end.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}

	segments, err := v.GetNonSilentSegments(config)
	if err != nil {
		t.Fatalf("GetNonSilentSegments: %v", err)
	}

	if len(segments) < 1 {
		t.Fatalf("expected at least 1 segment, got %d", len(segments))
	}

	last := segments[len(segments)-1]

	// silence-end.mp4 = 3s tone + 2s silence, total 5s.
	// The last segment should end before the trailing silence begins (i.e., before ~5s).
	const totalDuration = 5.0
	const trailingSilence = 2.0
	const tolerance = 0.7
	maxExpectedEnd := totalDuration - trailingSilence + tolerance

	if last.EndTime > maxExpectedEnd {
		t.Errorf("last segment ends at %.3f, expected before ~%.1f (end of tone portion)",
			last.EndTime, maxExpectedEnd)
	}
}

func TestGetNonSilentSegments_NoSilence(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}

	segments, err := v.GetNonSilentSegments(config)
	if err != nil {
		t.Fatalf("GetNonSilentSegments: %v", err)
	}

	if len(segments) != 1 {
		t.Fatalf("expected exactly 1 segment for continuous tone, got %d", len(segments))
	}

	seg := segments[0]
	const totalDuration = 5.0
	const tolerance = 0.5

	if seg.StartTime > tolerance {
		t.Errorf("segment starts at %.3f, expected near 0", seg.StartTime)
	}
	if seg.EndTime < totalDuration-tolerance {
		t.Errorf("segment ends at %.3f, expected near %.1f", seg.EndTime, totalDuration)
	}
}

func TestGetNonSilentSegments_AllSilence(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("all-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdRelaxed,
	}

	segments, err := v.GetNonSilentSegments(config)
	if err != nil {
		t.Fatalf("GetNonSilentSegments: %v", err)
	}

	if len(segments) != 0 {
		t.Fatalf("expected 0 segments for all-silence file, got %d: %+v", len(segments), segments)
	}
}

func TestGetNonSilentSegments_ZeroValueConfig(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("silence-middle.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Zero-value config should use defaults (SilenceThresholdModerate + SilenceDurationMedium).
	segments, err := v.GetNonSilentSegments(SilenceConfig{})
	if err != nil {
		t.Fatalf("GetNonSilentSegments with zero config: %v", err)
	}

	if len(segments) < 1 {
		t.Fatalf("expected at least 1 segment with zero-value config, got %d", len(segments))
	}
}

func TestGetNonSilentSegments_NoAudioStream(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-audio.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}

	// Either returns an error or returns segments — no panic is the key assertion.
	segments, err := v.GetNonSilentSegments(config)
	if err != nil {
		t.Logf("GetNonSilentSegments on no-audio.mp4 returned error (acceptable): %v", err)
		return
	}
	t.Logf("GetNonSilentSegments on no-audio.mp4 returned %d segments (no error)", len(segments))
}
