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
		zap.String("webhookURL", cfg.WebhookURL),
		zap.Strings("defaultRenditions", cfg.DefaultRenditions),
	)
	if cfg.InternalWebhookSecret == "" {
		// Every unsigned webhook would be rejected and every job would end
		// up in the DLQ — fail fast instead.
		logger.Fatal("INTERNAL_WEBHOOK_SECRET is required (HMAC key of the transcode webhook)")
	}
	if cfg.FaultExitBeforeAck {
		logger.Warn("FAULT_EXIT_BEFORE_ACK is on: the worker exits after each COMPLETED webhook (test mode)")
	}

	s3, err := storage.NewS3Client(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize S3 storage client", zap.Error(err))
	}

	tc := transcoder.NewTranscoder(cfg.FFmpegPath, cfg.FFprobePath, logger)

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
