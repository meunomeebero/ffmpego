package main

import (
	"fmt"
	"log"
	"os"

	"github.com/meunomeebero/ffmpego"
)

func main() {
	// Initialize the library
	ffmpeg := ffmpego.New()

	// Check if enough arguments were provided
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <input_file> <output_file>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Get video information
	fmt.Println("Getting video information...")
	videoInfo, err := ffmpeg.Video.GetInfo(inputFile)
	if err != nil {
		log.Fatalf("Error getting video information: %v", err)
	}

	// Display video information
	fmt.Printf("Resolution: %dx%d\n", videoInfo.Width, videoInfo.Height)
	fmt.Printf("FPS: %.2f\n", videoInfo.FrameRate)
	fmt.Printf("Duration: %.2f seconds\n", videoInfo.Duration)
	fmt.Printf("Video codec: %s\n", videoInfo.VideoCodec)
	fmt.Printf("Audio codec: %s\n", videoInfo.AudioCodec)

	// Resize video to 720p
	fmt.Println("\nResizing video to 720p...")
	videoConfig := &ffmpego.VideoConfig{
		TargetResolution: ffmpego.ResolutionHD,       // Using constant for 720p
		VideoCodec:       ffmpego.VideoCodecH264,     // Using constant for H.264
		CRF:              ffmpego.VideoQualityHigh,   // Using constant for high quality
		Preset:           ffmpego.PresetMedium,       // Using constant for medium preset
		PixelFormat:      ffmpego.PixelFormatYuv420p, // Using constant for pixel format
	}

	err = ffmpeg.Video.Resize(inputFile, outputFile, videoConfig)
	if err != nil {
		log.Fatalf("Error resizing video: %v", err)
	}

	fmt.Printf("\nVideo successfully resized: %s\n", outputFile)

	// Extract audio from video
	audioFile := outputFile + ".mp3"
	fmt.Printf("\nExtracting audio to %s...\n", audioFile)

	err = ffmpeg.Audio.ExtractFromVideo(inputFile, audioFile)
	if err != nil {
		log.Fatalf("Error extracting audio: %v", err)
	}

	fmt.Printf("Audio successfully extracted: %s\n", audioFile)

	fmt.Println("\nProcessing completed successfully!")
}
