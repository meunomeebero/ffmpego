# 🎬 FFmpego

> A friendly, easy-to-use Go library for working with video and audio files powered by FFmpeg

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.20-blue)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

FFmpego makes video and audio processing simple and intuitive. Whether you're a beginner or an experienced developer, you can start working with media files in just a few lines of code!

---

## 📋 Table of Contents

- [✨ Features](#-features)
- [🚀 Quick Start](#-quick-start)
- [📦 Installation](#-installation)
- [🎯 Basic Usage](#-basic-usage)
- [📖 Complete API Reference](#-complete-api-reference)
- [🎓 Examples](#-examples)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)

---

## ✨ Features

- 🎥 **Video Processing** - Get info, convert formats, extract segments
- 🎵 **Audio Processing** - Manipulate audio files with ease
- 🔇 **Silence Detection** - Automatically detect silent parts in media
- ✂️ **Segment Extraction** - Cut specific parts from your files
- 🔗 **Concatenation** - Join multiple files together
- 🎨 **Format Conversion** - Convert between different formats and qualities
- 🎯 **Simple API** - Designed for developers of all skill levels
- 📱 **Aspect Ratio Support** - Easy conversion for different screen sizes (16:9, 9:16, etc.)

---

## 🚀 Quick Start

Here's how to get video information in just 3 lines:

```go
v, _ := video.New("my-video.mp4")
info, _ := v.GetInfo()
fmt.Printf("Video: %dx%d, %.2f fps\n", info.Width, info.Height, info.FrameRate)
```

That's it! 🎉

---

## 📦 Installation

### Step 1: Install FFmpeg

FFmpego uses [FFmpeg](https://ffmpeg.org/) under the hood, so you need it installed first:

**macOS:**
```bash
brew install ffmpeg
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install ffmpeg
```

**Windows:**
Download from [ffmpeg.org](https://ffmpeg.org/download.html) and add to your PATH.

### Step 2: Install FFmpego

```bash
go get github.com/meunomeebero/ffmpego
```

---

## 🎯 Basic Usage

### Working with Videos 🎥

```go
import "github.com/meunomeebero/ffmpego/video"

// Open a video file
v, err := video.New("input.mp4")
if err != nil {
    log.Fatal(err)
}

// Get video information
info, _ := v.GetInfo()
fmt.Printf("Resolution: %dx%d\n", info.Width, info.Height)
fmt.Printf("Duration: %.2f seconds\n", info.Duration)
fmt.Printf("Frame Rate: %.2f fps\n", info.FrameRate)

// Detect silent parts (great for removing boring silence from videos!)
silenceConfig := video.SilenceConfig{
    MinSilenceDuration: 700,  // 700 milliseconds
    SilenceThreshold:   -30,  // -30 decibels
}
segments, _ := v.DetectSilence(silenceConfig)
fmt.Printf("Found %d parts with audio\n", len(segments))

// Extract a specific part (from 10 to 30 seconds)
v.ExtractSegment("clip.mp4", 10.0, 30.0, nil)

// Convert to a different format
config := video.ConvertConfig{
    Resolution:  "1920x1080",           // Full HD
    AspectRatio: video.AspectRatio16x9, // Widescreen
    VideoCodec:  video.CodecH264,       // Most compatible format
    Quality:     23,                     // Good quality (lower = better)
}
v.Convert("output.mp4", config)
```

### Working with Audio 🎵

```go
import "github.com/meunomeebero/ffmpego/audio"

// Open an audio file
a, err := audio.New("song.mp3")
if err != nil {
    log.Fatal(err)
}

// Get audio information
info, _ := a.GetInfo()
fmt.Printf("Sample Rate: %d Hz\n", info.SampleRate)
fmt.Printf("Channels: %d\n", info.Channels)
fmt.Printf("Duration: %.2f seconds\n", info.Duration)

// Detect silent parts
silenceConfig := audio.SilenceConfig{
    MinSilenceDuration: 500,  // 500ms
    SilenceThreshold:   -40,  // -40dB
}
segments, _ := a.DetectSilence(silenceConfig)

// Extract a portion of the audio
a.ExtractSegment("clip.mp3", 10.0, 30.0, nil)

// Convert to high quality
config := audio.ConvertConfig{
    SampleRate: audio.SampleRate48000, // Professional quality
    Channels:   2,                      // Stereo
    Codec:      audio.CodecAAC,         // Modern, efficient format
    Quality:    2,                      // High quality
    Bitrate:    320,                    // 320 kbps
}
a.Convert("output.m4a", config)
```

### Joining Multiple Files 🔗

```go
// Combine video segments
video.ConcatenateSegments([]string{
    "part1.mp4",
    "part2.mp4",
    "part3.mp4",
}, "final-video.mp4", nil)

// Combine audio files
audio.ConcatenateSegments([]string{
    "intro.mp3",
    "main.mp3",
    "outro.mp3",
}, "complete-audio.mp3", nil)
```

---

## 📖 Complete API Reference

### Video Functions

#### `video.New(path string) (*Video, error)`
Opens a video file for processing.

**Example:**
```go
v, err := video.New("my-video.mp4")
```

---

#### `v.GetInfo() (*Info, error)`
Gets detailed information about the video file.

**Returns:**
- Width and Height (in pixels)
- Duration (in seconds)
- Frame Rate ([FPS](https://en.wikipedia.org/wiki/Frame_rate))
- [Codec](https://en.wikipedia.org/wiki/Video_codec) information
- File size

**Example:**
```go
info, _ := v.GetInfo()
fmt.Printf("Video is %dx%d\n", info.Width, info.Height)
```

---

#### `v.DetectSilence(config SilenceConfig) ([]Segment, error)`
Finds all the parts of the video that have sound (non-silent segments).

**Useful for:** Automatically removing silent parts from recordings, lectures, or podcasts.

**Example:**
```go
config := video.SilenceConfig{
    MinSilenceDuration: 700,  // Silence must be at least 700ms
    SilenceThreshold:   -30,  // Volume below -30dB is considered silence
}
segments, _ := v.DetectSilence(config)
```

---

#### `v.ExtractSegment(outputPath string, startTime, endTime float64, config *ConvertConfig) error`
Cuts out a specific portion of the video.

**Example:**
```go
// Extract from 1 minute 30 seconds to 2 minutes
v.ExtractSegment("clip.mp4", 90.0, 120.0, nil)
```

---

#### `v.Convert(outputPath string, config ConvertConfig) error`
Converts the video to a different format, quality, or size.

**Example:**
```go
config := video.ConvertConfig{
    Resolution:  "1280x720",            // HD resolution
    AspectRatio: video.AspectRatio16x9, // Widescreen format
    Quality:     23,                     // Good quality
}
v.Convert("converted.mp4", config)
```

---

#### `video.ConcatenateSegments(segments []string, outputPath string, config *ConvertConfig) error`
Joins multiple video files into one.

**Example:**
```go
video.ConcatenateSegments([]string{
    "intro.mp4",
    "main-content.mp4",
    "outro.mp4",
}, "final.mp4", nil)
```

---

### Audio Functions

#### `audio.New(path string) (*Audio, error)`
Opens an audio file for processing.

---

#### `a.GetInfo() (*Info, error)`
Gets detailed information about the audio file.

**Returns:**
- [Sample Rate](https://en.wikipedia.org/wiki/Sampling_(signal_processing)) (quality of the audio)
- Channels (mono, stereo, etc.)
- Duration
- [Codec](https://en.wikipedia.org/wiki/Audio_codec)
- [Bitrate](https://en.wikipedia.org/wiki/Bit_rate) (quality metric)

---

#### `a.DetectSilence(config SilenceConfig) ([]Segment, error)`
Finds all the parts of the audio that have sound.

---

#### `a.ExtractSegment(outputPath string, startTime, endTime float64, config *ConvertConfig) error`
Cuts out a specific portion of the audio.

---

#### `a.Convert(outputPath string, config ConvertConfig) error`
Converts the audio to a different format or quality.

---

#### `audio.ConcatenateSegments(segments []string, outputPath string, config *ConvertConfig) error`
Joins multiple audio files into one.

---

## 🎛️ Configuration Options

### Video Configuration

```go
type ConvertConfig struct {
    Resolution  string      // Example: "1920x1080", "1280x720"
    AspectRatio AspectRatio // Screen format (see below)
    FrameRate   float64     // Frames per second (30, 60, etc.)
    VideoCodec  string      // Video compression format
    AudioCodec  string      // Audio compression format
    Quality     int         // 0-51 (lower = better quality, bigger file)
    Preset      string      // Encoding speed (see below)
    Bitrate     int         // Quality in kbps
}
```

**Common [Codecs](https://en.wikipedia.org/wiki/Video_codec):**
- `video.CodecH264` - H.264 (most compatible, works everywhere)
- `video.CodecH265` - H.265 (smaller files, newer)
- `video.CodecVP9` - VP9 (great for web)

**Encoding Presets:**
- `video.PresetUltrafast` - Very fast, larger files
- `video.PresetFast` - Fast encoding
- `video.PresetMedium` - Balanced (recommended)
- `video.PresetSlow` - Slow, better quality

**[Aspect Ratios](https://en.wikipedia.org/wiki/Aspect_ratio_(image)):**
- `video.AspectRatio16x9` - Widescreen (YouTube, TV)
- `video.AspectRatio9x16` - Vertical (Instagram Stories, TikTok)
- `video.AspectRatio4x3` - Old TV format
- `video.AspectRatio1x1` - Square (Instagram posts)
- `video.AspectRatio21x9` - Ultra-wide (cinema)

---

### Audio Configuration

```go
type ConvertConfig struct {
    SampleRate int    // Quality (44100, 48000 Hz)
    Channels   int    // 1 = mono, 2 = stereo
    Codec      string // Compression format
    Quality    int    // 0-9 (lower = better)
    Bitrate    int    // Quality in kbps
}
```

**Common [Codecs](https://en.wikipedia.org/wiki/Audio_codec):**
- `audio.CodecAAC` - AAC (modern, efficient)
- `audio.CodecMP3` - MP3 (universal compatibility)
- `audio.CodecFLAC` - FLAC (no quality loss, big files)
- `audio.CodecOpus` - Opus (best for voice)

**[Sample Rates](https://en.wikipedia.org/wiki/Sampling_(signal_processing)):**
- `audio.SampleRate44100` - CD quality (44.1 kHz)
- `audio.SampleRate48000` - Professional audio (48 kHz)

---

## 🎓 Examples

Check out the [`examples/`](examples/) folder for complete, working examples including:

1. 📊 Getting video information
2. 🔇 Detecting and working with silence
3. 🎨 Converting videos to different formats
4. 📱 Creating vertical videos (for mobile)
5. 🎵 Audio processing and conversion
6. ✂️ Extracting segments
7. 🔗 Joining multiple files

Each example is fully documented and ready to run!

---

## 🤝 Contributing

We welcome contributions from developers of all skill levels! Here's how you can help:

### 🐛 Found a Bug?

1. Check if it's already reported in [Issues](https://github.com/meunomeebero/ffmpego/issues)
2. If not, create a new issue with:
   - A clear title
   - Steps to reproduce the problem
   - What you expected to happen
   - What actually happened

### 💡 Have a Feature Idea?

Open an issue with the tag `enhancement` and describe:
- What you want to do
- Why it would be useful
- How you imagine it would work

### 🔧 Want to Contribute Code?

1. **Fork the repository** - Click the "Fork" button at the top right
2. **Create a branch** - `git checkout -b feature/my-new-feature`
3. **Make your changes** - Write clean, documented code
4. **Test your changes** - Make sure everything works
5. **Commit your changes** - `git commit -am 'Add some feature'`
6. **Push to the branch** - `git push origin feature/my-new-feature`
7. **Create a Pull Request** - Describe what you changed and why

### 📝 Code Guidelines

- Write simple, clear code that beginners can understand
- Add comments to explain complex parts
- Follow Go's standard formatting (`gofmt`)
- Keep functions small and focused on one task
- Add examples for new features

### 🧪 Running Tests

```bash
go test ./...
```

### 💬 Questions?

Feel free to open an issue with the tag `question` - we're here to help!

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Built with [FFmpeg](https://ffmpeg.org/) - the powerful multimedia framework
- Inspired by the need for simple, beginner-friendly media processing in Go
- Thanks to all our contributors! ❤️

---

## 📬 Contact

- **Issues:** [GitHub Issues](https://github.com/meunomeebero/ffmpego/issues)
- **Discussions:** [GitHub Discussions](https://github.com/meunomeebero/ffmpego/discussions)

---

<div align="center">

**Made with ❤️ for the Go community**

If this project helped you, consider giving it a ⭐!

</div>
