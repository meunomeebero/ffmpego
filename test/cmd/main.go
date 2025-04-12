package main

import (
	"flag"

	"github.com/meunomeebero/ffmpego"
)

func main() {
	input := flag.String("input", "", "input file")
	output := flag.String("output", "", "output file")
	flag.Parse()

	silenceConfig := ffmpego.SilenceConfig{
		MinSilenceLen: ffmpego.SilenceDurationShort,
		SilenceThresh: ffmpego.SilenceThresholdRelaxed,
	}

	videoConfig := ffmpego.VideoConfig{
		Preset: ffmpego.PresetMedium,
		CRF:    ffmpego.VideoQualityHigher,
	}

	ffmpeg := ffmpego.New()
	ffmpeg.Video.RemoveSilence(*input, *output, silenceConfig, &videoConfig)
}
