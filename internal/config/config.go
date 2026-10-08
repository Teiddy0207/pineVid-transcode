package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Defaults follow docs/message-contract.md §3.
type Config struct {
	NatsURL      string
	StreamName   string
	ConsumerName string

	// AckWait is how long NATS waits for an ack before redelivering; while a
	// job runs the worker extends it every InProgressInterval.
	AckWait            time.Duration
	InProgressInterval time.Duration
	MaxDeliver         int
	// MaxAckPending is consumer-wide: set it to the number of workers.
	MaxAckPending int
	// NakDelays[i] is the redelivery delay after failed attempt i+1.
	NakDelays []time.Duration

	// DefaultRenditions is used only when a job carries no renditions.
	DefaultRenditions []string

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool
	RawBucket      string
	HLSBucket      string

	// WebhookURL is the backend's signed transcode webhook.
	WebhookURL string
	// BackendBaseURL is used for the admin heartbeat; empty disables it.
	BackendBaseURL string
	FFmpegPath     string
	FFprobePath    string
	TempDir        string
	// InternalWebhookSecret signs the webhook (HMAC key) and authenticates
	// the heartbeat; must match the backend's INTERNAL_WEBHOOK_SECRET.
	InternalWebhookSecret string

	// FaultExitBeforeAck makes the worker exit right after a successful
	// COMPLETED webhook and before acking — test scenario 1 of the proposal
	// (§7.5). Never set it outside that test.
	FaultExitBeforeAck bool
}

func Load() *Config {
	useSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))
	faultExit, _ := strconv.ParseBool(getEnv("FAULT_EXIT_BEFORE_ACK", "false"))

	return &Config{
		NatsURL:      getEnv("NATS_URL", "nats://localhost:4222"),
		StreamName:   getEnv("STREAM_NAME", "VIDEO_JOBS"),
		ConsumerName: getEnv("CONSUMER_NAME", "transcode-worker"),

		AckWait:            getDuration("ACK_WAIT", 60*time.Second),
		InProgressInterval: getDuration("IN_PROGRESS_INTERVAL", 20*time.Second),
		MaxDeliver:         getInt("MAX_DELIVER", 3),
		MaxAckPending:      getInt("MAX_ACK_PENDING", 1),
		NakDelays:          getDurations("NAK_DELAYS", []time.Duration{10 * time.Second, 30 * time.Second}),

		DefaultRenditions: getList("RENDITIONS", []string{"720p", "480p"}),

		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		MinioUseSSL:    useSSL,
		RawBucket:      getEnv("RAW_BUCKET", "raw-videos"),
		HLSBucket:      getEnv("HLS_BUCKET", "hls-streams"),

		WebhookURL:     getEnv("BACKEND_WEBHOOK_URL", "http://localhost:8080/internal/webhooks/transcode"),
		BackendBaseURL: getEnv("BACKEND_BASE_URL", "http://localhost:8080"),
		FFmpegPath:     getEnv("FFMPEG_PATH", "ffmpeg"),
		FFprobePath:    getEnv("FFPROBE_PATH", "ffprobe"),
		TempDir:        getEnv("TEMP_DIR", "/tmp/transcode"),

		InternalWebhookSecret: getEnv("INTERNAL_WEBHOOK_SECRET", ""),
		FaultExitBeforeAck:    faultExit,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(getEnv(key, "")); err == nil && d > 0 {
		return d
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if n, err := strconv.Atoi(getEnv(key, "")); err == nil && n > 0 {
		return n
	}
	return fallback
}

func getList(key string, fallback []string) []string {
	var out []string
	for _, s := range strings.Split(getEnv(key, ""), ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func getDurations(key string, fallback []time.Duration) []time.Duration {
	var out []time.Duration
	for _, s := range getList(key, nil) {
		d, err := time.ParseDuration(s)
		if err != nil || d <= 0 {
			return fallback
		}
		out = append(out, d)
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
