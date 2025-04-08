# ffmpego

ffmpego is a Go wrapper library for [FFmpeg](https://ffmpeg.org/), facilitating common audio and video manipulation operations through a simple and intuitive API.

## Requirements

- [FFmpeg](https://ffmpeg.org/) installed and available in the system PATH
- Go 1.20 or higher

## Installation

```bash
go get github.com/meunomeebero/ffmpego
```

## Basic Usage

### Initialization

```go
import "github.com/meunomeebero/ffmpego"

// Initialize with default logger
ffmpeg := ffmpego.New()

// Or with custom logger
logger := ffmpego.NewDefaultLogger(os.Stdout)
ffmpeg := ffmpego.NewWithLogger(logger)
```

### Get Video Information

```go
videoInfo, err := ffmpeg.Video.GetInfo("video.mp4")
if err != nil {
    log.Fatalf("Error getting information: %v", err)
}

fmt.Printf("Resolution: %dx%d\n", videoInfo.Width, videoInfo.Height)
fmt.Printf("Duration: %.2f seconds\n", videoInfo.Duration)
fmt.Printf("FPS: %.2f\n", videoInfo.FrameRate)
```

### Resize Video

```go
config := &ffmpego.VideoConfig{
    TargetResolution: ffmpego.ResolutionHD,     // 1280x720
    VideoCodec:       ffmpego.VideoCodecH264,   // H.264
    CRF:              ffmpego.VideoQualityHigh, // 23 - High quality
    Preset:           ffmpego.PresetMedium,     // Medium preset
}

err := ffmpeg.Video.Resize("input.mp4", "output.mp4", config)
if err != nil {
    log.Fatalf("Error resizing: %v", err)
}
```

### Remove Silence from Audio

```go
silenceConfig := ffmpego.SilenceConfig{
    MinSilenceLen: ffmpego.SilenceDurationMedium,    // 700ms
    SilenceThresh: ffmpego.SilenceThresholdDefault,  // -30 dB
}

audioConfig := &ffmpego.AudioConfig{
    Codec:      ffmpego.AudioCodecMP3,     // MP3
    Quality:    ffmpego.AudioQualityHigh,  // High quality
    SampleRate: 44100,                     // 44.1kHz
}

err := ffmpeg.Audio.RemoveSilence("input.mp3", "output.mp3", silenceConfig, audioConfig)
if err != nil {
    log.Fatalf("Error removing silence: %v", err)
}
```

### Remove Silence from Video

```go
silenceConfig := ffmpego.SilenceConfig{
    MinSilenceLen: ffmpego.SilenceDurationMedium,    // 700ms
    SilenceThresh: ffmpego.SilenceThresholdDefault,  // -30 dB
}

videoConfig := &ffmpego.VideoConfig{
    VideoCodec: ffmpego.VideoCodecH264,    // H.264
    CRF:        ffmpego.VideoQualityHigh,  // High quality
    Preset:     ffmpego.PresetMedium,      // Medium preset
}

err := ffmpeg.Video.RemoveSilence("input.mp4", "output.mp4", silenceConfig, videoConfig)
if err != nil {
    log.Fatalf("Error removing silence: %v", err)
}
```

### Extract Audio from Video

```go
err := ffmpeg.Audio.ExtractFromVideo("input.mp4", "output.mp3")
if err != nil {
    log.Fatalf("Error extracting audio: %v", err)
}
```

## Main Structures

### FFmpeg

The main structure that provides access to the library's functionalities.

```go
type FFmpeg struct {
    Audio *AudioProcessor
    Video *VideoProcessor
}
```

### VideoConfig

Configuration settings for video processing.

```go
type VideoConfig struct {
    TargetResolution string  // Format: "WIDTHxHEIGHT" (e.g., "1920x1080")
    FrameRate        float64 // Target frame rate or 0 for original
    VideoCodec       string  // Video codec or empty for default
    AudioCodec       string  // Audio codec or empty for default
    PreserveCodecs   bool    // Keep original codecs
    CRF              int     // Constant Rate Factor (0-51, lower is better)
    Preset           string  // Encoding preset (ultrafast...veryslow)
    PixelFormat      string  // Pixel format or empty for default
}
```

### AudioConfig

Configuration settings for audio processing.

```go
type AudioConfig struct {
    SampleRate int    // Sample rate in Hz or 0 for original
    Channels   int    // Number of channels or 0 for original
    Quality    int    // Quality (0-9, lower is better) or 0 for default
    Codec      string // Audio codec or empty for default
    BitRate    int    // Bit rate in kbps or 0 for default/variable
}
```

### SilenceConfig

Configuration settings for silence detection and removal.

```go
type SilenceConfig struct {
    MinSilenceLen int // Minimum silence length in milliseconds
    SilenceThresh int // Silence threshold in dB
}
```

## License

This project is licensed under the terms of the MIT license. 