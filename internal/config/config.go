package config

import (
	"os"
	"strconv"
)

type Config struct {
	NatsURL           string
	StreamName        string
	ConsumerGroup     string
	MinioEndpoint     string
	MinioAccessKey    string
	MinioSecretKey    string
	MinioUseSSL       bool
	RawBucket         string
	HLSBucket         string
	BackendWebhookURL string
	BackendBaseURL    string
	FFmpegPath        string
	TempDir           string
}

func Load() *Config {
	useSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))

	return &Config{
		NatsURL:           getEnv("NATS_URL", "nats://localhost:4222"),
		StreamName:        getEnv("STREAM_NAME", "PIPEVID_STREAM"),
		ConsumerGroup:     getEnv("CONSUMER_GROUP", "transcode-workers"),
		MinioEndpoint:     getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:    getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:    getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		MinioUseSSL:       useSSL,
		RawBucket:         getEnv("RAW_BUCKET", "raw-videos"),
		HLSBucket:         getEnv("HLS_BUCKET", "hls-streams"),
		BackendWebhookURL: getEnv("BACKEND_WEBHOOK_URL", "http://localhost:8080/v1/transcode/callback"),
		BackendBaseURL:    getEnv("BACKEND_BASE_URL", "http://localhost:8080"),
		FFmpegPath:        getEnv("FFMPEG_PATH", "ffmpeg"),
		TempDir:           getEnv("TEMP_DIR", "/tmp/transcode"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
