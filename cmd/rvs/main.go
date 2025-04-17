package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/meunomeebero/ffmpego"
)

func main() {
	input := flag.String("input", "", "Input file path")
	output := flag.String("output", "output.mp4", "Output file path")
	flag.Parse()

	if *input == "" || *output == "" {
		fmt.Println("Error: input and output are required")
		os.Exit(1)
	}

	ffmpeg := ffmpego.New()

	videoConfig := ffmpego.VideoConfig{
		Quality: ffmpego.VideoQualityLossless,
		Preset:  ffmpego.PresetMedium,
	}

	silenceConfig := ffmpego.SilenceConfig{
		MinSilenceLen: ffmpego.SilenceDurationShort,
		SilenceThresh: ffmpego.SilenceThresholdRelaxed,
	}

	if err := ffmpeg.Video.RemoveSilence(*input, *output, silenceConfig, &videoConfig); err != nil {
		fmt.Println("Error: failed to remove silence", err)
		os.Exit(1)
	}
}
