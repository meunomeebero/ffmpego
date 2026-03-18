package ffutil

import (
	"math"
	"testing"
)

func TestParseSilenceOutput_RealOutput(t *testing.T) {
	// Real ffmpeg silencedetect output
	output := `[silencedetect @ 0x7f9] silence_start: 1.504
[silencedetect @ 0x7f9] silence_end: 3.200 | silence_duration: 1.696
[silencedetect @ 0x7f9] silence_start: 5.800
[silencedetect @ 0x7f9] silence_end: 7.100 | silence_duration: 1.300`

	starts, ends := ParseSilenceOutput(output)

	if len(starts) != 2 {
		t.Fatalf("expected 2 starts, got %d", len(starts))
	}
	if len(ends) != 2 {
		t.Fatalf("expected 2 ends, got %d", len(ends))
	}
	assertFloat(t, starts[0], 1.504, "starts[0]")
	assertFloat(t, starts[1], 5.800, "starts[1]")
	assertFloat(t, ends[0], 3.200, "ends[0]")
	assertFloat(t, ends[1], 7.100, "ends[1]")
}

func TestParseSilenceOutput_Empty(t *testing.T) {
	starts, ends := ParseSilenceOutput("")
	if len(starts) != 0 || len(ends) != 0 {
		t.Fatalf("expected empty, got %d starts and %d ends", len(starts), len(ends))
	}
}

func TestParseSilenceOutput_NoSilence(t *testing.T) {
	output := `size=N/A time=00:01:30.00 bitrate=N/A speed=25.3x`
	starts, ends := ParseSilenceOutput(output)
	if len(starts) != 0 || len(ends) != 0 {
		t.Fatalf("expected empty, got %d starts and %d ends", len(starts), len(ends))
	}
}

func TestParseSilenceOutput_SilenceExtendsToEnd(t *testing.T) {
	output := `[silencedetect @ 0x7f9] silence_start: 0.000
[silencedetect @ 0x7f9] silence_end: 2.000 | silence_duration: 2.000
[silencedetect @ 0x7f9] silence_start: 8.500`

	starts, ends := ParseSilenceOutput(output)
	if len(starts) != 2 {
		t.Fatalf("expected 2 starts, got %d", len(starts))
	}
	if len(ends) != 1 {
		t.Fatalf("expected 1 end, got %d", len(ends))
	}
}

func TestBuildNonSilentSegments_NoSilence(t *testing.T) {
	segs := BuildNonSilentSegments(nil, nil, 10.0, 0.5)
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	assertFloat(t, segs[0].StartTime, 0, "start")
	assertFloat(t, segs[0].EndTime, 10.0, "end")
	assertFloat(t, segs[0].Duration, 10.0, "duration")
}

func TestBuildNonSilentSegments_SilenceAtStart(t *testing.T) {
	starts := []float64{0.0}
	ends := []float64{2.0}
	segs := BuildNonSilentSegments(starts, ends, 10.0, 0.5)

	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	assertFloat(t, segs[0].StartTime, 2.0, "start")
	assertFloat(t, segs[0].EndTime, 10.0, "end")
}

func TestBuildNonSilentSegments_SilenceAtEnd(t *testing.T) {
	starts := []float64{8.0}
	ends := []float64{10.0}
	segs := BuildNonSilentSegments(starts, ends, 10.0, 0.5)

	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	assertFloat(t, segs[0].StartTime, 0, "start")
	assertFloat(t, segs[0].EndTime, 8.0, "end")
}

func TestBuildNonSilentSegments_SilenceExtendsToEnd(t *testing.T) {
	// More starts than ends — silence extends to end of file
	starts := []float64{2.0, 6.0}
	ends := []float64{4.0}
	segs := BuildNonSilentSegments(starts, ends, 10.0, 0.5)

	if len(segs) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segs))
	}
	assertFloat(t, segs[0].StartTime, 0, "seg0 start")
	assertFloat(t, segs[0].EndTime, 2.0, "seg0 end")
	assertFloat(t, segs[1].StartTime, 4.0, "seg1 start")
	assertFloat(t, segs[1].EndTime, 6.0, "seg1 end")
}

func TestBuildNonSilentSegments_MultipleSilences(t *testing.T) {
	starts := []float64{1.0, 4.0, 7.0}
	ends := []float64{2.0, 5.0, 8.0}
	segs := BuildNonSilentSegments(starts, ends, 10.0, 0.5)

	if len(segs) != 4 {
		t.Fatalf("expected 4 segments, got %d", len(segs))
	}
	assertFloat(t, segs[0].StartTime, 0, "seg0 start")
	assertFloat(t, segs[0].EndTime, 1.0, "seg0 end")
	assertFloat(t, segs[1].StartTime, 2.0, "seg1 start")
	assertFloat(t, segs[1].EndTime, 4.0, "seg1 end")
	assertFloat(t, segs[2].StartTime, 5.0, "seg2 start")
	assertFloat(t, segs[2].EndTime, 7.0, "seg2 end")
	assertFloat(t, segs[3].StartTime, 8.0, "seg3 start")
	assertFloat(t, segs[3].EndTime, 10.0, "seg3 end")
}

func TestBuildNonSilentSegments_ShortSegmentsFiltered(t *testing.T) {
	// Non-silent gap between silences is only 0.3s — should be filtered out
	starts := []float64{1.0, 3.3}
	ends := []float64{3.0, 5.0}
	segs := BuildNonSilentSegments(starts, ends, 10.0, 0.5)

	if len(segs) != 2 {
		t.Fatalf("expected 2 segments (short one filtered), got %d", len(segs))
	}
	assertFloat(t, segs[0].StartTime, 0, "seg0 start")
	assertFloat(t, segs[0].EndTime, 1.0, "seg0 end")
	assertFloat(t, segs[1].StartTime, 5.0, "seg1 start")
	assertFloat(t, segs[1].EndTime, 10.0, "seg1 end")
}

func TestBuildNonSilentSegments_EntireFileIsSilence(t *testing.T) {
	starts := []float64{0.0}
	ends := []float64{}
	segs := BuildNonSilentSegments(starts, ends, 10.0, 0.5)

	if len(segs) != 0 {
		t.Fatalf("expected 0 segments for entirely silent file, got %d", len(segs))
	}
}

func TestBuildNonSilentSegments_ZeroDuration(t *testing.T) {
	segs := BuildNonSilentSegments(nil, nil, 0, 0.5)
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	assertFloat(t, segs[0].Duration, 0, "duration")
}

func TestBuildNonSilentSegments_DurationField(t *testing.T) {
	starts := []float64{3.0}
	ends := []float64{5.0}
	segs := BuildNonSilentSegments(starts, ends, 10.0, 0.5)

	for _, seg := range segs {
		expected := seg.EndTime - seg.StartTime
		assertFloat(t, seg.Duration, expected, "duration == end - start")
	}
}

func assertFloat(t *testing.T, got, want float64, name string) {
	t.Helper()
	if math.Abs(got-want) > 0.001 {
		t.Errorf("%s: got %.3f, want %.3f", name, got, want)
	}
}
