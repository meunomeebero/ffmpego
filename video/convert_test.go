package video

import (
	"math"
	"path/filepath"
	"testing"
)

func TestConvert_BasicCodec(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out.mp4")
	config := ConvertConfig{
		VideoCodec: CodecH264,
		AudioCodec: CodecAAC,
		Quality:    28,
		Preset:     PresetUltrafast,
	}
	if err := v.Convert(out, config); err != nil {
		t.Fatalf("Convert: %v", err)
	}

	assertValidMedia(t, out)
	assertDuration(t, out, 5.0, 0.5)
}

func TestConvert_AspectRatio(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out.mp4")
	config := ConvertConfig{
		AspectRatio: AspectRatio16x9,
		VideoCodec:  CodecH264,
		Quality:     28,
		Preset:      PresetUltrafast,
	}
	if err := v.Convert(out, config); err != nil {
		t.Fatalf("Convert with aspect ratio: %v", err)
	}

	assertValidMedia(t, out)

	outV, err := New(out)
	if err != nil {
		t.Fatalf("New (output): %v", err)
	}
	info, err := outV.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (output): %v", err)
	}

	if info.Height == 0 {
		t.Fatalf("output Height is 0")
	}
	ratio := float64(info.Width) / float64(info.Height)
	want := 16.0 / 9.0
	if math.Abs(ratio-want) > 0.1 {
		t.Errorf("aspect ratio = %.4f, want ~%.4f (tolerance 0.1); Width=%d Height=%d",
			ratio, want, info.Width, info.Height)
	}

	if info.Width%2 != 0 {
		t.Errorf("Width %d is not even", info.Width)
	}
	if info.Height%2 != 0 {
		t.Errorf("Height %d is not even", info.Height)
	}
}

func TestConvert_OutputDirCreated(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	out := filepath.Join(t.TempDir(), "new", "dir", "out.mp4")
	config := ConvertConfig{
		VideoCodec: CodecH264,
		AudioCodec: CodecAAC,
		Quality:    28,
		Preset:     PresetUltrafast,
	}
	if err := v.Convert(out, config); err != nil {
		t.Fatalf("Convert to nested dir: %v", err)
	}

	assertValidMedia(t, out)
}
