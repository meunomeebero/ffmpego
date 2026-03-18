package audio

import (
	"path/filepath"
	"testing"
)

func TestConvert_BasicCodec(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "out.mp3")
	config := ConvertConfig{Codec: CodecMP3, Bitrate: 192}
	if err := a.Convert(outputPath, config); err != nil {
		t.Fatalf("Convert: %v", err)
	}

	assertValidMedia(t, outputPath)
	assertDuration(t, outputPath, 5.0, 0.5)
}

func TestConvert_SampleRate(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "out.aac")
	config := ConvertConfig{SampleRate: SampleRate48000, Codec: CodecAAC}
	if err := a.Convert(outputPath, config); err != nil {
		t.Fatalf("Convert: %v", err)
	}

	out, err := New(outputPath)
	if err != nil {
		t.Fatalf("New (output): %v", err)
	}

	info, err := out.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}

	if info.SampleRate != SampleRate48000 {
		t.Errorf("SampleRate = %d, want %d", info.SampleRate, SampleRate48000)
	}
}

func TestConvert_OutputDirCreated(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "new", "dir", "out.mp3")
	config := ConvertConfig{Codec: CodecMP3, Bitrate: 128}
	if err := a.Convert(outputPath, config); err != nil {
		t.Fatalf("Convert (deep dir): %v", err)
	}

	assertValidMedia(t, outputPath)
}
