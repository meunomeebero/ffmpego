# FFMPEGo Basic Example

This example demonstrates the basic functionality of the FFMPEGo library:

1. Getting information from a video file
2. Resizing a video to a specific resolution
3. Extracting audio from a video

## How to run

```bash
cd example/basic
go run main.go <input_file> <output_file>
```

Example:

```bash
go run main.go /path/to/video.mp4 /path/to/output.mp4
```

## What this example does

1. Loads the provided video file
2. Displays information such as resolution, FPS, duration, and codecs
3. Resizes the video to 720p (1280x720) using the h264 codec
4. Extracts the audio from the original video to a separate MP3 file

This example is useful for understanding the basic operations of the library. 