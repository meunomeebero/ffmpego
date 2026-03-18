package video

import (
	"path/filepath"
	"testing"
)

func TestExtractSegment_StreamCopy(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	out := filepath.Join(t.TempDir(), "segment.mp4")
	if err := v.ExtractSegment(out, 1.0, 3.0, nil); err != nil {
		t.Fatalf("ExtractSegment: %v", err)
	}

	assertValidMedia(t, out)
	assertDuration(t, out, 2.0, 0.5)
}

func TestExtractSegment_WithConfig(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	config := &ConvertConfig{
		VideoCodec: CodecH264,
		AudioCodec: CodecAAC,
		Quality:    28,
		Preset:     PresetUltrafast,
	}

	out := filepath.Join(t.TempDir(), "segment.mp4")
	if err := v.ExtractSegment(out, 1.0, 3.0, config); err != nil {
		t.Fatalf("ExtractSegment with config: %v", err)
	}

	assertValidMedia(t, out)
}

func TestExtractSegment_OutputDirCreated(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	out := filepath.Join(t.TempDir(), "subdir", "deep", "segment.mp4")
	if err := v.ExtractSegment(out, 1.0, 3.0, nil); err != nil {
		t.Fatalf("ExtractSegment to nested dir: %v", err)
	}

	assertValidMedia(t, out)
}

func TestConcatenateSegments_Basic(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tmpDir := t.TempDir()
	seg1 := filepath.Join(tmpDir, "seg1.mp4")
	seg2 := filepath.Join(tmpDir, "seg2.mp4")
	seg3 := filepath.Join(tmpDir, "seg3.mp4")

	if err := v.ExtractSegment(seg1, 0.0, 1.5, nil); err != nil {
		t.Fatalf("ExtractSegment seg1: %v", err)
	}
	if err := v.ExtractSegment(seg2, 1.5, 3.0, nil); err != nil {
		t.Fatalf("ExtractSegment seg2: %v", err)
	}
	if err := v.ExtractSegment(seg3, 3.0, 4.5, nil); err != nil {
		t.Fatalf("ExtractSegment seg3: %v", err)
	}

	out := filepath.Join(tmpDir, "concat.mp4")
	if err := ConcatenateSegments([]string{seg1, seg2, seg3}, out, nil); err != nil {
		t.Fatalf("ConcatenateSegments: %v", err)
	}

	assertValidMedia(t, out)
	assertDuration(t, out, 4.5, 0.5)
}

func TestConcatenateSegments_SingleSegment(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tmpDir := t.TempDir()
	seg := filepath.Join(tmpDir, "seg.mp4")
	if err := v.ExtractSegment(seg, 0.0, 2.0, nil); err != nil {
		t.Fatalf("ExtractSegment: %v", err)
	}

	out := filepath.Join(tmpDir, "out.mp4")
	if err := ConcatenateSegments([]string{seg}, out, nil); err != nil {
		t.Fatalf("ConcatenateSegments single: %v", err)
	}

	assertValidMedia(t, out)
}

func TestConcatenateSegments_EmptyList(t *testing.T) {
	t.Parallel()

	out := filepath.Join(t.TempDir(), "out.mp4")
	err := ConcatenateSegments([]string{}, out, nil)
	if err == nil {
		t.Fatal("expected error for empty segment list, got nil")
	}
}

func TestConcatenateSegments_MissingFile(t *testing.T) {
	t.Parallel()

	out := filepath.Join(t.TempDir(), "out.mp4")
	err := ConcatenateSegments([]string{"/nonexistent/file.mp4"}, out, nil)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
