package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/meunomeebero/ffmpego"
)

func main() {
	// Check if enough arguments were provided
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <input_file> <output_directory>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputDir := os.Args[2]

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Generate output filenames based on current timestamp
	timestamp := time.Now().Format("20060102_150405")
	baseName := filepath.Base(inputFile)
	baseNameWithoutExt := baseName[:len(baseName)-len(filepath.Ext(baseName))]

	videoOutput := filepath.Join(outputDir, baseNameWithoutExt+"_nosilence_"+timestamp+".mp4")
	audioOutput := filepath.Join(outputDir, baseNameWithoutExt+"_nosilence_"+timestamp+".mp3")

	// Initialize the library with a custom logger for debugging
	customLogger := ffmpego.NewDefaultLogger(os.Stdout)
	ffmpeg := ffmpego.NewWithLogger(customLogger)

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

	// Silence detection configuration
	silenceConfig := ffmpego.SilenceConfig{
		MinSilenceLen: ffmpego.SilenceDurationShort,    // 300ms of minimum silence
		SilenceThresh: ffmpego.SilenceThresholdDefault, // -30dB considered silence
	}

	// Audio configuration
	audioConfig := &ffmpego.AudioConfig{
		Codec:      ffmpego.AudioCodecMP3,    // Use MP3 codec
		Quality:    ffmpego.AudioQualityHigh, // High quality (lower values are better)
		SampleRate: 44100,                    // 44.1kHz
		Channels:   2,                        // Stereo
	}

	// Video configuration
	videoConfig := &ffmpego.VideoConfig{
		VideoCodec: ffmpego.VideoCodecH264,   // Use H.264 codec
		CRF:        ffmpego.VideoQualityHigh, // High quality
		Preset:     ffmpego.PresetMedium,     // Balance between speed and quality
		// Keep original resolution
	}

	// Process audio by removing silence
	fmt.Printf("\nRemoving silence from audio to %s...\n", audioOutput)
	err = ffmpeg.Audio.RemoveSilence(inputFile, audioOutput, silenceConfig, audioConfig)
	if err != nil {
		log.Fatalf("Error removing silence from audio: %v", err)
	}
	fmt.Printf("Audio processed successfully: %s\n", audioOutput)

	// Process video by removing silence
	fmt.Printf("\nRemoving silence from video to %s...\n", videoOutput)
	fmt.Println("This operation may take some time, depending on the video size...")
	err = ffmpeg.Video.RemoveSilence(inputFile, videoOutput, silenceConfig, videoConfig)
	if err != nil {
		log.Fatalf("Error removing silence from video: %v", err)
	}
	fmt.Printf("Video processed successfully: %s\n", videoOutput)

	// Display processed video information
	fmt.Println("\nGetting processed video information...")
	processedInfo, err := ffmpeg.Video.GetInfo(videoOutput)
	if err != nil {
		log.Printf("Error getting processed video information: %v", err)
	} else {
		fmt.Printf("New duration: %.2f seconds\n", processedInfo.Duration)
		fmt.Printf("Reduction: %.2f%%\n", 100.0*(1.0-processedInfo.Duration/videoInfo.Duration))
	}

	fmt.Println("\nProcessing completed successfully!")
}
