package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sonla/transcode/internal/config"
	"sonla/transcode/internal/storage"
	"sonla/transcode/internal/transcoder"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// jobTimeout bounds one transcode attempt; in-progress acks keep NATS from
// redelivering it meanwhile.
const jobTimeout = 30 * time.Minute

// Pipeline does the actual work for one job and returns the master playlist
// URL. It is an interface so the NATS handling can be tested on its own.
type Pipeline interface {
	Run(ctx context.Context, job TranscodeJob) (string, error)
}

// s3Pipeline: download from MinIO → FFmpeg HLS ladder → upload to MinIO.
type s3Pipeline struct {
	cfg *config.Config
	s3  *storage.S3Client
	tc  *transcoder.Transcoder
}

func (p *s3Pipeline) Run(ctx context.Context, job TranscodeJob) (string, error) {
	workDir := filepath.Join(p.cfg.TempDir, job.JobID)
	defer os.RemoveAll(workDir)

	rawPath := filepath.Join(workDir, "raw"+filepath.Ext(job.SourceKey))
	hlsDir := filepath.Join(workDir, "hls")

	if err := p.s3.DownloadRawVideo(ctx, job.SourceKey, rawPath); err != nil {
		return "", fmt.Errorf("download: %w", err)
	}

	renditions := job.Renditions
	if len(renditions) == 0 {
		renditions = p.cfg.DefaultRenditions
	}
	if _, err := p.tc.TranscodeHLS(ctx, rawPath, hlsDir, renditions); err != nil {
		return "", fmt.Errorf("transcode: %w", err)
	}

	// Output lives at a path fixed by video_id, so a re-transcode overwrites
	// the same objects instead of leaving orphans.
	if err := p.s3.UploadHLSFiles(ctx, hlsDir, "videos/"+job.VideoID); err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}

	return fmt.Sprintf("/%s/videos/%s/master.m3u8", p.cfg.HLSBucket, job.VideoID), nil
}

type TranscodeWorker struct {
	cfg       *config.Config
	logger    *zap.Logger
	nc        *nats.Conn
	js        nats.JetStreamContext
	pipeline  Pipeline
	webhook   *WebhookClient
	heartbeat *HeartbeatReporter
	// exit is os.Exit; tests replace it to simulate a crash.
	exit func(code int)
}

func NewTranscodeWorker(cfg *config.Config, s3 *storage.S3Client, tc *transcoder.Transcoder, logger *zap.Logger) (*TranscodeWorker, error) {
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server at %s: %w", cfg.NatsURL, err)
	}

	w, err := newWorker(cfg, nc, &s3Pipeline{cfg: cfg, s3: s3, tc: tc}, NewWebhookClient(cfg.WebhookURL, cfg.InternalWebhookSecret), logger)
	if err != nil {
		nc.Close()
		return nil, err
	}
	if cfg.BackendBaseURL != "" {
		w.heartbeat = NewHeartbeatReporter(cfg.BackendBaseURL, cfg.InternalWebhookSecret)
	}
	return w, nil
}

func newWorker(cfg *config.Config, nc *nats.Conn, pipeline Pipeline, webhook *WebhookClient, logger *zap.Logger) (*TranscodeWorker, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}
	return &TranscodeWorker{cfg: cfg, logger: logger, nc: nc, js: js, pipeline: pipeline, webhook: webhook, exit: os.Exit}, nil
}

// Start consumes jobs one at a time until ctx is cancelled.
func (w *TranscodeWorker) Start(ctx context.Context) error {
	defer w.nc.Close()

	if w.heartbeat != nil {
		w.heartbeat.Start(ctx, 10*time.Second)
	}

	if err := w.waitForStream(ctx); err != nil {
		return err
	}
	if err := w.ensureConsumer(); err != nil {
		return err
	}

	sub, err := w.js.PullSubscribe(SubjectTranscode, w.cfg.ConsumerName, nats.Bind(w.cfg.StreamName, w.cfg.ConsumerName))
	if err != nil {
		return fmt.Errorf("failed to bind pull consumer %s: %w", w.cfg.ConsumerName, err)
	}

	w.logger.Info("Transcode Worker started",
		zap.String("stream", w.cfg.StreamName),
		zap.String("consumer", w.cfg.ConsumerName),
		zap.Duration("ackWait", w.cfg.AckWait),
		zap.Int("maxDeliver", w.cfg.MaxDeliver),
	)

	for ctx.Err() == nil {
		fetchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		msgs, err := sub.Fetch(1, nats.Context(fetchCtx))
		cancel()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || errors.Is(err, nats.ErrTimeout) {
				continue
			}
			w.logger.Warn("Fetch failed", zap.Error(err))
			sleep(ctx, time.Second)
			continue
		}
		for _, msg := range msgs {
			w.processMessage(ctx, msg)
		}
	}

	w.logger.Info("Shutting down Transcode Worker")
	return nil
}

// waitForStream blocks until the backend has created the stream (it owns the
// stream config; the worker only binds to it).
func (w *TranscodeWorker) waitForStream(ctx context.Context) error {
	for {
		_, err := w.js.StreamInfo(w.cfg.StreamName)
		if err == nil {
			return nil
		}
		w.logger.Warn("Waiting for NATS stream (created by the backend)", zap.String("stream", w.cfg.StreamName), zap.Error(err))
		if !sleep(ctx, 2*time.Second) {
			return ctx.Err()
		}
	}
}

// ensureConsumer creates the durable pull consumer, or updates it to the
// contract values if it already exists.
func (w *TranscodeWorker) ensureConsumer() error {
	cc := &nats.ConsumerConfig{
		Durable:       w.cfg.ConsumerName,
		AckPolicy:     nats.AckExplicitPolicy,
		AckWait:       w.cfg.AckWait,
		MaxDeliver:    w.cfg.MaxDeliver,
		MaxAckPending: w.cfg.MaxAckPending,
		FilterSubject: SubjectTranscode,
		DeliverPolicy: nats.DeliverAllPolicy,
	}

	_, err := w.js.ConsumerInfo(w.cfg.StreamName, w.cfg.ConsumerName)
	switch {
	case errors.Is(err, nats.ErrConsumerNotFound):
		_, err = w.js.AddConsumer(w.cfg.StreamName, cc)
	case err == nil:
		_, err = w.js.UpdateConsumer(w.cfg.StreamName, cc)
	}
	if err != nil {
		return fmt.Errorf("failed to set up consumer %s: %w", w.cfg.ConsumerName, err)
	}
	return nil
}

func (w *TranscodeWorker) processMessage(ctx context.Context, msg *nats.Msg) {
	job, err := DecodeJob(msg.Data)
	if err != nil {
		w.logger.Error("Dropping invalid transcode job", zap.Error(err), zap.ByteString("data", msg.Data))
		_ = msg.Term()
		return
	}

	attempt := 1
	if meta, err := msg.Metadata(); err == nil {
		attempt = int(meta.NumDelivered)
	}
	final := attempt >= w.cfg.MaxDeliver
	log := w.logger.With(zap.String("jobID", job.JobID), zap.String("videoID", job.VideoID), zap.Int("attempt", attempt))
	log.Info("Received transcode job")

	if w.heartbeat != nil {
		w.heartbeat.SetProcessing(job.VideoID)
		defer w.heartbeat.SetIdle()
	}

	stop := w.keepInProgress(msg)
	defer stop()

	jobCtx, cancel := context.WithTimeout(ctx, jobTimeout)
	defer cancel()

	// PROCESSING only feeds the started_at timestamp; a failure here must not
	// fail the job.
	if _, err := w.webhook.Send(jobCtx, w.event(job, StatusProcessing, "", "")); err != nil {
		log.Warn("PROCESSING webhook failed", zap.Error(err))
	}

	masterURL, err := w.pipeline.Run(jobCtx, job)
	if err != nil {
		w.handleFailure(ctx, msg, job, attempt, final, err, log)
		return
	}

	res, err := w.webhook.Send(jobCtx, w.event(job, StatusCompleted, masterURL, ""))
	if err != nil {
		// Not acked: NATS redelivers, and the backend dedups the repeat.
		w.handleFailure(ctx, msg, job, attempt, final, fmt.Errorf("COMPLETED webhook: %w", err), log)
		return
	}
	if !res.Applied {
		log.Info("COMPLETED webhook was a no-op", zap.String("reason", res.Reason))
	}

	if w.cfg.FaultExitBeforeAck {
		log.Warn("FAULT_EXIT_BEFORE_ACK: exiting after the webhook, before the ack")
		w.exit(3)
		return
	}

	if err := msg.AckSync(); err != nil {
		log.Warn("Ack failed; the job will be redelivered and deduplicated by the backend", zap.Error(err))
		return
	}
	log.Info("Transcode job completed", zap.String("hlsMasterURL", masterURL))
}

// handleFailure retries with backoff, or dead-letters on the last attempt.
func (w *TranscodeWorker) handleFailure(ctx context.Context, msg *nats.Msg, job TranscodeJob, attempt int, final bool, cause error, log *zap.Logger) {
	if ctx.Err() != nil {
		// Shutting down: not the job's fault, hand it back right away.
		_ = msg.Nak()
		return
	}

	if !final {
		delay := w.nakDelay(attempt)
		log.Error("Transcode attempt failed, will retry", zap.Error(cause), zap.Duration("retryIn", delay))
		_ = msg.NakWithDelay(delay)
		return
	}

	log.Error("Transcode job failed on its last attempt, dead-lettering", zap.Error(cause))
	if err := w.publishDLQ(job, attempt, cause); err != nil {
		log.Error("Could not publish to the DLQ; job is lost", zap.Error(err))
		_ = msg.Term()
		return
	}
	_ = msg.AckSync()
}

func (w *TranscodeWorker) event(job TranscodeJob, status, masterURL, errMsg string) WebhookEvent {
	return WebhookEvent{
		SchemaVersion: SchemaVersion,
		JobID:         job.JobID,
		VideoID:       job.VideoID,
		Status:        status,
		HLSMasterURL:  masterURL,
		ErrorMessage:  errMsg,
	}
}

// nakDelay is the wait before redelivery after failed attempt n (1-based).
func (w *TranscodeWorker) nakDelay(attempt int) time.Duration {
	if len(w.cfg.NakDelays) == 0 {
		return 0
	}
	i := attempt - 1
	if i >= len(w.cfg.NakDelays) {
		i = len(w.cfg.NakDelays) - 1
	}
	if i < 0 {
		i = 0
	}
	return w.cfg.NakDelays[i]
}

func (w *TranscodeWorker) publishDLQ(job TranscodeJob, attempts int, cause error) error {
	data, err := json.Marshal(DLQMessage{
		SchemaVersion: SchemaVersion,
		JobID:         job.JobID,
		VideoID:       job.VideoID,
		Error:         cause.Error(),
		Attempts:      attempts,
		FailedAt:      time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	for try := 1; ; try++ {
		_, err = w.js.Publish(SubjectTranscodeDLQ, data, nats.MsgId("dlq-"+job.JobID))
		if err == nil || try == 3 {
			return err
		}
		time.Sleep(time.Duration(try) * time.Second)
	}
}

// keepInProgress extends the ack deadline every InProgressInterval while a
// job runs, so a long transcode is not redelivered to another worker.
func (w *TranscodeWorker) keepInProgress(msg *nats.Msg) (stop func()) {
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(w.cfg.InProgressInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := msg.InProgress(); err != nil {
					w.logger.Warn("InProgress ack failed", zap.Error(err))
				}
			}
		}
	}()
	return func() {
		close(done)
		wg.Wait()
	}
}

// sleep waits d or until ctx is done; false means ctx is done.
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
