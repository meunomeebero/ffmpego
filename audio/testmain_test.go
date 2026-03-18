package audio

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

	dir, err := os.MkdirTemp("", "ffmpego_audio_test_*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	testFixtureDir = dir

	if err := generateAudioFixtures(dir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate fixtures: %v\n", err)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func generateAudioFixtures(dir string) error {
	// no-silence.wav: 5s continuous 440Hz tone
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "sine=frequency=440:sample_rate=44100:duration=5",
		"-ac", "1",
		"-y", filepath.Join(dir, "no-silence.wav"))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("no-silence.wav: %w - %s", err, out)
	}

	// all-silence.wav: 5s silence
	cmd = exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "anullsrc=r=44100:cl=mono",
		"-t", "5",
		"-y", filepath.Join(dir, "all-silence.wav"))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("all-silence.wav: %w - %s", err, out)
	}

	// short.wav: 0.3s tone
	cmd = exec.Command("ffmpeg",
		"-f", "lavfi",
		"-i", "sine=frequency=440:sample_rate=44100:duration=0.3",
		"-ac", "1",
		"-y", filepath.Join(dir, "short.wav"))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("short.wav: %w - %s", err, out)
	}

	// silence-start.wav: 2s silence + 3s tone
	cmd = exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=3",
		"-filter_complex", "[0]atrim=duration=2[s];[s][1]concat=n=2:v=0:a=1",
		"-y", filepath.Join(dir, "silence-start.wav"))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("silence-start.wav: %w - %s", err, out)
	}

	// silence-middle.wav: 2s tone + 2s silence + 2s tone
	cmd = exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=2",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=2",
		"-filter_complex", "[1]atrim=duration=2[s];[0][s][2]concat=n=3:v=0:a=1",
		"-y", filepath.Join(dir, "silence-middle.wav"))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("silence-middle.wav: %w - %s", err, out)
	}

	// silence-end.wav: 3s tone + 2s silence
	cmd = exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=3",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono",
		"-filter_complex", "[1]atrim=duration=2[s];[0][s]concat=n=2:v=0:a=1",
		"-y", filepath.Join(dir, "silence-end.wav"))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("silence-end.wav: %w - %s", err, out)
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
