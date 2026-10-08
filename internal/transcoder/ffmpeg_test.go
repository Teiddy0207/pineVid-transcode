package transcoder

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func names(rs []Rendition) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Name
	}
	return out
}

func TestResolveRenditions(t *testing.T) {
	rs, err := ResolveRenditions([]string{"480p", "720p", "480p"})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(names(rs), ","); got != "720p,480p" {
		t.Fatalf("want highest first without duplicates, got %s", got)
	}
	if rs[1].VideoBitrateK != 1000 {
		t.Fatalf("480p must stay at 1000k for the 1.5 Mbps ABR test, got %d", rs[1].VideoBitrateK)
	}

	if _, err := ResolveRenditions([]string{"4k"}); err == nil {
		t.Fatal("unknown rendition must be rejected")
	}
	if _, err := ResolveRenditions(nil); err == nil {
		t.Fatal("empty list must be rejected")
	}
}

func TestSelectForSource(t *testing.T) {
	ladder, _ := ResolveRenditions([]string{"720p", "480p"})

	tests := []struct {
		source int
		want   string
	}{
		{1080, "720p,480p"},
		{720, "720p,480p"},
		{600, "480p"},
		{360, "360p"}, // smaller than every rung: keep the source height
		{0, "720p,480p"},
	}
	for _, tc := range tests {
		if got := strings.Join(names(SelectForSource(ladder, tc.source)), ","); got != tc.want {
			t.Errorf("source %dp: want %s, got %s", tc.source, tc.want, got)
		}
	}
}

func argValue(args []string, flag string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			return args[i+1]
		}
	}
	return ""
}

func TestBuildHLSArgs(t *testing.T) {
	rs, _ := ResolveRenditions([]string{"720p", "480p"})

	withAudio := BuildHLSArgs("in.mp4", "/out", rs, true)
	if got := argValue(withAudio, "-var_stream_map"); got != "v:0,a:0,name:720p v:1,a:1,name:480p" {
		t.Errorf("var_stream_map = %q", got)
	}
	if got := argValue(withAudio, "-force_key_frames"); got != "expr:gte(t,n_forced*4)" {
		t.Errorf("keyframes must be forced every segment, got %q", got)
	}
	if argValue(withAudio, "-hls_time") != "4" || argValue(withAudio, "-b:v:1") != "1000k" || argValue(withAudio, "-filter:v:0") != "scale=-2:720" {
		t.Errorf("unexpected ladder args: %v", withAudio)
	}

	noAudio := BuildHLSArgs("in.mp4", "/out", rs, false)
	if got := argValue(noAudio, "-var_stream_map"); got != "v:0,name:720p v:1,name:480p" {
		t.Errorf("var_stream_map without audio = %q", got)
	}
	for _, a := range noAudio {
		if a == "0:a:0" || a == "aac" {
			t.Fatalf("audio args present for a silent input: %v", noAudio)
		}
	}
}

// TestTranscodeHLS_RealFFmpeg runs the real binaries on a generated 1080p
// clip and checks the master playlist has exactly the 720p and 480p variants.
func TestTranscodeHLS_RealFFmpeg(t *testing.T) {
	for _, bin := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not installed", bin)
		}
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "src.mp4")
	gen := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc2=size=1920x1080:rate=30:duration=9",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=9",
		"-c:v", "libx264", "-preset", "ultrafast", "-c:a", "aac", "-shortest", src)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("generate source: %v\n%s", err, out)
	}

	tc := NewTranscoder("", "", zap.NewNop())
	outDir := filepath.Join(dir, "hls")
	rs, err := tc.TranscodeHLS(context.Background(), src, outDir, []string{"720p", "480p"})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(names(rs), ","); got != "720p,480p" {
		t.Fatalf("produced %s", got)
	}

	master, err := os.ReadFile(filepath.Join(outDir, "master.m3u8"))
	if err != nil {
		t.Fatal(err)
	}
	m := string(master)
	if n := strings.Count(m, "#EXT-X-STREAM-INF"); n != 2 {
		t.Fatalf("want 2 variants, got %d:\n%s", n, m)
	}
	for _, res := range []string{"RESOLUTION=1280x720", "RESOLUTION=854x480"} {
		if !strings.Contains(m, res) {
			t.Errorf("master playlist lacks %s:\n%s", res, m)
		}
	}

	// Segments follow the 4 s keyframe grid (last one may be shorter).
	media, err := os.ReadFile(filepath.Join(outDir, "v480p.m3u8"))
	if err != nil {
		t.Fatal(err)
	}
	for _, match := range regexp.MustCompile(`#EXTINF:([0-9.]+)`).FindAllStringSubmatch(string(media), -1) {
		d, _ := strconv.ParseFloat(match[1], 64)
		if d > SegmentSeconds+0.1 {
			t.Errorf("segment of %.3fs exceeds %ds", d, SegmentSeconds)
		}
	}
}
