# FFMPEGo Silence Removal Example

This advanced example demonstrates how to remove silent periods from audio and video files using the FFMPEGo library.

## How to run

```bash
cd example/silence
go run main.go <input_file> <output_directory>
```

Example:

```bash
go run main.go /path/to/video.mp4 /path/to/output/
```

## What this example does

1. Loads the provided video file
2. Displays basic information such as resolution, FPS, and duration
3. Processes the audio extracted from the video, removing silent periods
4. Processes the complete video, removing silent periods
5. Compares the original and final durations, showing the percentage reduction

This example is useful for:
- Optimizing the duration of video lessons by removing silent periods
- Creating shorter versions of meeting recordings
- Pre-processing videos for more efficient editing

### Silence Parameters

- `MinSilenceLen`: Minimum duration (in ms) that a segment needs to have to be considered silence (300ms in the example)
- `SilenceThresh`: Volume threshold (in dB) below which audio is considered silence (-30dB in the example)

These parameters can be adjusted in the code as needed. 