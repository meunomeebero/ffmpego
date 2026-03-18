# Integration Tests Design — ffmpego

## Goal

Add integration tests covering all public API functions, edge cases, and parallel execution safety. Tests verify that the library works correctly with real ffmpeg processes, not mocks.

## Fixture Generation

All test media files are generated via ffmpeg in `TestMain`. No binary files committed to git.

### Fixtures

| Name | Description | Duration |
|---|---|---|
| `silence-start` | 2s silence + 3s 440Hz tone | 5s |
| `silence-middle` | 2s tone + 2s silence + 2s tone | 6s |
| `silence-end` | 3s 440Hz tone + 2s silence | 5s |
| `no-silence` | 5s continuous 440Hz tone | 5s |
| `all-silence` | 5s total silence | 5s |
| `short` | 0.3s tone (below min segment threshold) | 0.3s |
| `no-audio` (video only) | 5s video with no audio stream | 5s |

Each fixture is ~50-100KB. Generated into a temp directory created by `TestMain`, cleaned up after all tests complete.

### Audio fixture generation (exact commands)

All audio fixtures use single-command complex filtergraphs to avoid concat format incompatibilities. Output format: WAV 44100Hz mono.

```bash
# no-silence: 5s continuous tone
ffmpeg -f lavfi -i "sine=frequency=440:sample_rate=44100:duration=5" -ac 1 no-silence.wav

# all-silence: 5s silence
ffmpeg -f lavfi -i "anullsrc=r=44100:cl=mono" -t 5 all-silence.wav

# silence-start: 2s silence + 3s tone
ffmpeg -f lavfi -i "anullsrc=r=44100:cl=mono" -t 2 -f lavfi -i "sine=frequency=440:sample_rate=44100:duration=3" \
  -filter_complex "[0][1]concat=n=2:v=0:a=1" silence-start.wav

# silence-middle: 2s tone + 2s silence + 2s tone
ffmpeg -f lavfi -i "sine=frequency=440:sample_rate=44100:duration=2" \
  -f lavfi -i "anullsrc=r=44100:cl=mono" -t 2 \
  -f lavfi -i "sine=frequency=440:sample_rate=44100:duration=2" \
  -filter_complex "[0][1][2]concat=n=3:v=0:a=1" silence-middle.wav

# silence-end: 3s tone + 2s silence
ffmpeg -f lavfi -i "sine=frequency=440:sample_rate=44100:duration=3" \
  -f lavfi -i "anullsrc=r=44100:cl=mono" -t 2 \
  -filter_complex "[0][1]concat=n=2:v=0:a=1" silence-end.wav

# short: 0.3s tone
ffmpeg -f lavfi -i "sine=frequency=440:sample_rate=44100:duration=0.3" -ac 1 short.wav
```

### Video fixture generation (exact commands)

Video fixtures are generated in a single pass using `-filter_complex` with both video (`testsrc2`, 320x240, 15fps) and audio filter chains. Output format: MP4 with H.264 + AAC.

```bash
# no-silence: 5s video with continuous tone
ffmpeg -f lavfi -i "testsrc2=size=320x240:rate=15:duration=5" \
  -f lavfi -i "sine=frequency=440:sample_rate=44100:duration=5" \
  -c:v libx264 -preset ultrafast -crf 28 -c:a aac -ac 1 -shortest no-silence.mp4

# silence-middle: 2s tone + 2s silence + 2s tone (video)
ffmpeg -f lavfi -i "testsrc2=size=320x240:rate=15:duration=6" \
  -f lavfi -i "sine=frequency=440:sample_rate=44100:duration=2" \
  -f lavfi -i "anullsrc=r=44100:cl=mono" -t 2 \
  -f lavfi -i "sine=frequency=440:sample_rate=44100:duration=2" \
  -filter_complex "[1][2][3]concat=n=3:v=0:a=1[a]" -map 0:v -map "[a]" \
  -c:v libx264 -preset ultrafast -crf 28 -c:a aac -ac 1 -shortest silence-middle.mp4

# (same pattern for silence-start, silence-end, all-silence, short)

# no-audio: 5s video without audio stream
ffmpeg -f lavfi -i "testsrc2=size=320x240:rate=15:duration=5" \
  -c:v libx264 -preset ultrafast -crf 28 -an no-audio.mp4
```

### Generation strategy

Each package (`audio/`, `video/`) has a `testmain_test.go` with:

```go
var testFixtureDir string

func TestMain(m *testing.M) {
    // Skip gracefully if ffmpeg is not installed
    if _, err := exec.LookPath("ffmpeg"); err != nil {
        fmt.Println("skipping integration tests: ffmpeg not found in PATH")
        os.Exit(0)
    }

    dir, err := os.MkdirTemp("", "ffmpego_test_*")
    if err != nil {
        log.Fatal(err)
    }
    testFixtureDir = dir

    if err := generateFixtures(dir); err != nil {
        os.RemoveAll(dir)
        log.Fatalf("failed to generate test fixtures: %v", err)
    }

    code := m.Run()

    os.RemoveAll(dir)
    os.Exit(code)
}
```

A helper `generateFixtures(dir)` creates all fixtures using `exec.Command("ffmpeg", ...)`. If ffmpeg is not available, `TestMain` prints a message and exits with code 0 (skip), not code 1 (fail).

## File Structure

```
audio/
  testmain_test.go        # TestMain + fixture generation
  info_test.go            # New + GetInfo tests
  silence_test.go         # GetNonSilentSegments tests
  segment_test.go         # ExtractSegment + ConcatenateSegments tests
  convert_test.go         # Convert tests
  remove_silence_test.go  # RemoveSilence + parallel tests
video/
  testmain_test.go        # TestMain + fixture generation
  info_test.go            # New + GetInfo tests
  silence_test.go         # GetNonSilentSegments tests
  segment_test.go         # ExtractSegment + ConcatenateSegments tests
  convert_test.go         # Convert tests
  remove_silence_test.go  # RemoveSilence + parallel tests
```

## Test Inventory

### Constructors and Info (6 tests per package)

| Test | What it verifies |
|---|---|
| `TestNew_ValidFile` | Creates instance successfully with a valid file |
| `TestNew_FileNotFound` | Returns clear error for nonexistent file |
| `TestNew_PermissionDenied` | Returns clear error for permission denied. Skipped when running as root (`os.Getuid() == 0`) since root bypasses file permissions. |
| `TestGetInfo_ReturnsCorrectMetadata` | Duration, sample rate/resolution, codec match expected values |
| `TestGetInfo_Cached` | Two calls return equal values. Note: we cannot directly verify ffprobe was not spawned without mocking; this test verifies value equality and the defensive copy behavior. |
| `TestGetInfo_CopyNotShared` | Modifying returned `*Info` does not affect subsequent `GetInfo()` calls |

### Silence Detection (7 tests per package)

| Test | Fixture | What it verifies |
|---|---|---|
| `TestGetNonSilentSegments_SilenceAtStart` | silence-start | First segment starts after the silence; segment count > 0 |
| `TestGetNonSilentSegments_SilenceMiddle` | silence-middle | Returns 2 segments with a gap in the middle |
| `TestGetNonSilentSegments_SilenceAtEnd` | silence-end | Last segment ends before the trailing silence |
| `TestGetNonSilentSegments_NoSilence` | no-silence | Returns exactly 1 segment spanning the full duration |
| `TestGetNonSilentSegments_AllSilence` | all-silence | Returns 0 segments (everything is below threshold) |
| `TestGetNonSilentSegments_ZeroValueConfig` | silence-middle | `SilenceConfig{}` applies defaults and works correctly |
| `TestGetNonSilentSegments_NoAudioStream` (video only) | no-audio | Calling silence detection on a video with no audio stream — verifies behavior (error or full-file segment) |

### Segment Operations (7 tests per package)

| Test | What it verifies |
|---|---|
| `TestExtractSegment_StreamCopy` | Extracts segment with `nil` config (stream copy), output is valid media |
| `TestExtractSegment_WithConfig` | Extracts with explicit ConvertConfig, output has requested properties |
| `TestExtractSegment_OutputDirCreated` | Output to nonexistent subdirectory — directory is created automatically |
| `TestConcatenateSegments_Basic` | Joins 3 segments, output is valid and duration ≈ sum of inputs |
| `TestConcatenateSegments_SingleSegment` | Joins 1 segment (common edge case: only one non-silent region) |
| `TestConcatenateSegments_EmptyList` | Returns error for empty slice |
| `TestConcatenateSegments_MissingFile` | Returns error when a segment file does not exist |

### Convert (audio: 2 tests, video: 3 tests)

| Test | Package | What it verifies |
|---|---|---|
| `TestConvert_BasicCodec` | both | Converts to a different codec, output is valid with the new codec |
| `TestConvert_SampleRate` | audio only | Converts sample rate, output has the requested rate |
| `TestConvert_AspectRatio` | video only | Converts aspect ratio, output dimensions match expected ratio (even numbers) |
| `TestConvert_OutputDirCreated` | both | Output to nonexistent directory — created automatically |

### RemoveSilence (5 tests per package)

| Test | Fixture | What it verifies |
|---|---|---|
| `TestRemoveSilence_ProducesShorterOutput` | silence-middle | Output duration < input duration |
| `TestRemoveSilence_NoSilence_CopiesFile` | no-silence | Output duration ≈ input duration |
| `TestRemoveSilence_AllSilence` | all-silence | Returns error (no audible content) |
| `TestRemoveSilence_ShortFile` | short | Handles very short file (0.3s) — returns error or produces valid output |
| `TestRemoveSilence_TempFilesCleanedUp` | silence-middle | After completion, no `ffmpego_silence_*` dirs remain in os.TempDir() |

### Parallelism (5 tests per package)

| Test | What it verifies |
|---|---|
| `TestParallel_SameInput_DifferentOutputs` | 2 goroutines process the same input file to different outputs; both produce valid media |
| `TestParallel_DifferentInputs` | 2 goroutines process different files simultaneously; both succeed |
| `TestParallel_SameOutput_NoPanic` | 2 goroutines write to the same output path; neither panics. Note: the output file validity is OS/ffmpeg-dependent and not asserted. |
| `TestParallel_RemoveSilence_Concurrent` | 3 RemoveSilence calls run simultaneously; all complete without error |
| `TestParallel_TempFilesIsolated` | After concurrent operations, no orphaned temp files remain |

## Output Validation

Every test that produces a media file validates:

1. **File exists** and size > 0
2. **ffprobe succeeds** — runs `ffprobe -v error -show_entries format=duration` on output
3. **Duration is coherent** — for extractions and silence removal, checks duration is within tolerance (±0.5s of expected)

Shared helper functions:

```go
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
        t.Fatalf("output has invalid duration: %s", string(out))
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
    diff := actual - expected
    if diff < 0 {
        diff = -diff
    }
    if diff > tolerance {
        t.Errorf("duration %.2fs outside tolerance: expected %.2f ± %.2f", actual, expected, tolerance)
    }
}
```

## Test Parallelism

- All independent tests use `t.Parallel()` for faster execution
- Each test creates its own output directory via `t.TempDir()` to avoid path conflicts
- Explicit parallelism tests use `sync.WaitGroup` to force concurrent execution
- The `TestParallel_SameOutput_NoPanic` test uses a barrier (`sync.WaitGroup`) to ensure both goroutines start at the same time

## Constraints

- **Requires ffmpeg and ffprobe** in PATH — tests exit with code 0 (skip) if not found, not code 1 (fail)
- **No external Go dependencies** — only stdlib + the library itself
- **Tolerance for timing** — ffmpeg operations have slight variance; duration checks use ±0.5s tolerance
- **No mocks** — all tests run real ffmpeg processes
- **Root detection** — permission tests are skipped when running as root (Docker CI)
- **Test execution time** — expected ~30-60s total due to ffmpeg process spawning; `t.Parallel()` helps

## Total Test Count

| Category | Audio | Video |
|---|---|---|
| Constructors + Info | 6 | 6 |
| Silence Detection | 6 | 7 (includes no-audio) |
| Segment Operations | 7 | 7 |
| Convert | 3 | 3 |
| RemoveSilence | 5 | 5 |
| Parallelism | 5 | 5 |
| **Total** | **32** | **33** |
