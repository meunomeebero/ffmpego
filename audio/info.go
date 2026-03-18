package audio

import (
	"fmt"
	"os"
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

// GetInfo retrieves information about the audio file.
// Results are cached — subsequent calls return the cached value without invoking ffprobe.
func (a *Audio) GetInfo() (*Info, error) {
	if a.info != nil {
		copy := *a.info
		return &copy, nil
	}

	output, err := ffprobeAudio(a.path)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio info: %w", err)
	}

	info := &Info{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "sample_rate=") {
			info.SampleRate, _ = strconv.Atoi(strings.TrimPrefix(line, "sample_rate="))
		} else if strings.HasPrefix(line, "channels=") {
			info.Channels, _ = strconv.Atoi(strings.TrimPrefix(line, "channels="))
		} else if strings.HasPrefix(line, "codec_name=") {
			info.Codec = strings.TrimPrefix(line, "codec_name=")
		} else if strings.HasPrefix(line, "bit_rate=") {
			info.BitRate, _ = strconv.Atoi(strings.TrimPrefix(line, "bit_rate="))
		} else if strings.HasPrefix(line, "duration=") {
			info.Duration, _ = strconv.ParseFloat(strings.TrimPrefix(line, "duration="), 64)
		}
	}

	fileInfo, err := os.Stat(a.path)
	if err == nil {
		info.FileSizeBytes = fileInfo.Size()
	}

	a.info = info
	copy := *info
	return &copy, nil
}
