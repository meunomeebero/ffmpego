package video

import (
	"os"
	"testing"
)

func TestNew_ValidFile(t *testing.T) {
	t.Parallel()

	path := fixture("no-silence.mp4")
	v, err := New(path)
	if err != nil {
		t.Fatalf("New(%q) returned unexpected error: %v", path, err)
	}
	if v.Path() != path {
		t.Fatalf("Path() = %q, want %q", v.Path(), path)
	}
}

func TestNew_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := New("/nonexistent/file.mp4")
	if err == nil {
		t.Fatal("New(\"/nonexistent/file.mp4\") expected an error, got nil")
	}
}

func TestNew_PermissionDenied(t *testing.T) {
	t.Parallel()

	if os.Getuid() == 0 {
		t.Skip("skipping permission test: running as root")
	}

	// Create a directory with no permissions so that stat on files inside it fails.
	restrictedDir, err := os.MkdirTemp(testFixtureDir, "noperm_dir_*")
	if err != nil {
		t.Fatalf("failed to create restricted dir: %v", err)
	}
	defer func() {
		// Restore permissions before cleanup so os.RemoveAll can delete it.
		os.Chmod(restrictedDir, 0755)
		os.RemoveAll(restrictedDir)
	}()

	// Create a file inside before locking down the directory.
	f, err := os.CreateTemp(restrictedDir, "file_*.mp4")
	if err != nil {
		t.Fatalf("failed to create file inside restricted dir: %v", err)
	}
	path := f.Name()
	f.Close()

	// Remove all permissions from the directory so os.Stat on `path` fails.
	if err := os.Chmod(restrictedDir, 0000); err != nil {
		t.Fatalf("failed to chmod dir: %v", err)
	}

	_, err = New(path)
	if err == nil {
		t.Fatalf("New(%q) expected an error for inaccessible file, got nil", path)
	}
}

func TestGetInfo_ReturnsCorrectMetadata(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	info, err := v.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}

	if info.Width != 320 {
		t.Errorf("Width = %d, want 320", info.Width)
	}
	if info.Height != 240 {
		t.Errorf("Height = %d, want 240", info.Height)
	}

	const wantFrameRate = 15.0
	const frameRateTolerance = 1.0
	frDiff := info.FrameRate - wantFrameRate
	if frDiff < 0 {
		frDiff = -frDiff
	}
	if frDiff > frameRateTolerance {
		t.Errorf("FrameRate = %.3f, want ~%.1f (tolerance %.1f)", info.FrameRate, wantFrameRate, frameRateTolerance)
	}

	const wantDuration = 5.0
	const durationTolerance = 0.5
	durDiff := info.Duration - wantDuration
	if durDiff < 0 {
		durDiff = -durDiff
	}
	if durDiff > durationTolerance {
		t.Errorf("Duration = %.3f, want ~%.1f (tolerance %.1f)", info.Duration, wantDuration, durationTolerance)
	}

	if info.VideoCodec != "h264" {
		t.Errorf("VideoCodec = %q, want %q", info.VideoCodec, "h264")
	}

	if info.FileSizeBytes <= 0 {
		t.Errorf("FileSizeBytes = %d, want > 0", info.FileSizeBytes)
	}
}

func TestGetInfo_Cached(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	info1, err := v.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (first call): %v", err)
	}

	info2, err := v.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (second call): %v", err)
	}

	if *info1 != *info2 {
		t.Errorf("GetInfo returned different values on second call: %+v vs %+v", info1, info2)
	}
}

func TestGetInfo_CopyNotShared(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	info1, err := v.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (first call): %v", err)
	}

	originalDuration := info1.Duration
	originalWidth := info1.Width
	info1.Duration = 999
	info1.Width = 1

	info2, err := v.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (second call): %v", err)
	}

	if info2.Duration != originalDuration {
		t.Errorf("GetInfo second call reflected Duration mutation: got %.3f, want %.3f",
			info2.Duration, originalDuration)
	}
	if info2.Width != originalWidth {
		t.Errorf("GetInfo second call reflected Width mutation: got %d, want %d",
			info2.Width, originalWidth)
	}
}
