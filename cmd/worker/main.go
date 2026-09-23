package main

import (
	"context"
	"os"
	"os/signal"
	"sonla/transcode/internal/config"
	"sonla/transcode/internal/storage"
	"sonla/transcode/internal/transcoder"
	"sonla/transcode/internal/worker"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := config.Load()
	logger.Info("Starting PipeVid Transcode Worker Daemon...",
		zap.String("natsURL", cfg.NatsURL),
		zap.String("rawBucket", cfg.RawBucket),
		zap.String("hlsBucket", cfg.HLSBucket),
		zap.String("webhookURL", cfg.BackendWebhookURL),
	)

	s3, err := storage.NewS3Client(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize S3 storage client", zap.Error(err))
	}

	tc := transcoder.NewTranscoder(cfg.FFmpegPath, logger)

	w, err := worker.NewTranscodeWorker(cfg, s3, tc, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Transcode Worker", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := w.Start(ctx); err != nil {
		logger.Error("Transcode Worker encountered runtime error", zap.Error(err))
	}
}
