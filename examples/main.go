package main

import (
	"fmt"
	"log"

	"github.com/meunomeebero/ffmpego/audio"
	"github.com/meunomeebero/ffmpego/video"
)

func main() {
	// ==========================================
	// Exemplo 1: Obter informações de um vídeo
	// ==========================================
	fmt.Println("=== Example 1: Get Video Info ===")

	v, err := video.New("input.mp4")
	if err != nil {
		log.Printf("Error opening video: %v", err)
	} else {
		info, err := v.GetInfo()
		if err != nil {
			log.Printf("Error getting video info: %v", err)
		} else {
			fmt.Printf("Video: %dx%d, %.2f fps, Duration: %.2fs\n",
				info.Width, info.Height, info.FrameRate, info.Duration)
		}
	}

	// ==========================================
	// Exemplo 2: Detectar silêncio e extrair segmentos
	// ==========================================
	fmt.Println("\n=== Example 2: Detect Silence and Extract Segments ===")

	v, err = video.New("input.mp4")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		silenceConfig := video.SilenceConfig{
			MinSilenceDuration: 700,
			SilenceThreshold:   -30,
		}

		segments, err := v.DetectSilence(silenceConfig)
		if err != nil {
			log.Printf("Error detecting silence: %v", err)
		} else {
			fmt.Printf("Found %d non-silent segments:\n", len(segments))
			for i, seg := range segments {
				fmt.Printf("  Segment %d: %.2fs - %.2fs (%.2fs)\n",
					i+1, seg.StartTime, seg.EndTime, seg.Duration)
			}

			// Extrair o primeiro segmento
			if len(segments) > 0 {
				err = v.ExtractSegment("segment_001.mp4",
					segments[0].StartTime,
					segments[0].EndTime,
					nil)
				if err != nil {
					log.Printf("Error extracting segment: %v", err)
				} else {
					fmt.Println("First segment extracted successfully!")
				}
			}
		}
	}

	// ==========================================
	// Exemplo 3: Converter vídeo para 16:9 em HD
	// ==========================================
	fmt.Println("\n=== Example 3: Convert Video to 16:9 HD ===")

	v, err = video.New("input.mp4")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		config := video.ConvertConfig{
			Resolution:  "1280x720", // HD
			AspectRatio: video.AspectRatio16x9,
			FrameRate:   30,
			VideoCodec:  video.CodecH264,
			AudioCodec:  video.CodecAAC,
			Quality:     23, // Boa qualidade
			Preset:      video.PresetMedium,
		}

		err = v.Convert("output_hd.mp4", config)
		if err != nil {
			log.Printf("Error converting video: %v", err)
		} else {
			fmt.Println("Video converted successfully!")
		}
	}

	// ==========================================
	// Exemplo 4: Converter vídeo para formato vertical (9:16)
	// ==========================================
	fmt.Println("\n=== Example 4: Convert Video to Vertical Format (9:16) ===")

	v, err = video.New("input.mp4")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		config := video.ConvertConfig{
			AspectRatio: video.AspectRatio9x16,
			VideoCodec:  video.CodecH264,
			Quality:     20,
			Preset:      video.PresetFast,
		}

		err = v.Convert("output_vertical.mp4", config)
		if err != nil {
			log.Printf("Error converting video: %v", err)
		} else {
			fmt.Println("Vertical video created successfully!")
		}
	}

	// ==========================================
	// Exemplo 5: Obter informações de um áudio
	// ==========================================
	fmt.Println("\n=== Example 5: Get Audio Info ===")

	a, err := audio.New("audio.mp3")
	if err != nil {
		log.Printf("Error opening audio: %v", err)
	} else {
		info, err := a.GetInfo()
		if err != nil {
			log.Printf("Error getting audio info: %v", err)
		} else {
			fmt.Printf("Audio: %d Hz, %d channels, Codec: %s, Duration: %.2fs\n",
				info.SampleRate, info.Channels, info.Codec, info.Duration)
		}
	}

	// ==========================================
	// Exemplo 6: Detectar silêncio no áudio
	// ==========================================
	fmt.Println("\n=== Example 6: Detect Silence in Audio ===")

	a, err = audio.New("audio.mp3")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		silenceConfig := audio.SilenceConfig{
			MinSilenceDuration: 500,
			SilenceThreshold:   -40,
		}

		segments, err := a.DetectSilence(silenceConfig)
		if err != nil {
			log.Printf("Error detecting silence: %v", err)
		} else {
			fmt.Printf("Found %d non-silent audio segments:\n", len(segments))
			for i, seg := range segments {
				fmt.Printf("  Segment %d: %.2fs - %.2fs\n",
					i+1, seg.StartTime, seg.EndTime)
			}
		}
	}

	// ==========================================
	// Exemplo 7: Converter áudio para alta qualidade
	// ==========================================
	fmt.Println("\n=== Example 7: Convert Audio to High Quality ===")

	a, err = audio.New("audio.mp3")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		config := audio.ConvertConfig{
			SampleRate: audio.SampleRate48000,
			Channels:   2, // Stereo
			Codec:      audio.CodecAAC,
			Quality:    2, // Alta qualidade
			Bitrate:    320,
		}

		err = a.Convert("output_hq_audio.m4a", config)
		if err != nil {
			log.Printf("Error converting audio: %v", err)
		} else {
			fmt.Println("Audio converted successfully!")
		}
	}

	// ==========================================
	// Exemplo 8: Extrair segmento de áudio
	// ==========================================
	fmt.Println("\n=== Example 8: Extract Audio Segment ===")

	a, err = audio.New("audio.mp3")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		// Extrair dos 10s aos 30s
		err = a.ExtractSegment("audio_segment.mp3", 10.0, 30.0, nil)
		if err != nil {
			log.Printf("Error extracting audio segment: %v", err)
		} else {
			fmt.Println("Audio segment extracted successfully!")
		}
	}

	// ==========================================
	// Exemplo 9: Concatenar segmentos de vídeo
	// ==========================================
	fmt.Println("\n=== Example 9: Concatenate Video Segments ===")

	segmentPaths := []string{
		"segment_001.mp4",
		"segment_002.mp4",
		"segment_003.mp4",
	}

	err = video.ConcatenateSegments(segmentPaths, "concatenated.mp4", nil)
	if err != nil {
		log.Printf("Error concatenating video segments: %v", err)
	} else {
		fmt.Println("Video segments concatenated successfully!")
	}

	// ==========================================
	// Exemplo 10: Concatenar segmentos de áudio
	// ==========================================
	fmt.Println("\n=== Example 10: Concatenate Audio Segments ===")

	audioSegmentPaths := []string{
		"audio_segment_001.mp3",
		"audio_segment_002.mp3",
		"audio_segment_003.mp3",
	}

	err = audio.ConcatenateSegments(audioSegmentPaths, "concatenated_audio.mp3", nil)
	if err != nil {
		log.Printf("Error concatenating audio segments: %v", err)
	} else {
		fmt.Println("Audio segments concatenated successfully!")
	}
}
