package audio

import (
	"os"
	"testing"
)

func TestNew_ValidFile(t *testing.T) {
	t.Parallel()

	path := fixture("no-silence.wav")
	a, err := New(path)
	if err != nil {
		t.Fatalf("New(%q) returned unexpected error: %v", path, err)
	}
	if a.Path() != path {
		t.Fatalf("Path() = %q, want %q", a.Path(), path)
	}
}

func TestNew_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := New("/nonexistent/file.wav")
	if err == nil {
		t.Fatal("New(\"/nonexistent/file.wav\") expected an error, got nil")
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
	f, err := os.CreateTemp(restrictedDir, "file_*.wav")
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

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	info, err := a.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}

	if info.SampleRate != 44100 {
		t.Errorf("SampleRate = %d, want 44100", info.SampleRate)
	}
	if info.Channels != 1 {
		t.Errorf("Channels = %d, want 1", info.Channels)
	}

	const wantDuration = 5.0
	const durationTolerance = 0.5
	diff := info.Duration - wantDuration
	if diff < 0 {
		diff = -diff
	}
	if diff > durationTolerance {
		t.Errorf("Duration = %.3f, want ~%.1f (tolerance %.1f)", info.Duration, wantDuration, durationTolerance)
	}

	if info.FileSizeBytes <= 0 {
		t.Errorf("FileSizeBytes = %d, want > 0", info.FileSizeBytes)
	}
}

func TestGetInfo_Cached(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	info1, err := a.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (first call): %v", err)
	}

	info2, err := a.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (second call): %v", err)
	}

	if *info1 != *info2 {
		t.Errorf("GetInfo returned different values on second call: %+v vs %+v", info1, info2)
	}
}

func TestGetInfo_CopyNotShared(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	info1, err := a.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (first call): %v", err)
	}

	originalSampleRate := info1.SampleRate
	info1.SampleRate = 99999

	info2, err := a.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (second call): %v", err)
	}

	if info2.SampleRate != originalSampleRate {
		t.Errorf("GetInfo second call reflected mutation: SampleRate = %d, want %d",
			info2.SampleRate, originalSampleRate)
	}
}
