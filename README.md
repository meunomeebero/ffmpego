# FFmpego

> A friendly, easy-to-use Go library for working with video and audio files powered by FFmpeg

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.22-blue)](https://go.dev/)
[![Tests](https://github.com/meunomeebero/ffmpego/actions/workflows/test.yml/badge.svg)](https://github.com/meunomeebero/ffmpego/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

FFmpego makes video and audio processing simple and intuitive. You don't need to know about codecs, bitrates, or encoding — just tell it what you want and it handles the rest.

---

## Features

- **Silence Removal** - Automatically cut silent parts from videos and audio
- **Video Processing** - Get info, convert formats, extract segments
- **Audio Processing** - Manipulate audio files with ease
- **Segment Extraction** - Cut specific parts from your files (fast, no quality loss)
- **Concatenation** - Join multiple files together
- **Format Conversion** - Convert between different formats and qualities
- **Aspect Ratio Support** - Easy conversion for different screen sizes (16:9, 9:16, etc.)
- **Automatic Checks** - Verifies ffmpeg is installed when you start

---

## Quick Start

Remove silence from a video in 4 lines:

```go
v, err := video.New("my-video.mp4")
if err != nil {
    log.Fatal(err)
}

err = v.RemoveSilence("clean-video.mp4", video.SilenceConfig{})
if err != nil {
    log.Fatal(err)
}
```

That's it — no need to configure anything. Sensible defaults are applied automatically.

---

## Installation

### Step 1: Install FFmpeg

FFmpego uses [FFmpeg](https://ffmpeg.org/) under the hood. The library checks for it automatically and gives you a clear error if it's missing.

**macOS:**
```bash
brew install ffmpeg
```

**Ubuntu/Debian:**
```bash
sudo apt update && sudo apt install ffmpeg
```

**Windows:**
Download from [ffmpeg.org](https://ffmpeg.org/download.html) and add to your PATH.

### Step 2: Install FFmpego

```bash
go get github.com/meunomeebero/ffmpego
```

---

## Basic Usage

### Remove Silence (the main feature)

The most common use case — remove breathing pauses, dead air, and silence from recordings:

```go
import "github.com/meunomeebero/ffmpego/video"

v, err := video.New("recording.mp4")
if err != nil {
    log.Fatal(err)
}

// With default settings (works great for most videos)
err = v.RemoveSilence("clean.mp4", video.SilenceConfig{})

// Or fine-tune the sensitivity
err = v.RemoveSilence("clean.mp4", video.SilenceConfig{
    MinSilenceDuration: video.SilenceDurationMedium,    // 700ms - balanced
    SilenceThreshold:   video.SilenceThresholdModerate, // -30dB - good for most videos
})
```

Works for audio too:

```go
import "github.com/meunomeebero/ffmpego/audio"

a, err := audio.New("podcast.mp3")
if err != nil {
    log.Fatal(err)
}

err = a.RemoveSilence("clean-podcast.mp3", audio.SilenceConfig{})
```

### Working with Videos

```go
import "github.com/meunomeebero/ffmpego/video"

// Open a video file
v, err := video.New("input.mp4")
if err != nil {
    log.Fatal(err)
}

// Get video information
info, err := v.GetInfo()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Resolution: %dx%d\n", info.Width, info.Height)
fmt.Printf("Duration: %.2f seconds\n", info.Duration)
fmt.Printf("Frame Rate: %.2f fps\n", info.FrameRate)

// Extract a clip (from 10s to 30s) — fast, preserves original quality
err = v.ExtractSegment("clip.mp4", 10.0, 30.0, nil)

// Convert to a different format
err = v.Convert("output.mp4", video.ConvertConfig{
    Resolution: "1920x1080",
    VideoCodec: video.CodecH264,
    Quality:    23,
})
```

### Working with Audio

```go
import "github.com/meunomeebero/ffmpego/audio"

// Open an audio file
a, err := audio.New("song.mp3")
if err != nil {
    log.Fatal(err)
}

// Get audio information
info, err := a.GetInfo()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Sample Rate: %d Hz\n", info.SampleRate)
fmt.Printf("Channels: %d\n", info.Channels)
fmt.Printf("Duration: %.2f seconds\n", info.Duration)

// Extract a portion — fast, preserves original quality
err = a.ExtractSegment("clip.mp3", 10.0, 30.0, nil)

// Convert to a different format
err = a.Convert("output.m4a", audio.ConvertConfig{
    SampleRate: audio.SampleRate48000,
    Channels:   2,
    Codec:      audio.CodecAAC,
    Bitrate:    320,
})
```

### Joining Multiple Files

```go
// Combine video segments
err := video.ConcatenateSegments([]string{
    "part1.mp4",
    "part2.mp4",
    "part3.mp4",
}, "final-video.mp4", nil)

// Combine audio files
err := audio.ConcatenateSegments([]string{
    "intro.mp3",
    "main.mp3",
    "outro.mp3",
}, "complete-audio.mp3", nil)
```

---

## API Reference

### Video

| Function | Description |
|---|---|
| `video.New(path)` | Open a video file. Checks if ffmpeg is installed. |
| `v.GetInfo()` | Get resolution, duration, fps, codec, file size. Results are cached. |
| `v.RemoveSilence(output, config)` | Remove silent parts. Parallel processing, preserves quality. |
| `v.GetNonSilentSegments(config)` | Detect which parts have audio. Returns time ranges. |
| `v.ExtractSegment(output, start, end, config)` | Cut a clip. Pass `nil` for config to keep original quality. |
| `v.Convert(output, config)` | Convert format, resolution, codec, quality. |
| `video.ConcatenateSegments(paths, output, config)` | Join multiple video files into one. |

### Audio

| Function | Description |
|---|---|
| `audio.New(path)` | Open an audio file. Checks if ffmpeg is installed. |
| `a.GetInfo()` | Get sample rate, channels, codec, bitrate, duration. Results are cached. |
| `a.RemoveSilence(output, config)` | Remove silent parts. Parallel processing, preserves quality. |
| `a.GetNonSilentSegments(config)` | Detect which parts have audio. Returns time ranges. |
| `a.ExtractSegment(output, start, end, config)` | Cut a clip. Pass `nil` for config to keep original quality. |
| `a.Convert(output, config)` | Convert format, sample rate, codec, quality. |
| `audio.ConcatenateSegments(paths, output, config)` | Join multiple audio files into one. |

---

## Configuration

### Silence Detection

```go
config := video.SilenceConfig{
    MinSilenceDuration: video.SilenceDurationMedium,    // How long a pause must be to count as "silence"
    SilenceThreshold:   video.SilenceThresholdModerate, // How quiet must it be to count as "silence"
}
```

You can also pass `SilenceConfig{}` with no fields — sensible defaults are applied automatically (700ms, -30dB).

**Duration presets** (how long a pause must be):

| Constant | Duration | Use case |
|---|---|---|
| `SilenceDurationVeryShort` | 200ms | Very aggressive, catches every tiny pause |
| `SilenceDurationShort` | 500ms | Good for fast-paced content |
| `SilenceDurationMedium` | 700ms | **Recommended** — balanced for most content |
| `SilenceDurationLong` | 1000ms | Only removes long pauses |
| `SilenceDurationVeryLong` | 2000ms | Very conservative, only obvious gaps |

**Threshold presets** (how quiet is "silent"):

| Constant | Level | Use case |
|---|---|---|
| `SilenceThresholdVeryStrict` | -50dB | Detects even very quiet sounds |
| `SilenceThresholdStrict` | -40dB | Good for clean audio recordings |
| `SilenceThresholdModerate` | -30dB | **Recommended** — works for most content |
| `SilenceThresholdRelaxed` | -20dB | Only loud parts are kept |
| `SilenceThresholdVeryRelaxed` | -10dB | Only very loud parts are kept |

### Video Conversion

```go
config := video.ConvertConfig{
    Resolution:  "1920x1080",            // e.g., "1280x720", "3840x2160"
    AspectRatio: video.AspectRatio16x9,  // Screen format
    FrameRate:   30,                     // Frames per second
    VideoCodec:  video.CodecH264,        // Compression format
    AudioCodec:  video.CodecAAC,         // Audio compression
    Quality:     23,                     // 0-51, lower = better quality
    Preset:      video.PresetMedium,     // Speed vs quality trade-off
    Bitrate:     5000,                   // kbps
}
```

**Video [codecs](https://en.wikipedia.org/wiki/Video_codec):**
- `video.CodecH264` — Most compatible, works everywhere
- `video.CodecH265` — Smaller files, newer devices
- `video.CodecVP9` — Great for web

**Encoding [presets](https://trac.ffmpeg.org/wiki/Encode/H.264#Preset):**
- `video.PresetUltrafast` — Very fast encoding, larger files
- `video.PresetFast` — Fast encoding
- `video.PresetMedium` — Balanced (recommended)
- `video.PresetSlow` — Slow encoding, better compression

**[Aspect ratios](https://en.wikipedia.org/wiki/Aspect_ratio_(image)):**
- `video.AspectRatio16x9` — Widescreen (YouTube, TV)
- `video.AspectRatio9x16` — Vertical (Instagram Stories, TikTok)
- `video.AspectRatio4x3` — Classic TV format
- `video.AspectRatio1x1` — Square (Instagram posts)
- `video.AspectRatio21x9` — Ultra-wide (cinema)

### Audio Conversion

```go
config := audio.ConvertConfig{
    SampleRate: audio.SampleRate48000, // Audio quality
    Channels:   2,                     // 1 = mono, 2 = stereo
    Codec:      audio.CodecAAC,        // Compression format
    Bitrate:    320,                   // kbps
}
```

**Audio [codecs](https://en.wikipedia.org/wiki/Audio_codec):**
- `audio.CodecAAC` — Modern, efficient
- `audio.CodecMP3` — Universal compatibility
- `audio.CodecFLAC` — Lossless (no quality loss, larger files)
- `audio.CodecOpus` — Best for voice and streaming

**[Sample rates](https://en.wikipedia.org/wiki/Sampling_(signal_processing)):**
- `audio.SampleRate44100` — CD quality (44.1 kHz)
- `audio.SampleRate48000` — Professional audio (48 kHz)

---

## Examples

Check out the [`examples/`](examples/) folder:

- **[`remove_video_silence/`](examples/remove_video_silence/)** — Remove silence from a video file (the main use case)
- **[`main.go`](examples/main.go)** — 10 examples covering all features: info, silence detection, conversion, extraction, concatenation

Run an example:

```bash
go run examples/remove_video_silence/main.go input.mp4 output.mp4
```

---

## Contributing

Contributions are welcome from developers of all skill levels!

- **Found a bug?** Open an [issue](https://github.com/meunomeebero/ffmpego/issues) with steps to reproduce
- **Feature idea?** Open an issue with the `enhancement` tag
- **Want to contribute code?** Fork, create a branch, make changes, open a PR

### Code Guidelines

- Write simple, clear code that beginners can understand
- Follow Go formatting (`gofmt`)
- Keep functions small and focused
- Add tests for new logic — the CI runs 78 tests with race detection on every PR

```bash
go test ./... -race
```

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

Built with [FFmpeg](https://ffmpeg.org/) | [Issues](https://github.com/meunomeebero/ffmpego/issues) | [Discussions](https://github.com/meunomeebero/ffmpego/discussions)
