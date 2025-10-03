package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/meunomeebero/ffmpego/video"
)

func main() {
	startTime := time.Now()

	// Input and output paths
	inputPath := "tmp/ffmpego-ex.mp4"
	outputPath := "tmp/no-silence.mp4"
	tempDir := "tmp/segments"

	fmt.Println("=== Remove Video Silence Example ===")
	fmt.Printf("Input: %s\n", inputPath)
	fmt.Printf("Output: %s\n\n", outputPath)

	// Step 1: Open video
	fmt.Println("Step 1: Opening video...")
	v, err := video.New(inputPath)
	if err != nil {
		log.Fatalf("Error opening video: %v", err)
	}

	// Get video info
	info, err := v.GetInfo()
	if err != nil {
		log.Fatalf("Error getting video info: %v", err)
	}
	fmt.Printf("Video info: %dx%d, %.2f fps, Duration: %.2fs\n\n",
		info.Width, info.Height, info.FrameRate, info.Duration)

	// Step 2: Get non-silent segments
	fmt.Println("Step 2: Getting non-silent segments...")
	silenceConfig := video.SilenceConfig{
		MinSilenceDuration: 700,  // 700ms minimum silence
		SilenceThreshold:   -30,  // -30dB threshold
	}

	segments, err := v.GetNonSilentSegments(silenceConfig)
	if err != nil {
		log.Fatalf("Error getting non-silent segments: %v", err)
	}

	if len(segments) == 0 {
		log.Fatalf("No non-silent segments found!")
	}

	fmt.Printf("Found %d non-silent segments:\n", len(segments))
	for i, seg := range segments {
		fmt.Printf("  Segment %d: %.2fs - %.2fs (%.2fs)\n",
			i+1, seg.StartTime, seg.EndTime, seg.Duration)
	}
	fmt.Println()

	// Step 3: Create temp directory
	fmt.Println("Step 3: Creating temporary directory...")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Fatalf("Error creating temp directory: %v", err)
	}
	defer func() {
		fmt.Println("\nCleaning up temporary files...")
		os.RemoveAll(tempDir)
	}()

	// Step 4: Extract segments in parallel using goroutines
	fmt.Printf("\nStep 4: Extracting %d segments in parallel...\n", len(segments))

	// Create channels for work distribution
	type extractJob struct {
		index   int
		segment video.Segment
	}

	type extractResult struct {
		index int
		path  string
		err   error
	}

	jobs := make(chan extractJob, len(segments))
	results := make(chan extractResult, len(segments))

	// Number of workers (parallel goroutines)
	numWorkers := 4
	if len(segments) < numWorkers {
		numWorkers = len(segments)
	}

	// Start workers
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				segmentPath := filepath.Join(tempDir, fmt.Sprintf("segment_%03d.mp4", job.index))

				fmt.Printf("  [Worker %d] Extracting segment %d...\n", workerID, job.index+1)

				err := v.ExtractSegment(
					segmentPath,
					job.segment.StartTime,
					job.segment.EndTime,
					nil,
				)

				results <- extractResult{
					index: job.index,
					path:  segmentPath,
					err:   err,
				}
			}
		}(w)
	}

	// Send jobs to workers
	for i, seg := range segments {
		jobs <- extractJob{index: i, segment: seg}
	}
	close(jobs)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	segmentPaths := make([]string, len(segments))
	for result := range results {
		if result.err != nil {
			log.Fatalf("Error extracting segment %d: %v", result.index+1, result.err)
		}
		segmentPaths[result.index] = result.path
		fmt.Printf("  ✓ Segment %d extracted successfully\n", result.index+1)
	}

	fmt.Printf("\nAll %d segments extracted successfully!\n", len(segments))

	// Step 5: Concatenate segments
	fmt.Println("\nStep 5: Concatenating segments...")

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	err = video.ConcatenateSegments(segmentPaths, outputPath, nil)
	if err != nil {
		log.Fatalf("Error concatenating segments: %v", err)
	}

	// Step 6: Done!
	duration := time.Since(startTime)
	fmt.Printf("\n✓ Success! Video without silence saved to: %s\n", outputPath)
	fmt.Printf("Total processing time: %.2fs\n", duration.Seconds())

	// Get output video info
	outputVideo, err := video.New(outputPath)
	if err == nil {
		outputInfo, err := outputVideo.GetInfo()
		if err == nil {
			fmt.Printf("Output duration: %.2fs (reduced by %.2fs)\n",
				outputInfo.Duration,
				info.Duration-outputInfo.Duration)
		}
	}
}
