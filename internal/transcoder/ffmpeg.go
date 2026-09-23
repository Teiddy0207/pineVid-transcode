package transcoder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"
)

type Transcoder struct {
	ffmpegPath string
	logger     *zap.Logger
}

func NewTranscoder(ffmpegPath string, logger *zap.Logger) *Transcoder {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	return &Transcoder{
		ffmpegPath: ffmpegPath,
		logger:     logger,
	}
}

// TranscodeHLS converts an input video file into a multi-bitrate HLS ladder with master.m3u8
func (t *Transcoder) TranscodeHLS(ctx context.Context, inputPath, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	masterPlaylistPath := filepath.Join(outputDir, "master.m3u8")

	// Probe: check if input has audio stream
	probeCmd := exec.CommandContext(ctx, t.ffmpegPath, "-i", inputPath, "-hide_banner")
	probeOutput, _ := probeCmd.CombinedOutput()
	probeStr := string(probeOutput)

	hasAudio := false
	if contains(probeStr, "Audio:") {
		hasAudio = true
	}

	t.logger.Info("Starting FFmpeg HLS transcoding",
		zap.String("input", inputPath),
		zap.String("outputDir", outputDir),
		zap.Bool("hasAudio", hasAudio),
	)

	var args []string
	if hasAudio {
		args = []string{
			"-y",
			"-i", inputPath,
			"-preset", "veryfast",
			"-g", "48",
			"-sc_threshold", "0",
			// 4 video + 4 audio output streams
			"-map", "0:v:0", "-map", "0:a:0",
			"-map", "0:v:0", "-map", "0:a:0",
			"-map", "0:v:0", "-map", "0:a:0",
			"-map", "0:v:0", "-map", "0:a:0",
			"-filter:v:0", "scale=-2:1080", "-b:v:0", "5000k",
			"-filter:v:1", "scale=-2:720", "-b:v:1", "2800k",
			"-filter:v:2", "scale=-2:480", "-b:v:2", "1400k",
			"-filter:v:3", "scale=-2:360", "-b:v:3", "800k",
			"-c:v", "h264", "-c:a", "aac", "-ar", "48000",
			"-f", "hls",
			"-hls_time", "4",
			"-hls_playlist_type", "vod",
			"-hls_flags", "independent_segments",
			"-hls_segment_filename", filepath.Join(outputDir, "v%v_file%03d.ts"),
			"-master_pl_name", "master.m3u8",
			"-var_stream_map", "v:0,a:0,name:1080p v:1,a:1,name:720p v:2,a:2,name:480p v:3,a:3,name:360p",
			filepath.Join(outputDir, "v%v.m3u8"),
		}
	} else {
		args = []string{
			"-y",
			"-i", inputPath,
			"-preset", "veryfast",
			"-g", "48",
			"-sc_threshold", "0",
			"-an",
			"-map", "0:v:0",
			"-map", "0:v:0",
			"-map", "0:v:0",
			"-map", "0:v:0",
			"-filter:v:0", "scale=-2:1080", "-b:v:0", "5000k",
			"-filter:v:1", "scale=-2:720", "-b:v:1", "2800k",
			"-filter:v:2", "scale=-2:480", "-b:v:2", "1400k",
			"-filter:v:3", "scale=-2:360", "-b:v:3", "800k",
			"-c:v", "h264",
			"-f", "hls",
			"-hls_time", "4",
			"-hls_playlist_type", "vod",
			"-hls_flags", "independent_segments",
			"-hls_segment_filename", filepath.Join(outputDir, "v%v_file%03d.ts"),
			"-master_pl_name", "master.m3u8",
			"-var_stream_map", "v:0,name:1080p v:1,name:720p v:2,name:480p v:3,name:360p",
			filepath.Join(outputDir, "v%v.m3u8"),
		}
	}

	cmd := exec.CommandContext(ctx, t.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.logger.Error("FFmpeg execution failed",
			zap.Error(err),
			zap.String("ffmpegOutput", string(output)),
		)
		return fmt.Errorf("ffmpeg transcoding failed: %w\nOutput: %s", err, string(output))
	}

	if _, statErr := os.Stat(masterPlaylistPath); os.IsNotExist(statErr) {
		return fmt.Errorf("master.m3u8 was not generated. FFmpeg output: %s", string(output))
	}

	t.logger.Info("FFmpeg HLS transcoding completed successfully",
		zap.String("masterPlaylist", masterPlaylistPath),
	)
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
