package transcoder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// SegmentSeconds is the HLS segment length; keyframes are forced on the same
// grid so every rendition's segments line up and the player can switch
// quality at any segment boundary.
const SegmentSeconds = 4

const (
	maxBitrateRatio = 1.07
	bufsizeRatio    = 1.5
)

// Rendition is one rung of the HLS ladder.
type Rendition struct {
	Name          string
	Height        int
	VideoBitrateK int
	AudioBitrateK int
}

// Ladder lists every rendition a job may request. Experiments use 720p +
// 480p; 480p is kept at ~1.1 Mbps with audio so playback stays smooth on the
// 1.5 Mbps throttled link of criterion A1/A2.
var Ladder = map[string]Rendition{
	"1080p": {Name: "1080p", Height: 1080, VideoBitrateK: 5000, AudioBitrateK: 192},
	"720p":  {Name: "720p", Height: 720, VideoBitrateK: 2800, AudioBitrateK: 128},
	"480p":  {Name: "480p", Height: 480, VideoBitrateK: 1000, AudioBitrateK: 128},
	"360p":  {Name: "360p", Height: 360, VideoBitrateK: 800, AudioBitrateK: 96},
}

// ResolveRenditions maps requested names to ladder rungs, highest first.
func ResolveRenditions(names []string) ([]Rendition, error) {
	seen := map[string]bool{}
	var out []Rendition
	for _, n := range names {
		r, ok := Ladder[strings.TrimSpace(n)]
		if !ok {
			return nil, fmt.Errorf("unknown rendition %q", n)
		}
		if !seen[r.Name] {
			seen[r.Name] = true
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no renditions requested")
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Height > out[j].Height })
	return out, nil
}

// SelectForSource drops rungs taller than the source (no upscaling). If none
// is left, the source is kept at its own height with the lowest rung's
// bitrates. sourceHeight <= 0 (unknown) keeps every rung.
func SelectForSource(rs []Rendition, sourceHeight int) []Rendition {
	if sourceHeight <= 0 {
		return rs
	}
	var out []Rendition
	for _, r := range rs {
		if r.Height <= sourceHeight {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		lowest := rs[len(rs)-1]
		out = []Rendition{{
			Name:          fmt.Sprintf("%dp", sourceHeight),
			Height:        sourceHeight,
			VideoBitrateK: lowest.VideoBitrateK,
			AudioBitrateK: lowest.AudioBitrateK,
		}}
	}
	return out
}

// BuildHLSArgs returns the ffmpeg arguments for a multi-rendition HLS VOD
// with a master playlist at outDir/master.m3u8.
func BuildHLSArgs(input, outDir string, rs []Rendition, hasAudio bool) []string {
	args := []string{
		"-y", "-hide_banner", "-loglevel", "error",
		"-i", input,
		"-preset", "veryfast",
		"-sc_threshold", "0",
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", SegmentSeconds),
	}

	for range rs {
		args = append(args, "-map", "0:v:0")
		if hasAudio {
			args = append(args, "-map", "0:a:0")
		}
	}

	streamMap := make([]string, 0, len(rs))
	for i, r := range rs {
		args = append(args,
			fmt.Sprintf("-filter:v:%d", i), fmt.Sprintf("scale=-2:%d", r.Height),
			fmt.Sprintf("-b:v:%d", i), fmt.Sprintf("%dk", r.VideoBitrateK),
			fmt.Sprintf("-maxrate:v:%d", i), fmt.Sprintf("%dk", int(float64(r.VideoBitrateK)*maxBitrateRatio)),
			fmt.Sprintf("-bufsize:v:%d", i), fmt.Sprintf("%dk", int(float64(r.VideoBitrateK)*bufsizeRatio)),
		)
		entry := fmt.Sprintf("v:%d", i)
		if hasAudio {
			args = append(args, fmt.Sprintf("-b:a:%d", i), fmt.Sprintf("%dk", r.AudioBitrateK))
			entry += fmt.Sprintf(",a:%d", i)
		}
		streamMap = append(streamMap, entry+",name:"+r.Name)
	}

	args = append(args, "-c:v", "libx264", "-profile:v", "main", "-pix_fmt", "yuv420p")
	if hasAudio {
		args = append(args, "-c:a", "aac", "-ar", "48000")
	}

	return append(args,
		"-f", "hls",
		"-hls_time", fmt.Sprint(SegmentSeconds),
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-hls_segment_filename", filepath.Join(outDir, "v%v_file%03d.ts"),
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", strings.Join(streamMap, " "),
		filepath.Join(outDir, "v%v.m3u8"),
	)
}

type Transcoder struct {
	ffmpegPath  string
	ffprobePath string
	logger      *zap.Logger
}

func NewTranscoder(ffmpegPath, ffprobePath string, logger *zap.Logger) *Transcoder {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	if ffprobePath == "" {
		ffprobePath = "ffprobe"
	}
	return &Transcoder{ffmpegPath: ffmpegPath, ffprobePath: ffprobePath, logger: logger}
}

// Probe returns the height of the first video stream and whether the input
// has an audio stream.
func (t *Transcoder) Probe(ctx context.Context, input string) (height int, hasAudio bool, err error) {
	out, err := exec.CommandContext(ctx, t.ffprobePath,
		"-v", "error", "-show_entries", "stream=codec_type,height", "-of", "json", input,
	).Output()
	if err != nil {
		return 0, false, fmt.Errorf("ffprobe %s: %w", input, err)
	}

	var probe struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Height    int    `json:"height"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &probe); err != nil {
		return 0, false, fmt.Errorf("ffprobe output: %w", err)
	}

	for _, s := range probe.Streams {
		switch s.CodecType {
		case "video":
			if height == 0 {
				height = s.Height
			}
		case "audio":
			hasAudio = true
		}
	}
	if height == 0 {
		return 0, false, fmt.Errorf("no video stream in %s", input)
	}
	return height, hasAudio, nil
}

// TranscodeHLS converts inputPath into an HLS ladder under outputDir and
// returns the renditions actually produced.
func (t *Transcoder) TranscodeHLS(ctx context.Context, inputPath, outputDir string, renditions []string) ([]Rendition, error) {
	requested, err := ResolveRenditions(renditions)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	height, hasAudio, err := t.Probe(ctx, inputPath)
	if err != nil {
		return nil, err
	}
	rs := SelectForSource(requested, height)

	t.logger.Info("Starting FFmpeg HLS transcoding",
		zap.String("input", inputPath),
		zap.Int("sourceHeight", height),
		zap.Bool("hasAudio", hasAudio),
		zap.Int("renditions", len(rs)),
	)

	cmd := exec.CommandContext(ctx, t.ffmpegPath, BuildHLSArgs(inputPath, outputDir, rs, hasAudio)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg transcoding failed: %w\nOutput: %s", err, string(output))
	}

	if _, err := os.Stat(filepath.Join(outputDir, "master.m3u8")); err != nil {
		return nil, fmt.Errorf("master.m3u8 was not generated: %w", err)
	}
	return rs, nil
}
