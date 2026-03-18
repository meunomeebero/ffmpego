package audio

import "os/exec"

func ffprobeAudio(path string) ([]byte, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=sample_rate,channels,codec_name,bit_rate",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1",
		path)
	return cmd.Output()
}
