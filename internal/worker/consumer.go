package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sonla/transcode/internal/config"
	"sonla/transcode/internal/storage"
	"sonla/transcode/internal/transcoder"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// maxDeliverAttempts must match the nats.MaxDeliver value passed to QueueSubscribe below.
const maxDeliverAttempts = 3

type TranscodeJobPayload struct {
	VideoID      string `json:"video_id"`
	RawObjectKey string `json:"raw_object_key"`
	UserID       string `json:"user_id"`
}

type WebhookPayload struct {
	VideoID      string `json:"video_id"`
	Status       string `json:"status"` // "complete" or "failed"
	HLSMasterURL string `json:"hls_master_url,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type TranscodeWorker struct {
	cfg        *config.Config
	s3         *storage.S3Client
	transcoder *transcoder.Transcoder
	logger     *zap.Logger
	nc         *nats.Conn
	js         nats.JetStreamContext
	heartbeat  *HeartbeatReporter
}

func NewTranscodeWorker(cfg *config.Config, s3 *storage.S3Client, tc *transcoder.Transcoder, logger *zap.Logger) (*TranscodeWorker, error) {
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server at %s: %w", cfg.NatsURL, err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	return &TranscodeWorker{
		cfg:        cfg,
		s3:         s3,
		transcoder: tc,
		logger:     logger,
		nc:         nc,
		js:         js,
		heartbeat:  NewHeartbeatReporter(cfg.BackendBaseURL),
	}, nil
}

func (w *TranscodeWorker) Start(ctx context.Context) error {
	w.heartbeat.Start(ctx, 10*time.Second)

	// Ensure NATS stream exists
	streamName := w.cfg.StreamName
	_, err := w.js.StreamInfo(streamName)
	if err != nil {
		_, err = w.js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{"video.transcode"},
		})
		if err != nil {
			w.logger.Warn("Failed to create NATS stream (might already exist)", zap.Error(err))
		}
	}

	sub, err := w.js.QueueSubscribe(
		"video.transcode",
		w.cfg.ConsumerGroup,
		func(msg *nats.Msg) {
			w.processMessage(ctx, msg)
		},
		nats.ManualAck(),
		nats.AckWait(15*time.Minute),
		nats.MaxDeliver(maxDeliverAttempts),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to NATS queue video.transcode: %w", err)
	}

	w.logger.Info("Transcode Worker Daemon started listening for video.transcode jobs",
		zap.String("natsURL", w.cfg.NatsURL),
		zap.String("stream", streamName),
		zap.String("consumerGroup", w.cfg.ConsumerGroup),
	)

	<-ctx.Done()
	w.logger.Info("Shutting down Transcode Worker Daemon...")
	_ = sub.Unsubscribe()
	w.nc.Close()
	return nil
}

func (w *TranscodeWorker) processMessage(parentCtx context.Context, msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(parentCtx, 30*time.Minute)
	defer cancel()

	var payload TranscodeJobPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		w.logger.Error("Invalid NATS job payload format", zap.Error(err), zap.String("data", string(msg.Data)))
		_ = msg.Term()
		return
	}

	w.heartbeat.SetProcessing(payload.VideoID)
	defer w.heartbeat.SetIdle()

	// Determine whether NATS will still redeliver this message if we Nak it.
	// Only report a permanent "failed" status to the backend (and Ack, stopping
	// redelivery) once this was the last delivery attempt — otherwise Nak so
	// JetStream actually retries, instead of every failure being final.
	isFinalAttempt := true
	if meta, metaErr := msg.Metadata(); metaErr == nil {
		isFinalAttempt = meta.NumDelivered >= maxDeliverAttempts
	} else {
		w.logger.Warn("Failed to read NATS message metadata, treating as final attempt", zap.Error(metaErr))
	}

	w.logger.Info("Received video transcode job",
		zap.String("videoID", payload.VideoID),
		zap.String("rawKey", payload.RawObjectKey),
		zap.Bool("isFinalAttempt", isFinalAttempt),
	)

	handleRetryableFailure := func(stage string, err error) {
		w.logger.Error("Transcode job stage failed", zap.String("stage", stage), zap.Error(err), zap.String("videoID", payload.VideoID))
		if isFinalAttempt {
			w.notifyBackendFailure(payload.VideoID, err.Error())
			_ = msg.Ack()
			return
		}
		// Leave the message unacked/nak'd so NATS JetStream redelivers it.
		_ = msg.Nak()
	}

	// Prepare local temp directories
	workDir := filepath.Join(w.cfg.TempDir, payload.VideoID)
	rawFilePath := filepath.Join(workDir, "raw.mp4")
	hlsOutDir := filepath.Join(workDir, "hls")

	defer func() {
		_ = os.RemoveAll(workDir)
	}()

	// 1. Download raw video from S3
	w.logger.Info("Downloading raw video from S3", zap.String("videoID", payload.VideoID))
	if err := w.s3.DownloadRawVideo(ctx, payload.RawObjectKey, rawFilePath); err != nil {
		handleRetryableFailure("download", err)
		return
	}

	// 2. Run FFmpeg HLS Transcoding
	w.logger.Info("Transcoding video to HLS ladder", zap.String("videoID", payload.VideoID))
	if err := w.transcoder.TranscodeHLS(ctx, rawFilePath, hlsOutDir); err != nil {
		handleRetryableFailure("transcode", err)
		return
	}

	// 3. Upload HLS files to S3 bucket hls-streams
	s3Prefix := fmt.Sprintf("videos/%s", payload.VideoID)
	w.logger.Info("Uploading HLS playlist and segments to S3", zap.String("videoID", payload.VideoID), zap.String("s3Prefix", s3Prefix))
	if err := w.s3.UploadHLSFiles(ctx, hlsOutDir, s3Prefix); err != nil {
		handleRetryableFailure("upload", err)
		return
	}

	// 4. Send Webhook callback to Main Backend
	hlsMasterURL := fmt.Sprintf("/%s/videos/%s/master.m3u8", w.cfg.HLSBucket, payload.VideoID)
	w.logger.Info("Notifying Main Backend of completed transcoding",
		zap.String("videoID", payload.VideoID),
		zap.String("hlsMasterURL", hlsMasterURL),
	)
	if err := w.notifyBackendSuccess(payload.VideoID, hlsMasterURL); err != nil {
		w.logger.Warn("Failed to send webhook callback to backend (will still ack NATS)", zap.Error(err))
	}

	// 5. Acknowledge NATS message
	_ = msg.Ack()
	w.logger.Info("Video transcode job completed successfully", zap.String("videoID", payload.VideoID))
}

func (w *TranscodeWorker) notifyBackendSuccess(videoID, hlsMasterURL string) error {
	payload := WebhookPayload{
		VideoID:      videoID,
		Status:       "complete",
		HLSMasterURL: hlsMasterURL,
	}
	return w.sendWebhook(payload)
}

func (w *TranscodeWorker) notifyBackendFailure(videoID, errorMsg string) {
	payload := WebhookPayload{
		VideoID:      videoID,
		Status:       "failed",
		ErrorMessage: errorMsg,
	}
	_ = w.sendWebhook(payload)
}

func (w *TranscodeWorker) sendWebhook(payload WebhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", w.cfg.BackendWebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
