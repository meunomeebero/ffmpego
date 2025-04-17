package video

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/meunomeebero/ffmpego/internal/types" // Import internal types
)

// GetInfo retrieves information about a video file
func GetInfo(videoPath string) (*types.VideoInfo, error) { // Use types.VideoInfo
	// Check if FFprobe is available
	_, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, fmt.Errorf("ffprobe not found in PATH: %w", err)
	}

	// Get video stream information
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,r_frame_rate,codec_name,pix_fmt",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1",
		videoPath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// Parse output
	info := &types.VideoInfo{} // Use types.VideoInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "width=") {
			widthStr := strings.TrimPrefix(line, "width=")
			info.Width, _ = strconv.Atoi(widthStr)
		} else if strings.HasPrefix(line, "height=") {
			heightStr := strings.TrimPrefix(line, "height=")
			info.Height, _ = strconv.Atoi(heightStr)
		} else if strings.HasPrefix(line, "r_frame_rate=") {
			frStr := strings.TrimPrefix(line, "r_frame_rate=")
			frParts := strings.Split(frStr, "/")
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
			durStr := strings.TrimPrefix(line, "duration=")
			info.Duration, _ = strconv.ParseFloat(durStr, 64)
		} else if strings.HasPrefix(line, "pix_fmt=") {
			info.PixelFormat = strings.TrimPrefix(line, "pix_fmt=")
		}
	}

	// Get audio codec info
	cmd = exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1",
		videoPath)

	output, err = cmd.Output()
	if err == nil {
		lines = strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "codec_name=") {
				info.AudioCodec = strings.TrimPrefix(line, "codec_name=")
				break
			}
		}
	}

	return info, nil
}
