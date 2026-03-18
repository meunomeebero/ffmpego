package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/meunomeebero/ffmpego/video"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.mp4> <output.mp4>\n", os.Args[0])
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]
	startTime := time.Now()

	fmt.Println("=== Remove Video Silence ===")
	fmt.Printf("Input: %s\n", inputPath)
	fmt.Printf("Output: %s\n\n", outputPath)

	v, err := video.New(inputPath)
	if err != nil {
		log.Fatalf("Error opening video: %v", err)
	}

	info, err := v.GetInfo()
	if err != nil {
		log.Fatalf("Error getting video info: %v", err)
	}
	fmt.Printf("Video: %dx%d, %.2f fps, Duration: %.2fs\n\n", info.Width, info.Height, info.FrameRate, info.Duration)

	fmt.Println("Removing silence...")
	err = v.RemoveSilence(outputPath, video.SilenceConfig{
		MinSilenceDuration: video.SilenceDurationMedium,
		SilenceThreshold:   video.SilenceThresholdModerate,
	})
	if err != nil {
		log.Fatalf("Error removing silence: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("\nDone! Saved to: %s\n", outputPath)
	fmt.Printf("Processing time: %.2fs\n", duration.Seconds())

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
