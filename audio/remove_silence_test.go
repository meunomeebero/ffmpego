package audio

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func tempDirNames(t *testing.T, prefix string) map[string]struct{} {
	t.Helper()

	entries, err := os.ReadDir(os.TempDir())
	if err != nil {
		t.Fatalf("tempDirNames: failed to read temp dir: %v", err)
	}

	names := make(map[string]struct{})
	for _, e := range entries {
		if e.IsDir() && len(e.Name()) >= len(prefix) && e.Name()[:len(prefix)] == prefix {
			names[e.Name()] = struct{}{}
		}
	}
	return names
}

// --- RemoveSilence tests ---

func TestRemoveSilence_ProducesShorterOutput(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("silence-middle.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	info, err := a.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "out.wav")
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}
	if err := a.RemoveSilence(outputPath, config); err != nil {
		t.Fatalf("RemoveSilence: %v", err)
	}

	assertValidMedia(t, outputPath)

	out, err := New(outputPath)
	if err != nil {
		t.Fatalf("New (output): %v", err)
	}
	outInfo, err := out.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (output): %v", err)
	}

	if outInfo.Duration >= info.Duration {
		t.Errorf("output duration %.3fs is not shorter than input %.3fs", outInfo.Duration, info.Duration)
	}
}

func TestRemoveSilence_NoSilence_CopiesFile(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("no-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "out.wav")
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}
	if err := a.RemoveSilence(outputPath, config); err != nil {
		t.Fatalf("RemoveSilence: %v", err)
	}

	assertDuration(t, outputPath, 5.0, 1.0)
}

func TestRemoveSilence_AllSilence(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("all-silence.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "out.wav")
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdRelaxed,
	}
	err = a.RemoveSilence(outputPath, config)
	if err == nil {
		t.Fatal("RemoveSilence on all-silence file: expected error, got nil")
	}
}

func TestRemoveSilence_ShortFile(t *testing.T) {
	t.Parallel()

	a, err := New(fixture("short.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "out.wav")
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}

	// Either succeeds or returns a clear error — no panic.
	err = a.RemoveSilence(outputPath, config)
	if err != nil {
		// Acceptable: short file may produce no segments or some other error.
		t.Logf("RemoveSilence on short.wav returned error (acceptable): %v", err)
		return
	}

	assertValidMedia(t, outputPath)
}

func TestRemoveSilence_TempFilesCleanedUp(t *testing.T) {
	// NOT parallel — runs after all parallel tests complete, so no
	// other test is creating ffmpego_silence_* dirs concurrently.
	const prefix = "ffmpego_silence_"
	before := tempDirNames(t, prefix)

	a, err := New(fixture("silence-middle.wav"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "out.wav")
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}
	if err := a.RemoveSilence(outputPath, config); err != nil {
		t.Fatalf("RemoveSilence: %v", err)
	}

	after := tempDirNames(t, prefix)
	for name := range after {
		if _, existed := before[name]; !existed {
			t.Errorf("temp dir %q was not cleaned up", name)
		}
	}
}

// --- Parallel tests ---

func TestParallel_SameInput_DifferentOutputs(t *testing.T) {
	t.Parallel()

	config := SilenceConfig{}

	var wg sync.WaitGroup
	errs := make([]error, 2)

	for i := 0; i < 2; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, err := New(fixture("no-silence.wav"))
			if err != nil {
				errs[i] = err
				return
			}
			outputPath := filepath.Join(t.TempDir(), "out.wav")
			errs[i] = a.RemoveSilence(outputPath, config)
		}()
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: RemoveSilence failed: %v", i, err)
		}
	}
}

func TestParallel_DifferentInputs(t *testing.T) {
	t.Parallel()

	inputs := []string{"silence-start.wav", "silence-end.wav"}
	config := SilenceConfig{}

	var wg sync.WaitGroup
	errs := make([]error, len(inputs))

	for i, input := range inputs {
		i, input := i, input
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, err := New(fixture(input))
			if err != nil {
				errs[i] = err
				return
			}
			outputPath := filepath.Join(t.TempDir(), "out.wav")
			errs[i] = a.RemoveSilence(outputPath, config)
		}()
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d (%s): RemoveSilence failed: %v", i, inputs[i], err)
		}
	}
}

func TestParallel_SameOutput_NoPanic(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "out.wav")
	config := SilenceConfig{}

	var ready sync.WaitGroup
	var start sync.WaitGroup
	ready.Add(2)
	start.Add(1)

	var wg sync.WaitGroup

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, err := New(fixture("no-silence.wav"))
			if err != nil {
				return
			}
			ready.Done()
			start.Wait()
			// Ignore errors — we just want no panic.
			_ = a.RemoveSilence(outputPath, config)
		}()
	}

	ready.Wait()
	start.Done()
	wg.Wait()
}

func TestParallel_RemoveSilence_Concurrent(t *testing.T) {
	t.Parallel()

	config := SilenceConfig{}
	const n = 3

	var wg sync.WaitGroup
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, err := New(fixture("silence-middle.wav"))
			if err != nil {
				errs[i] = err
				return
			}
			outputPath := filepath.Join(t.TempDir(), "out.wav")
			errs[i] = a.RemoveSilence(outputPath, config)
		}()
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: RemoveSilence failed: %v", i, err)
		}
	}
}

func TestParallel_TempFilesIsolated(t *testing.T) {
	// NOT parallel — runs after all parallel tests complete, so the
	// snapshot/diff approach works without false positives from sibling tests.
	const prefix = "ffmpego_silence_"
	before := tempDirNames(t, prefix)

	config := SilenceConfig{}
	const n = 3

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, err := New(fixture("silence-middle.wav"))
			if err != nil {
				return
			}
			outputPath := filepath.Join(t.TempDir(), "out.wav")
			_ = a.RemoveSilence(outputPath, config)
		}()
	}
	wg.Wait()

	after := tempDirNames(t, prefix)
	for name := range after {
		if _, existed := before[name]; !existed {
			t.Errorf("orphaned temp dir %q after concurrent operations", name)
		}
	}
}
