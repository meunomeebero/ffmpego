package video

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Info contains information about a video file
type Info struct {
	Width         int
	Height        int
	Duration      float64
	FrameRate     float64
	VideoCodec    string
	AudioCodec    string
	PixelFormat   string
	FileSizeBytes int64
}

// GetInfo retrieves information about the video file.
// Results are cached — subsequent calls return the cached value without invoking ffprobe.
func (v *Video) GetInfo() (*Info, error) {
	if v.info != nil {
		copy := *v.info
		return &copy, nil
	}

	output, err := ffprobeVideo(v.path)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	info := &Info{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "width=") {
			info.Width, _ = strconv.Atoi(strings.TrimPrefix(line, "width="))
		} else if strings.HasPrefix(line, "height=") {
			info.Height, _ = strconv.Atoi(strings.TrimPrefix(line, "height="))
		} else if strings.HasPrefix(line, "r_frame_rate=") {
			frParts := strings.Split(strings.TrimPrefix(line, "r_frame_rate="), "/")
			if len(frParts) == 2 {
				num, _ := strconv.ParseFloat(frParts[0], 64)
				den, _ := strconv.ParseFloat(frParts[1], 64)
				if den > 0 {
					info.FrameRate = num / den
				}
			}
		} else if strings.HasPrefix(line, "codec_name=") {
			info.VideoCodec = strings.TrimPrefix(line, "codec_name=")
		} else if strings.HasPrefix(line, "duration=") {
			info.Duration, _ = strconv.ParseFloat(strings.TrimPrefix(line, "duration="), 64)
		} else if strings.HasPrefix(line, "pix_fmt=") {
			info.PixelFormat = strings.TrimPrefix(line, "pix_fmt=")
		}
	}

	audioOutput, err := ffprobeAudioStream(v.path)
	if err == nil {
		for _, line := range strings.Split(string(audioOutput), "\n") {
			if strings.HasPrefix(line, "codec_name=") {
				info.AudioCodec = strings.TrimPrefix(line, "codec_name=")
				break
			}
		}
	}

	fileInfo, err := os.Stat(v.path)
	if err == nil {
		info.FileSizeBytes = fileInfo.Size()
	}

	v.info = info
	copy := *info
	return &copy, nil
}
