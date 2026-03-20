package video

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

// RemoveSilence tests

func TestRemoveSilence_ProducesShorterOutput(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("silence-middle.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	inInfo, err := v.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (input): %v", err)
	}
	inputDuration := inInfo.Duration

	out := filepath.Join(t.TempDir(), "out.mp4")
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}
	if err := v.RemoveSilence(out, config); err != nil {
		t.Fatalf("RemoveSilence: %v", err)
	}

	assertValidMedia(t, out)

	outV, err := New(out)
	if err != nil {
		t.Fatalf("New (output): %v", err)
	}
	outInfo, err := outV.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo (output): %v", err)
	}

	if outInfo.Duration >= inputDuration {
		t.Errorf("output duration %.3f should be shorter than input %.3f", outInfo.Duration, inputDuration)
	}
}

func TestRemoveSilence_NoSilence_CopiesFile(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("no-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out.mp4")
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}
	if err := v.RemoveSilence(out, config); err != nil {
		t.Fatalf("RemoveSilence: %v", err)
	}

	assertValidMedia(t, out)
	assertDuration(t, out, 5.0, 1.0)
}

func TestRemoveSilence_AllSilence(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("all-silence.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out.mp4")
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdRelaxed,
	}
	err = v.RemoveSilence(out, config)
	if err == nil {
		t.Fatal("expected error for all-silence file, got nil")
	}
}

func TestRemoveSilence_ShortFile(t *testing.T) {
	t.Parallel()

	v, err := New(fixture("short.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out.mp4")
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}
	// Either succeeds or returns an error — no panic is the key assertion.
	err = v.RemoveSilence(out, config)
	if err != nil {
		t.Logf("RemoveSilence on short.mp4 returned error (acceptable): %v", err)
	} else {
		t.Log("RemoveSilence on short.mp4 succeeded")
	}
}

func TestRemoveSilence_TempFilesCleanedUp(t *testing.T) {
	// NOT parallel — runs after all parallel tests complete, so no
	// other test is creating ffmpego_silence_* dirs concurrently.
	const prefix = "ffmpego_silence_"
	before := countTempDirs(t, prefix)

	v, err := New(fixture("silence-middle.mp4"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out.mp4")
	config := SilenceConfig{
		MinSilenceDuration: SilenceDurationShort,
		SilenceThreshold:   SilenceThresholdModerate,
	}
	if err := v.RemoveSilence(out, config); err != nil {
		t.Fatalf("RemoveSilence: %v", err)
	}

	after := countTempDirs(t, prefix)
	for name := range after {
		if _, existed := before[name]; !existed {
			t.Errorf("temp dir %q was not cleaned up", name)
		}
	}
}

// Parallel tests

func TestParallel_SameInput_DifferentOutputs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	out1 := filepath.Join(tmpDir, "out1.mp4")
	out2 := filepath.Join(tmpDir, "out2.mp4")

	config := SilenceConfig{}

	var wg sync.WaitGroup
	errs := make([]error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		v, err := New(fixture("silence-middle.mp4"))
		if err != nil {
			errs[0] = err
			return
		}
		errs[0] = v.RemoveSilence(out1, config)
	}()
	go func() {
		defer wg.Done()
		v, err := New(fixture("silence-middle.mp4"))
		if err != nil {
			errs[1] = err
			return
		}
		errs[1] = v.RemoveSilence(out2, config)
	}()
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: %v", i+1, err)
		}
	}

	assertValidMedia(t, out1)
	assertValidMedia(t, out2)
}

func TestParallel_DifferentInputs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	out1 := filepath.Join(tmpDir, "out1.mp4")
	out2 := filepath.Join(tmpDir, "out2.mp4")

	config := SilenceConfig{}

	var wg sync.WaitGroup
	errs := make([]error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		v, err := New(fixture("silence-start.mp4"))
		if err != nil {
			errs[0] = err
			return
		}
		errs[0] = v.RemoveSilence(out1, config)
	}()
	go func() {
		defer wg.Done()
		v, err := New(fixture("silence-end.mp4"))
		if err != nil {
			errs[1] = err
			return
		}
		errs[1] = v.RemoveSilence(out2, config)
	}()
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: %v", i+1, err)
		}
	}

	assertValidMedia(t, out1)
	assertValidMedia(t, out2)
}

func TestParallel_SameOutput_NoPanic(t *testing.T) {
	t.Parallel()

	out := filepath.Join(t.TempDir(), "out.mp4")
	config := SilenceConfig{}

	var barrier sync.WaitGroup
	barrier.Add(2)

	var wg sync.WaitGroup
	wg.Add(2)

	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			v, err := New(fixture("silence-middle.mp4"))
			if err != nil {
				barrier.Done()
				return
			}
			barrier.Done()
			barrier.Wait()
			// Ignore error — writing same output concurrently may fail.
			_ = v.RemoveSilence(out, config)
		}()
	}

	wg.Wait()
	// No panic is the key assertion; file validity is not checked.
}

func TestParallel_RemoveSilence_Concurrent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	config := SilenceConfig{}

	var wg sync.WaitGroup
	errs := make([]error, 3)

	inputs := []string{
		fixture("silence-start.mp4"),
		fixture("silence-middle.mp4"),
		fixture("silence-end.mp4"),
	}

	for i, input := range inputs {
		i, input := i, input
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := New(input)
			if err != nil {
				errs[i] = err
				return
			}
			out := filepath.Join(tmpDir, "out"+string(rune('0'+i))+".mp4")
			errs[i] = v.RemoveSilence(out, config)
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: %v", i+1, err)
		}
	}
}

func TestParallel_TempFilesIsolated(t *testing.T) {
	// NOT parallel — runs after all parallel tests complete, so the
	// snapshot/diff approach works without false positives from sibling tests.
	const prefix = "ffmpego_silence_"
	before := countTempDirs(t, prefix)

	config := SilenceConfig{}
	var wg sync.WaitGroup
	errs := make([]error, 3)

	inputs := []string{
		fixture("silence-start.mp4"),
		fixture("silence-middle.mp4"),
		fixture("silence-end.mp4"),
	}

	for i, input := range inputs {
		i, input := i, input
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := New(input)
			if err != nil {
				errs[i] = err
				return
			}
			out := filepath.Join(t.TempDir(), fmt.Sprintf("out%d.mp4", i))
			errs[i] = v.RemoveSilence(out, config)
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: %v", i+1, err)
		}
	}

	after := countTempDirs(t, prefix)
	for name := range after {
		if _, existed := before[name]; !existed {
			t.Errorf("orphaned temp dir %q after concurrent operations", name)
		}
	}
}
