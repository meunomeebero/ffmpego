package video

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var testFixtureDir string

func TestMain(m *testing.M) {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Println("skipping: ffmpeg not found in PATH")
		os.Exit(0)
	}

	dir, err := os.MkdirTemp("", "ffmpego_video_test_*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	testFixtureDir = dir

	if err := generateVideoFixtures(dir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate fixtures: %v\n", err)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func runFFmpeg(dir, name string, args ...string) error {
	args = append(args, "-y", filepath.Join(dir, name))
	cmd := exec.Command("ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w - %s", name, err, out)
	}
	return nil
}

func generateVideoFixtures(dir string) error {
	const videoSrc = "testsrc2=size=320x240:rate=15"
	const encArgs = "-c:v libx264 -preset ultrafast -crf 28 -c:a aac -ac 1"

	// no-silence.mp4: 5s video + 5s sine tone, simple mux
	if err := runFFmpeg(dir, "no-silence.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=5",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=5",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1",
		"-shortest",
	); err != nil {
		return err
	}

	// all-silence.mp4: 5s video + anullsrc (silence), simple mux
	if err := runFFmpeg(dir, "all-silence.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=5",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1",
		"-t", "5",
		"-shortest",
	); err != nil {
		return err
	}

	// short.mp4: 0.3s video + sine
	if err := runFFmpeg(dir, "short.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=0.3",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=0.3",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1",
		"-shortest",
	); err != nil {
		return err
	}

	// no-audio.mp4: 5s video, no audio stream
	if err := runFFmpeg(dir, "no-audio.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=5",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-an",
	); err != nil {
		return err
	}

	// silence-start.mp4: 5s video + (2s silence concat 3s sine)
	_ = encArgs
	if err := runFFmpeg(dir, "silence-start.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=5",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=3",
		"-filter_complex", "[1]atrim=duration=2[s];[s][2]concat=n=2:v=0:a=1[a]",
		"-map", "0:v",
		"-map", "[a]",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1",
		"-shortest",
	); err != nil {
		return err
	}

	// silence-middle.mp4: 6s video + (2s sine + 2s silence + 2s sine)
	if err := runFFmpeg(dir, "silence-middle.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=6",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=2",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=2",
		"-filter_complex", "[2]atrim=duration=2[s];[1][s][3]concat=n=3:v=0:a=1[a]",
		"-map", "0:v",
		"-map", "[a]",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1",
		"-shortest",
	); err != nil {
		return err
	}

	// silence-end.mp4: 5s video + (3s sine concat 2s silence)
	if err := runFFmpeg(dir, "silence-end.mp4",
		"-f", "lavfi", "-i", videoSrc+":duration=5",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=3",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono",
		"-filter_complex", "[2]atrim=duration=2[s];[1][s]concat=n=2:v=0:a=1[a]",
		"-map", "0:v",
		"-map", "[a]",
		"-c:v", "libx264", "-preset", "ultrafast", "-crf", "28",
		"-c:a", "aac", "-ac", "1",
		"-shortest",
	); err != nil {
		return err
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
		t.Fatalf("assertValidMedia: file not found: %s: %v", path, err)
	}
	if info.Size() == 0 {
		t.Fatalf("assertValidMedia: file is empty: %s", path)
	}

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1",
		path)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("assertValidMedia: ffprobe failed on %s: %v", path, err)
	}

	output := strings.TrimSpace(string(out))
	if !strings.HasPrefix(output, "duration=") {
		t.Fatalf("assertValidMedia: unexpected ffprobe output for %s: %s", path, output)
	}

	durStr := strings.TrimPrefix(output, "duration=")
	dur, err := strconv.ParseFloat(durStr, 64)
	if err != nil || dur <= 0 {
		t.Fatalf("assertValidMedia: invalid duration %q for %s", durStr, path)
	}
}

func assertDuration(t *testing.T, path string, expected, tolerance float64) {
	t.Helper()

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1",
		path)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("assertDuration: ffprobe failed on %s: %v", path, err)
	}

	output := strings.TrimSpace(string(out))
	durStr := strings.TrimPrefix(output, "duration=")
	dur, err := strconv.ParseFloat(durStr, 64)
	if err != nil {
		t.Fatalf("assertDuration: failed to parse duration %q for %s: %v", durStr, path, err)
	}

	diff := dur - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Fatalf("assertDuration: %s duration %.3f differs from expected %.3f by %.3f (tolerance %.3f)",
			path, dur, expected, diff, tolerance)
	}
}

func countTempDirs(t *testing.T, prefix string) map[string]struct{} {
	t.Helper()

	entries, err := os.ReadDir(os.TempDir())
	if err != nil {
		t.Fatalf("countTempDirs: failed to read temp dir: %v", err)
	}

	names := make(map[string]struct{})
	for _, e := range entries {
		if e.IsDir() && len(e.Name()) >= len(prefix) && e.Name()[:len(prefix)] == prefix {
			names[e.Name()] = struct{}{}
		}
	}
	return names
}
