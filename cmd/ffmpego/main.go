package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/meunomeebero/ffmpego/audio"
	"github.com/meunomeebero/ffmpego/video"
)

const usage = `ffmpego - simple media processing from the terminal

Usage:
  ffmpego -rs <input> <output>    Remove silence from video or audio

Examples:
  ffmpego -rs recording.mp4 clean.mp4
  ffmpego -rs podcast.mp3 clean.mp3

Install:
  go install github.com/meunomeebero/ffmpego/cmd/ffmpego@latest`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "-rs":
		removeSilence()
	case "-h", "--help", "help":
		fmt.Println(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s\n", os.Args[1], usage)
		os.Exit(1)
	}
}

func removeSilence() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: ffmpego -rs <input> <output>")
		os.Exit(1)
	}

	input := os.Args[2]
	output := os.Args[3]
	start := time.Now()

	// Detect file type by trying video first, fall back to audio
	if isVideo(input) {
		v, err := video.New(input)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		info, err := v.GetInfo()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Printf("Input: %s (%dx%d, %.0ffps, %.1fs)\n", input, info.Width, info.Height, info.FrameRate, info.Duration)
		fmt.Println("Removing silence...")

		err = v.RemoveSilence(output, video.SilenceConfig{})
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		outV, _ := video.New(output)
		outInfo, _ := outV.GetInfo()
		fmt.Printf("Output: %s (%.1fs, saved %.1fs)\n", output, outInfo.Duration, info.Duration-outInfo.Duration)
	} else {
		a, err := audio.New(input)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		info, err := a.GetInfo()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Printf("Input: %s (%.1fs, %dHz)\n", input, info.Duration, info.SampleRate)
		fmt.Println("Removing silence...")

		err = a.RemoveSilence(output, audio.SilenceConfig{})
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		outA, _ := audio.New(output)
		outInfo, _ := outA.GetInfo()
		fmt.Printf("Output: %s (%.1fs, saved %.1fs)\n", output, outInfo.Duration, info.Duration-outInfo.Duration)
	}

	fmt.Printf("Done in %.1fs\n", time.Since(start).Seconds())
}

func isVideo(path string) bool {
	v, err := video.New(path)
	if err != nil {
		return false
	}
	info, err := v.GetInfo()
	if err != nil {
		return false
	}
	return info.Width > 0 && info.Height > 0
}
