package audio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractSegment_StreamCopy(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "segment.wav")
	if err := a.ExtractSegment(outputPath, 1.0, 3.0, nil); err != nil {
		t.Fatalf("ExtractSegment: %v", err)
	}

	assertValidMedia(t, outputPath)
	assertDuration(t, outputPath, 2.0, 0.2)
}

func TestExtractSegment_WithConfig(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "segment.mp3")
	config := &ConvertConfig{Codec: CodecMP3, Bitrate: 128}
	if err := a.ExtractSegment(outputPath, 1.0, 3.0, config); err != nil {
		t.Fatalf("ExtractSegment with config: %v", err)
	}

	assertValidMedia(t, outputPath)
}

func TestExtractSegment_OutputDirCreated(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "subdir", "deep", "segment.wav")
	if err := a.ExtractSegment(outputPath, 0.0, 2.0, nil); err != nil {
		t.Fatalf("ExtractSegment (deep dir): %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file not found after creating deep dirs: %v", err)
	}
}

func TestConcatenateSegments_Basic(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tmpDir := t.TempDir()

	seg1 := filepath.Join(tmpDir, "seg1.wav")
	seg2 := filepath.Join(tmpDir, "seg2.wav")
	seg3 := filepath.Join(tmpDir, "seg3.wav")

	if err := a.ExtractSegment(seg1, 0.0, 1.5, nil); err != nil {
		t.Fatalf("ExtractSegment seg1: %v", err)
	}
	if err := a.ExtractSegment(seg2, 1.5, 3.0, nil); err != nil {
		t.Fatalf("ExtractSegment seg2: %v", err)
	}
	if err := a.ExtractSegment(seg3, 3.0, 4.5, nil); err != nil {
		t.Fatalf("ExtractSegment seg3: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "concat.wav")
	if err := ConcatenateSegments([]string{seg1, seg2, seg3}, outputPath, nil); err != nil {
		t.Fatalf("ConcatenateSegments: %v", err)
	}

	assertDuration(t, outputPath, 4.5, 0.3)
}

func TestConcatenateSegments_SingleSegment(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tmpDir := t.TempDir()
	seg1 := filepath.Join(tmpDir, "seg1.wav")
	if err := a.ExtractSegment(seg1, 0.0, 2.0, nil); err != nil {
		t.Fatalf("ExtractSegment: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "out.wav")
	if err := ConcatenateSegments([]string{seg1}, outputPath, nil); err != nil {
		t.Fatalf("ConcatenateSegments with single segment: %v", err)
	}

	assertValidMedia(t, outputPath)
}

func TestConcatenateSegments_EmptyList(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "out.wav")
	err := ConcatenateSegments([]string{}, outputPath, nil)
	if err == nil {
		t.Fatal("ConcatenateSegments with empty list: expected error, got nil")
	}
}

func TestConcatenateSegments_MissingFile(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "out.wav")
	err := ConcatenateSegments([]string{"/nonexistent/segment.wav"}, outputPath, nil)
	if err == nil {
		t.Fatal("ConcatenateSegments with missing file: expected error, got nil")
	}
}
