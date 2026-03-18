package video

import "os/exec"

func ffprobeVideo(path string) ([]byte, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,r_frame_rate,codec_name,pix_fmt",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1",
		path)
	return cmd.Output()
}

func ffprobeAudioStream(path string) ([]byte, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1",
		path)
	return cmd.Output()
}
