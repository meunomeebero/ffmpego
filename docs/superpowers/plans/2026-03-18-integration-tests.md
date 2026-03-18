# Integration Tests Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add 65 integration tests covering all public API functions, edge cases, and parallel execution safety for both audio/ and video/ packages.

**Architecture:** Each package gets a `testmain_test.go` that generates media fixtures via ffmpeg at test time (no binaries in git). Test files mirror source files. Shared helpers (`assertValidMedia`, `assertDuration`) live in `testmain_test.go`. All tests use `t.Parallel()` and `t.TempDir()` for isolation.

**Tech Stack:** Go testing, ffmpeg/ffprobe CLI, no external test dependencies.

**Spec:** `docs/superpowers/specs/2026-03-18-integration-tests-design.md`

---

## Chunk 1: Audio Foundation (TestMain + Info + Silence)

### Task 1: Audio TestMain and Fixture Generation

**Files:**
- Create: `audio/testmain_test.go`

- [ ] **Step 1: Create `audio/testmain_test.go` with TestMain, fixture generation, and test helpers**

```go
package audio

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var testFixtureDir string

func TestMain(m *testing.M) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Println("skipping integration tests: ffmpeg not found in PATH")
		os.Exit(0)
	}

	dir, err := os.MkdirTemp("", "ffmpego_audio_test_*")
	if err != nil {
		log.Fatal(err)
	}
	testFixtureDir = dir

	if err := generateAudioFixtures(dir); err != nil {
		os.RemoveAll(dir)
		log.Fatalf("failed to generate test fixtures: %v", err)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func generateAudioFixtures(dir string) error {
	fixtures := []struct {
		name string
		args []string
	}{
		{
			"no-silence.wav",
			[]string{"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=5", "-ac", "1", "-y"},
		},
		{
			"all-silence.wav",
			[]string{"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono", "-t", "5", "-y"},
		},
		{
			"short.wav",
			[]string{"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=0.3", "-ac", "1", "-y"},
		},
	}

	// Simple fixtures (single source)
	for _, f := range fixtures {
		args := append(f.args, filepath.Join(dir, f.name))
		cmd := exec.Command("ffmpeg", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("generating %s: %w\n%s", f.name, err, out)
		}
	}

	// Complex fixtures (concat filtergraph)
	concatFixtures := []struct {
		name   string
		args   []string
	}{
		{
			"silence-start.wav",
			[]string{
				"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono", "-t", "2",
				"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=3",
				"-filter_complex", "[0][1]concat=n=2:v=0:a=1", "-y",
			},
		},
		{
			"silence-middle.wav",
			[]string{
				"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=2",
				"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono", "-t", "2",
				"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=2",
				"-filter_complex", "[0][1][2]concat=n=3:v=0:a=1", "-y",
			},
		},
		{
			"silence-end.wav",
			[]string{
				"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=3",
				"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono", "-t", "2",
				"-filter_complex", "[0][1]concat=n=2:v=0:a=1", "-y",
			},
		},
	}

	for _, f := range concatFixtures {
		args := append(f.args, filepath.Join(dir, f.name))
		cmd := exec.Command("ffmpeg", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("generating %s: %w\n%s", f.name, err, out)
		}
	}

	return nil
}

func fixture(name string) string {
	return filepath.Join(testFixtureDir, name)
}

func assertValidMedia(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("output file not accessible: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output file is empty")
	}
	cmd := exec.Command("ffprobe", "-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0", path)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ffprobe failed on output: %v", err)
	}
	dur, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || dur <= 0 {
		t.Fatalf("output has invalid duration: %q", strings.TrimSpace(string(out)))
	}
}

func assertDuration(t *testing.T, path string, expected, tolerance float64) {
	t.Helper()
	cmd := exec.Command("ffprobe", "-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0", path)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ffprobe failed: %v", err)
	}
	actual, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		t.Fatalf("failed to parse duration: %v", err)
	}
	if math.Abs(actual-expected) > tolerance {
		t.Errorf("duration %.2fs outside tolerance: expected %.2f ± %.2f", actual, expected, tolerance)
	}
}
```

- [ ] **Step 2: Verify fixtures generate correctly**

Run: `cd /Users/robertojunior/Documents/dev/me/bero/tools/ffmpego && go test ./audio/ -run TestMain -v -count=1`
Expected: exits 0 (no tests match "TestMain" but TestMain runs and generates fixtures without error)

- [ ] **Step 3: Commit**

```bash
git add audio/testmain_test.go
git commit -m "test(audio): add TestMain with fixture generation and test helpers"
```

---

### Task 2: Audio Info Tests

**Files:**
- Create: `audio/info_test.go`

- [ ] **Step 1: Create `audio/info_test.go`**

```go
package audio

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_ValidFile(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if a.Path() != fixture("no-silence.wav") {
		t.Errorf("Path() = %q, want %q", a.Path(), fixture("no-silence.wav"))
	}
}

func TestNew_FileNotFound(t *testing.T) {
	t.Parallel()
	_, err := New("/nonexistent/file.wav")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestNew_PermissionDenied(t *testing.T) {
	t.Parallel()
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "noperm.wav")
	os.WriteFile(path, []byte("fake"), 0000)
	_, err := New(path)
	if err == nil {
		t.Fatal("expected error for unreadable file, got nil")
	}
}

func TestGetInfo_ReturnsCorrectMetadata(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	info, err := a.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo() error: %v", err)
	}
	if info.SampleRate != 44100 {
		t.Errorf("SampleRate = %d, want 44100", info.SampleRate)
	}
	if info.Channels != 1 {
		t.Errorf("Channels = %d, want 1", info.Channels)
	}
	if info.Duration < 4.5 || info.Duration > 5.5 {
		t.Errorf("Duration = %.2f, want ~5.0", info.Duration)
	}
	if info.FileSizeBytes <= 0 {
		t.Errorf("FileSizeBytes = %d, want > 0", info.FileSizeBytes)
	}
}

func TestGetInfo_Cached(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	info1, err := a.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo() #1 error: %v", err)
	}
	info2, err := a.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo() #2 error: %v", err)
	}
	if info1.Duration != info2.Duration || info1.SampleRate != info2.SampleRate {
		t.Error("cached GetInfo() returned different values")
	}
}

func TestGetInfo_CopyNotShared(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	info1, _ := a.GetInfo()
	info1.Duration = 999.0
	info1.SampleRate = 1

	info2, _ := a.GetInfo()
	if info2.Duration == 999.0 {
		t.Error("modifying returned Info corrupted the cache (Duration)")
	}
	if info2.SampleRate == 1 {
		t.Error("modifying returned Info corrupted the cache (SampleRate)")
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./audio/ -run "TestNew|TestGetInfo" -v -count=1`
Expected: 6 PASS

- [ ] **Step 3: Commit**

```bash
git add audio/info_test.go
git commit -m "test(audio): add constructor and GetInfo integration tests"
```

---

### Task 3: Audio Silence Detection Tests

**Files:**
- Create: `audio/silence_test.go`

- [ ] **Step 1: Create `audio/silence_test.go`**

```go
package audio

import (
	"testing"
)

func TestGetNonSilentSegments_SilenceAtStart(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("silence-start.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}
	segments, err := a.GetNonSilentSegments(config)
	if err != nil {
		t.Fatalf("GetNonSilentSegments() error: %v", err)
	}
	if len(segments) == 0 {
		t.Fatal("expected at least 1 segment, got 0")
	}
	// First segment should start after the 2s silence
	if segments[0].StartTime < 1.0 {
		t.Errorf("first segment starts at %.2f, expected after silence (~2s)", segments[0].StartTime)
	}
}

func TestGetNonSilentSegments_SilenceMiddle(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("silence-middle.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}
	segments, err := a.GetNonSilentSegments(config)
	if err != nil {
		t.Fatalf("GetNonSilentSegments() error: %v", err)
	}
	if len(segments) < 2 {
		t.Fatalf("expected at least 2 segments, got %d", len(segments))
	}
}

func TestGetNonSilentSegments_SilenceAtEnd(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("silence-end.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}
	segments, err := a.GetNonSilentSegments(config)
	if err != nil {
		t.Fatalf("GetNonSilentSegments() error: %v", err)
	}
	if len(segments) == 0 {
		t.Fatal("expected at least 1 segment, got 0")
	}
	info, _ := a.GetInfo()
	last := segments[len(segments)-1]
	if last.EndTime > info.Duration-1.0 {
		t.Errorf("last segment ends at %.2f, expected before trailing silence (~3s)", last.EndTime)
	}
}

func TestGetNonSilentSegments_NoSilence(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}
	segments, err := a.GetNonSilentSegments(config)
	if err != nil {
		t.Fatalf("GetNonSilentSegments() error: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment (full file), got %d", len(segments))
	}
	if segments[0].StartTime > 0.1 {
		t.Errorf("segment should start near 0, got %.2f", segments[0].StartTime)
	}
}

func TestGetNonSilentSegments_AllSilence(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("all-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdRelaxed,
	}
	segments, err := a.GetNonSilentSegments(config)
	if err != nil {
		t.Fatalf("GetNonSilentSegments() error: %v", err)
	}
	if len(segments) != 0 {
		t.Errorf("expected 0 segments for all-silence file, got %d", len(segments))
	}
}

func TestGetNonSilentSegments_ZeroValueConfig(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("silence-middle.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	// Zero-value config should apply defaults
	segments, err := a.GetNonSilentSegments(SilenceConfig{})
	if err != nil {
		t.Fatalf("GetNonSilentSegments() with zero config error: %v", err)
	}
	if len(segments) < 2 {
		t.Fatalf("expected at least 2 segments with default config, got %d", len(segments))
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./audio/ -run "TestGetNonSilentSegments" -v -count=1`
Expected: 6 PASS

- [ ] **Step 3: Commit**

```bash
git add audio/silence_test.go
git commit -m "test(audio): add silence detection integration tests"
```

---

## Chunk 2: Audio Remaining (Segment + Convert + RemoveSilence + Parallel)

### Task 4: Audio Segment Tests

**Files:**
- Create: `audio/segment_test.go`

- [ ] **Step 1: Create `audio/segment_test.go`**

```go
package audio

import (
	"path/filepath"
	"testing"
)

func TestExtractSegment_StreamCopy(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "segment.wav")
	err = a.ExtractSegment(out, 1.0, 3.0, nil)
	if err != nil {
		t.Fatalf("ExtractSegment() error: %v", err)
	}
	assertValidMedia(t, out)
	assertDuration(t, out, 2.0, 0.5)
}

func TestExtractSegment_WithConfig(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "segment.mp3")
	config := &ConvertConfig{
		Codec:   CodecMP3,
		Bitrate: 128,
	}
	err = a.ExtractSegment(out, 1.0, 3.0, config)
	if err != nil {
		t.Fatalf("ExtractSegment() with config error: %v", err)
	}
	assertValidMedia(t, out)
	assertDuration(t, out, 2.0, 0.5)
}

func TestExtractSegment_OutputDirCreated(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "subdir", "deep", "segment.wav")
	err = a.ExtractSegment(out, 0.0, 2.0, nil)
	if err != nil {
		t.Fatalf("ExtractSegment() error: %v", err)
	}
	assertValidMedia(t, out)
}

func TestConcatenateSegments_Basic(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	dir := t.TempDir()
	seg1 := filepath.Join(dir, "s1.wav")
	seg2 := filepath.Join(dir, "s2.wav")
	seg3 := filepath.Join(dir, "s3.wav")
	a.ExtractSegment(seg1, 0.0, 1.5, nil)
	a.ExtractSegment(seg2, 1.5, 3.0, nil)
	a.ExtractSegment(seg3, 3.0, 4.5, nil)

	out := filepath.Join(dir, "concat.wav")
	err = ConcatenateSegments([]string{seg1, seg2, seg3}, out, nil)
	if err != nil {
		t.Fatalf("ConcatenateSegments() error: %v", err)
	}
	assertValidMedia(t, out)
	assertDuration(t, out, 4.5, 0.5)
}

func TestConcatenateSegments_SingleSegment(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	dir := t.TempDir()
	seg := filepath.Join(dir, "only.wav")
	a.ExtractSegment(seg, 0.0, 2.0, nil)

	out := filepath.Join(dir, "concat.wav")
	err = ConcatenateSegments([]string{seg}, out, nil)
	if err != nil {
		t.Fatalf("ConcatenateSegments() single segment error: %v", err)
	}
	assertValidMedia(t, out)
}

func TestConcatenateSegments_EmptyList(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "out.wav")
	err := ConcatenateSegments([]string{}, out, nil)
	if err == nil {
		t.Fatal("expected error for empty segment list, got nil")
	}
}

func TestConcatenateSegments_MissingFile(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "out.wav")
	err := ConcatenateSegments([]string{"/nonexistent/file.wav"}, out, nil)
	if err == nil {
		t.Fatal("expected error for missing segment file, got nil")
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./audio/ -run "TestExtractSegment|TestConcatenateSegments" -v -count=1`
Expected: 7 PASS

- [ ] **Step 3: Commit**

```bash
git add audio/segment_test.go
git commit -m "test(audio): add segment extraction and concatenation tests"
```

---

### Task 5: Audio Convert Tests

**Files:**
- Create: `audio/convert_test.go`

- [ ] **Step 1: Create `audio/convert_test.go`**

```go
package audio

import (
	"path/filepath"
	"testing"
)

func TestConvert_BasicCodec(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "converted.mp3")
	err = a.Convert(out, ConvertConfig{
		Codec:   CodecMP3,
		Bitrate: 192,
	})
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}
	assertValidMedia(t, out)
	assertDuration(t, out, 5.0, 0.5)
}

func TestConvert_SampleRate(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "resampled.wav")
	err = a.Convert(out, ConvertConfig{
		SampleRate: SampleRate48000,
		Codec:      CodecAAC,
	})
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}
	assertValidMedia(t, out)

	// Verify sample rate via fresh GetInfo
	converted, _ := New(out)
	info, err := converted.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo on converted: %v", err)
	}
	if info.SampleRate != 48000 {
		t.Errorf("SampleRate = %d, want 48000", info.SampleRate)
	}
}

func TestConvert_OutputDirCreated(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "new", "dir", "out.mp3")
	err = a.Convert(out, ConvertConfig{Codec: CodecMP3})
	if err != nil {
		t.Fatalf("Convert() error: %v", err)
	}
	assertValidMedia(t, out)
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./audio/ -run "TestConvert" -v -count=1`
Expected: 3 PASS

- [ ] **Step 3: Commit**

```bash
git add audio/convert_test.go
git commit -m "test(audio): add conversion integration tests"
```

---

### Task 6: Audio RemoveSilence and Parallel Tests

**Files:**
- Create: `audio/remove_silence_test.go`

- [ ] **Step 1: Create `audio/remove_silence_test.go`**

```go
package audio

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestRemoveSilence_ProducesShorterOutput(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("silence-middle.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "shorter.wav")
	err = a.RemoveSilence(out, SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	})
	if err != nil {
		t.Fatalf("RemoveSilence() error: %v", err)
	}
	assertValidMedia(t, out)

	inputInfo, _ := a.GetInfo()
	outA, _ := New(out)
	outInfo, _ := outA.GetInfo()
	if outInfo.Duration >= inputInfo.Duration {
		t.Errorf("output duration %.2f should be less than input %.2f", outInfo.Duration, inputInfo.Duration)
	}
}

func TestRemoveSilence_NoSilence_CopiesFile(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "same.wav")
	err = a.RemoveSilence(out, SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	})
	if err != nil {
		t.Fatalf("RemoveSilence() error: %v", err)
	}
	assertValidMedia(t, out)
	assertDuration(t, out, 5.0, 1.0)
}

func TestRemoveSilence_AllSilence(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("all-silence.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "empty.wav")
	err = a.RemoveSilence(out, SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdRelaxed,
	})
	if err == nil {
		t.Fatal("expected error for all-silence file, got nil")
	}
}

func TestRemoveSilence_ShortFile(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("short.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "short-out.wav")
	// Short file (0.3s) - may produce error or valid output depending on segment threshold
	err = a.RemoveSilence(out, SilenceConfig{})
	// Either succeeds or returns a clear error — no panic
	if err != nil {
		t.Logf("RemoveSilence on short file returned error (acceptable): %v", err)
	} else {
		assertValidMedia(t, out)
	}
}

func TestRemoveSilence_TempFilesCleanedUp(t *testing.T) {
	t.Parallel()
	a, err := New(fixture("silence-middle.wav"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "clean.wav")

	// Count ffmpego temp dirs before
	beforeDirs := countTempDirs(t, "ffmpego_silence_")

	err = a.RemoveSilence(out, SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	})
	if err != nil {
		t.Fatalf("RemoveSilence() error: %v", err)
	}

	// Count after — should be same as before
	afterDirs := countTempDirs(t, "ffmpego_silence_")
	if afterDirs > beforeDirs {
		t.Errorf("found %d orphaned temp dirs (before: %d, after: %d)", afterDirs-beforeDirs, beforeDirs, afterDirs)
	}
}

// --- Parallelism tests ---

func TestParallel_SameInput_DifferentOutputs(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	errs := make([]error, 2)
	outs := make([]string, 2)

	for i := 0; i < 2; i++ {
		i := i
		outs[i] = filepath.Join(t.TempDir(), "out.wav")
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, err := New(fixture("silence-middle.wav"))
			if err != nil {
				errs[i] = err
				return
			}
			errs[i] = a.RemoveSilence(outs[i], SilenceConfig{})
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d error: %v", i, err)
		}
		assertValidMedia(t, outs[i])
	}
}

func TestParallel_DifferentInputs(t *testing.T) {
	t.Parallel()
	inputs := []string{
		fixture("silence-start.wav"),
		fixture("silence-end.wav"),
	}
	var wg sync.WaitGroup
	errs := make([]error, len(inputs))
	outs := make([]string, len(inputs))

	for i, input := range inputs {
		i, input := i, input
		outs[i] = filepath.Join(t.TempDir(), "out.wav")
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, err := New(input)
			if err != nil {
				errs[i] = err
				return
			}
			errs[i] = a.RemoveSilence(outs[i], SilenceConfig{})
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d error: %v", i, err)
		}
		assertValidMedia(t, outs[i])
	}
}

func TestParallel_SameOutput_NoPanic(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "shared.wav")

	var barrier sync.WaitGroup
	barrier.Add(1)

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			barrier.Wait() // start at the same time
			a, err := New(fixture("silence-middle.wav"))
			if err != nil {
				return
			}
			a.RemoveSilence(out, SilenceConfig{}) // ignore error, just ensure no panic
		}()
	}
	barrier.Done() // release both goroutines
	wg.Wait()
	// If we got here, no panic occurred
}

func TestParallel_RemoveSilence_Concurrent(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	errs := make([]error, 3)

	for i := 0; i < 3; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, err := New(fixture("silence-middle.wav"))
			if err != nil {
				errs[i] = err
				return
			}
			out := filepath.Join(t.TempDir(), "concurrent.wav")
			errs[i] = a.RemoveSilence(out, SilenceConfig{})
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("concurrent RemoveSilence %d failed: %v", i, err)
		}
	}
}

func TestParallel_TempFilesIsolated(t *testing.T) {
	t.Parallel()
	before := countTempDirs(t, "ffmpego_silence_")

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, _ := New(fixture("silence-middle.wav"))
			out := filepath.Join(t.TempDir(), "isolated.wav")
			a.RemoveSilence(out, SilenceConfig{})
		}()
	}
	wg.Wait()

	after := countTempDirs(t, "ffmpego_silence_")
	if after > before {
		t.Errorf("orphaned temp dirs found: before=%d, after=%d", before, after)
	}
}

func countTempDirs(t *testing.T, prefix string) int {
	t.Helper()
	entries, err := os.ReadDir(os.TempDir())
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			count++
		}
	}
	return count
}
```

- [ ] **Step 2: Run all audio tests**

Run: `go test ./audio/ -v -count=1`
Expected: 32 PASS

- [ ] **Step 3: Commit**

```bash
git add audio/remove_silence_test.go
git commit -m "test(audio): add RemoveSilence and parallel execution tests"
```

---

## Chunk 3: Video Foundation (TestMain + Info + Silence)

### Task 7: Video TestMain and Fixture Generation

**Files:**
- Create: `video/testmain_test.go`

Mirrors audio `testmain_test.go` but generates MP4 fixtures with video+audio using `testsrc2` and `-filter_complex`. Uses `libx264 -preset ultrafast -crf 28` for fast encoding. Includes the same `assertValidMedia`, `assertDuration`, `fixture`, and `countTempDirs` helpers.

Key differences from audio:
- `generateVideoFixtures(dir)` creates `.mp4` files with both video and audio streams
- Adds `no-audio.mp4` fixture using `-an` flag
- Uses `testsrc2=size=320x240:rate=15` as video source

- [ ] **Step 1: Create `video/testmain_test.go`**

```go
package video

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var testFixtureDir string

func TestMain(m *testing.M) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Println("skipping integration tests: ffmpeg not found in PATH")
		os.Exit(0)
	}

	dir, err := os.MkdirTemp("", "ffmpego_video_test_*")
	if err != nil {
		log.Fatal(err)
	}
	testFixtureDir = dir

	if err := generateVideoFixtures(dir); err != nil {
		os.RemoveAll(dir)
		log.Fatalf("failed to generate test fixtures: %v", err)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

const videoSrc = "testsrc2=size=320x240:rate=15"
var videoEnc = []string{"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28", "-c:a", "aac", "-ac", "1"}

func generateVideoFixtures(dir string) error {
	// no-silence: 5s video + 5s tone
	if err := runFFmpeg(dir, "no-silence.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=5",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=5",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1", "-shortest",
	); err != nil {
		return err
	}

	// all-silence: 5s video + 5s silence
	if err := runFFmpeg(dir, "all-silence.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=5",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono", "-t", "5",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1", "-shortest",
	); err != nil {
		return err
	}

	// short: 0.3s video + 0.3s tone
	if err := runFFmpeg(dir, "short.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=0.3",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=0.3",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1", "-shortest",
	); err != nil {
		return err
	}

	// no-audio: 5s video only
	if err := runFFmpeg(dir, "no-audio.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=5",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28", "-an",
	); err != nil {
		return err
	}

	// silence-start: 5s video + (2s silence concat 3s tone)
	if err := runFFmpeg(dir, "silence-start.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=5",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono", "-t", "2",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=3",
		"-filter_complex", "[1][2]concat=n=2:v=0:a=1[a]",
		"-map", "0:v", "-map", "[a]",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1", "-shortest",
	); err != nil {
		return err
	}

	// silence-middle: 6s video + (2s tone + 2s silence + 2s tone)
	if err := runFFmpeg(dir, "silence-middle.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=6",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=2",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono", "-t", "2",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=2",
		"-filter_complex", "[1][2][3]concat=n=3:v=0:a=1[a]",
		"-map", "0:v", "-map", "[a]",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1", "-shortest",
	); err != nil {
		return err
	}

	// silence-end: 5s video + (3s tone + 2s silence)
	if err := runFFmpeg(dir, "silence-end.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=5",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=3",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono", "-t", "2",
		"-filter_complex", "[1][2]concat=n=2:v=0:a=1[a]",
		"-map", "0:v", "-map", "[a]",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1", "-shortest",
	); err != nil {
		return err
	}

	return nil
}

func runFFmpeg(dir, name string, args ...string) error {
	args = append(args, "-y", filepath.Join(dir, name))
	cmd := exec.Command("ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("generating %s: %w\n%s", name, err, out)
	}
	return nil
}

func fixture(name string) string {
	return filepath.Join(testFixtureDir, name)
}

func assertValidMedia(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("output file not accessible: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output file is empty")
	}
	cmd := exec.Command("ffprobe", "-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0", path)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ffprobe failed on output: %v", err)
	}
	dur, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || dur <= 0 {
		t.Fatalf("output has invalid duration: %q", strings.TrimSpace(string(out)))
	}
}

func assertDuration(t *testing.T, path string, expected, tolerance float64) {
	t.Helper()
	cmd := exec.Command("ffprobe", "-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0", path)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ffprobe failed: %v", err)
	}
	actual, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		t.Fatalf("failed to parse duration: %v", err)
	}
	if math.Abs(actual-expected) > tolerance {
		t.Errorf("duration %.2fs outside tolerance: expected %.2f ± %.2f", actual, expected, tolerance)
	}
}

func countTempDirs(t *testing.T, prefix string) int {
	t.Helper()
	entries, err := os.ReadDir(os.TempDir())
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			count++
		}
	}
	return count
}
```

- [ ] **Step 2: Verify fixtures generate**

Run: `go test ./video/ -run TestMain -v -count=1`
Expected: exits 0

- [ ] **Step 3: Commit**

```bash
git add video/testmain_test.go
git commit -m "test(video): add TestMain with fixture generation and test helpers"
```

---

### Task 8: Video Info Tests

**Files:**
- Create: `video/info_test.go`

Mirrors audio `info_test.go` but checks video-specific fields: Width=320, Height=240, FrameRate~15, VideoCodec="h264".

- [ ] **Step 1: Create `video/info_test.go`**

```go
package video

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_ValidFile(t *testing.T) {
	t.Parallel()
	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if v.Path() != fixture("no-silence.mp4") {
		t.Errorf("Path() = %q, want %q", v.Path(), fixture("no-silence.mp4"))
	}
}

func TestNew_FileNotFound(t *testing.T) {
	t.Parallel()
	_, err := New("/nonexistent/file.mp4")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestNew_PermissionDenied(t *testing.T) {
	t.Parallel()
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "noperm.mp4")
	os.WriteFile(path, []byte("fake"), 0000)
	_, err := New(path)
	if err == nil {
		t.Fatal("expected error for unreadable file, got nil")
	}
}

func TestGetInfo_ReturnsCorrectMetadata(t *testing.T) {
	t.Parallel()
	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	info, err := v.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo() error: %v", err)
	}
	if info.Width != 320 {
		t.Errorf("Width = %d, want 320", info.Width)
	}
	if info.Height != 240 {
		t.Errorf("Height = %d, want 240", info.Height)
	}
	if info.FrameRate < 14 || info.FrameRate > 16 {
		t.Errorf("FrameRate = %.2f, want ~15", info.FrameRate)
	}
	if info.Duration < 4.5 || info.Duration > 5.5 {
		t.Errorf("Duration = %.2f, want ~5.0", info.Duration)
	}
	if info.VideoCodec != "h264" {
		t.Errorf("VideoCodec = %q, want h264", info.VideoCodec)
	}
	if info.FileSizeBytes <= 0 {
		t.Errorf("FileSizeBytes = %d, want > 0", info.FileSizeBytes)
	}
}

func TestGetInfo_Cached(t *testing.T) {
	t.Parallel()
	v, _ := New(fixture("no-silence.mp4"))
	info1, _ := v.GetInfo()
	info2, _ := v.GetInfo()
	if info1.Duration != info2.Duration || info1.Width != info2.Width {
		t.Error("cached GetInfo() returned different values")
	}
}

func TestGetInfo_CopyNotShared(t *testing.T) {
	t.Parallel()
	v, _ := New(fixture("no-silence.mp4"))
	info1, _ := v.GetInfo()
	info1.Duration = 999.0
	info1.Width = 1

	info2, _ := v.GetInfo()
	if info2.Duration == 999.0 {
		t.Error("modifying returned Info corrupted the cache (Duration)")
	}
	if info2.Width == 1 {
		t.Error("modifying returned Info corrupted the cache (Width)")
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./video/ -run "TestNew|TestGetInfo" -v -count=1`
Expected: 6 PASS

- [ ] **Step 3: Commit**

```bash
git add video/info_test.go
git commit -m "test(video): add constructor and GetInfo integration tests"
```

---

### Task 9: Video Silence Detection Tests

**Files:**
- Create: `video/silence_test.go`

Mirrors audio silence tests plus `TestGetNonSilentSegments_NoAudioStream` for video-only fixture.

- [ ] **Step 1: Create `video/silence_test.go`**

```go
package video

import (
	"testing"
)

func TestGetNonSilentSegments_SilenceAtStart(t *testing.T) {
	t.Parallel()
	v, _ := New(fixture("silence-start.mp4"))
	segments, err := v.GetNonSilentSegments(SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(segments) == 0 {
		t.Fatal("expected at least 1 segment")
	}
	if segments[0].StartTime < 1.0 {
		t.Errorf("first segment starts at %.2f, expected after ~2s silence", segments[0].StartTime)
	}
}

func TestGetNonSilentSegments_SilenceMiddle(t *testing.T) {
	t.Parallel()
	v, _ := New(fixture("silence-middle.mp4"))
	segments, err := v.GetNonSilentSegments(SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(segments) < 2 {
		t.Fatalf("expected at least 2 segments, got %d", len(segments))
	}
}

func TestGetNonSilentSegments_SilenceAtEnd(t *testing.T) {
	t.Parallel()
	v, _ := New(fixture("silence-end.mp4"))
	segments, err := v.GetNonSilentSegments(SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(segments) == 0 {
		t.Fatal("expected at least 1 segment")
	}
	info, _ := v.GetInfo()
	last := segments[len(segments)-1]
	if last.EndTime > info.Duration-1.0 {
		t.Errorf("last segment ends at %.2f, expected before trailing silence", last.EndTime)
	}
}

func TestGetNonSilentSegments_NoSilence(t *testing.T) {
	t.Parallel()
	v, _ := New(fixture("no-silence.mp4"))
	segments, err := v.GetNonSilentSegments(SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
}

func TestGetNonSilentSegments_AllSilence(t *testing.T) {
	t.Parallel()
	v, _ := New(fixture("all-silence.mp4"))
	segments, err := v.GetNonSilentSegments(SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdRelaxed,
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(segments) != 0 {
		t.Errorf("expected 0 segments, got %d", len(segments))
	}
}

func TestGetNonSilentSegments_ZeroValueConfig(t *testing.T) {
	t.Parallel()
	v, _ := New(fixture("silence-middle.mp4"))
	segments, err := v.GetNonSilentSegments(SilenceConfig{})
	if err != nil {
		t.Fatalf("error with zero config: %v", err)
	}
	if len(segments) < 2 {
		t.Fatalf("expected at least 2 segments, got %d", len(segments))
	}
}

func TestGetNonSilentSegments_NoAudioStream(t *testing.T) {
	t.Parallel()
	v, _ := New(fixture("no-audio.mp4"))
	segments, err := v.GetNonSilentSegments(SilenceConfig{})
	// Video without audio: either returns error or treats whole file as non-silent
	if err != nil {
		t.Logf("silence detection on no-audio video returned error (acceptable): %v", err)
	} else {
		t.Logf("silence detection on no-audio video returned %d segments", len(segments))
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./video/ -run "TestGetNonSilentSegments" -v -count=1`
Expected: 7 PASS

- [ ] **Step 3: Commit**

```bash
git add video/silence_test.go
git commit -m "test(video): add silence detection integration tests including no-audio edge case"
```

---

## Chunk 4: Video Remaining (Segment + Convert + RemoveSilence + Parallel)

### Task 10: Video Segment Tests

**Files:**
- Create: `video/segment_test.go`

Mirrors audio segment tests but with `.mp4` outputs.

- [ ] **Step 1: Create `video/segment_test.go`**

Same structure as `audio/segment_test.go` (Task 4) but using `video.New`, `.mp4` file extensions, and `video.ConvertConfig{VideoCodec: CodecH264, AudioCodec: CodecAAC, Quality: 28, Preset: PresetUltrafast}` for the WithConfig test.

- [ ] **Step 2: Run tests**

Run: `go test ./video/ -run "TestExtractSegment|TestConcatenateSegments" -v -count=1`
Expected: 7 PASS

- [ ] **Step 3: Commit**

```bash
git add video/segment_test.go
git commit -m "test(video): add segment extraction and concatenation tests"
```

---

### Task 11: Video Convert Tests

**Files:**
- Create: `video/convert_test.go`

- [ ] **Step 1: Create `video/convert_test.go`**

Tests: `TestConvert_BasicCodec` (convert to H265), `TestConvert_AspectRatio` (convert to 16:9, verify even dimensions), `TestConvert_OutputDirCreated`.

For `TestConvert_AspectRatio`: convert the fixture to `AspectRatio16x9`, then open the output with `video.New` + `GetInfo()` and verify that `info.Width / info.Height` approximates 16/9 and both are even numbers.

- [ ] **Step 2: Run tests**

Run: `go test ./video/ -run "TestConvert" -v -count=1`
Expected: 3 PASS

- [ ] **Step 3: Commit**

```bash
git add video/convert_test.go
git commit -m "test(video): add conversion integration tests including aspect ratio"
```

---

### Task 12: Video RemoveSilence and Parallel Tests

**Files:**
- Create: `video/remove_silence_test.go`

- [ ] **Step 1: Create `video/remove_silence_test.go`**

Same structure as `audio/remove_silence_test.go` (Task 6) but using `video.New`, `.mp4` extensions. Includes all 5 RemoveSilence tests + all 5 parallel tests.

- [ ] **Step 2: Run all video tests**

Run: `go test ./video/ -v -count=1`
Expected: 33 PASS

- [ ] **Step 3: Commit**

```bash
git add video/remove_silence_test.go
git commit -m "test(video): add RemoveSilence and parallel execution tests"
```

---

## Final Verification

### Task 13: Run Full Test Suite

- [ ] **Step 1: Run all tests across the entire project**

Run: `go test ./... -v -count=1`
Expected: 65 tests PASS (32 audio + 33 video + 13 ffutil unit tests = 78 total)

- [ ] **Step 2: Run with race detector**

Run: `go test ./... -race -count=1`
Expected: PASS with no race conditions detected

- [ ] **Step 3: Final commit if any fixes were needed**

```bash
git add -A
git commit -m "test: fix any issues found during full test suite run"
```
