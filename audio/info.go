package audio

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Info contains information about an audio file
type Info struct {
	Duration      float64
	SampleRate    int
	Channels      int
	Codec         string
	BitRate       int
	FileSizeBytes int64
}

// GetInfo retrieves information about the audio file
func (a *Audio) GetInfo() (*Info, error) {
	// Check if FFprobe is available
	_, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, fmt.Errorf("ffprobe not found in PATH: %w", err)
	}

	// Get audio stream information
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=sample_rate,channels,codec_name,bit_rate",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1",
		a.path)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get audio info: %w", err)
	}

	// Parse output
	info := &Info{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "sample_rate=") {
			sampleRateStr := strings.TrimPrefix(line, "sample_rate=")
			info.SampleRate, _ = strconv.Atoi(sampleRateStr)
		} else if strings.HasPrefix(line, "channels=") {
			channelsStr := strings.TrimPrefix(line, "channels=")
			info.Channels, _ = strconv.Atoi(channelsStr)
		} else if strings.HasPrefix(line, "codec_name=") {
			info.Codec = strings.TrimPrefix(line, "codec_name=")
		} else if strings.HasPrefix(line, "bit_rate=") {
			bitRateStr := strings.TrimPrefix(line, "bit_rate=")
			info.BitRate, _ = strconv.Atoi(bitRateStr)
		} else if strings.HasPrefix(line, "duration=") {
			durStr := strings.TrimPrefix(line, "duration=")
			info.Duration, _ = strconv.ParseFloat(durStr, 64)
		}
	}

	// Add file size information
	fileInfo, err := os.Stat(a.path)
	if err == nil {
		info.FileSizeBytes = fileInfo.Size()
	}

	return info, nil
}
