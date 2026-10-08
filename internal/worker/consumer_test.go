package worker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sonla/transcode/internal/config"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// These tests exercise the contract rules (docs/message-contract.md §3) on a
// real JetStream server they may freely modify:
//
//	docker run --rm -d -p 54222:4222 nats:2.14-alpine -js
//	TEST_NATS_URL=nats://localhost:54222 go test ./internal/worker/
//
// They run sequentially: every test recreates the VIDEO_JOBS stream.

type pipelineFunc func(ctx context.Context, job TranscodeJob) (string, error)

func (f pipelineFunc) Run(ctx context.Context, job TranscodeJob) (string, error) { return f(ctx, job) }

// fakeBackend records webhooks and answers with a scripted status code.
type fakeBackend struct {
	mu     sync.Mutex
	events []WebhookEvent
	keys   []string
	// fail[i] = true makes the i-th COMPLETED webhook answer 500.
	fail map[int]bool
	done int
	srv  *httptest.Server
}

func newFakeBackend(t *testing.T, secret string) *fakeBackend {
	b := &fakeBackend{fail: map[int]bool{}}
	b.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		ts, _ := strconv.ParseInt(r.Header.Get("X-Timestamp"), 10, 64)
		if r.Header.Get("X-Signature") != Sign(secret, ts, body) {
			http.Error(w, "bad signature", http.StatusUnauthorized)
			return
		}
		var ev WebhookEvent
		_ = json.Unmarshal(body, &ev)

		b.mu.Lock()
		defer b.mu.Unlock()
		b.events = append(b.events, ev)
		b.keys = append(b.keys, r.Header.Get("Idempotency-Key"))
		if ev.Status == StatusCompleted {
			b.done++
			if b.fail[b.done] {
				http.Error(w, "boom", http.StatusInternalServerError)
				return
			}
		}
		_, _ = w.Write([]byte(`{"applied":true}`))
	}))
	t.Cleanup(b.srv.Close)
	return b
}

func (b *fakeBackend) count(status string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := 0
	for _, ev := range b.events {
		if ev.Status == status {
			n++
		}
	}
	return n
}

type harness struct {
	t      *testing.T
	cfg    *config.Config
	js     nats.JetStreamContext
	worker *TranscodeWorker
	stop   func()
}

func setup(t *testing.T, p Pipeline, backend *fakeBackend, tweak func(*config.Config)) *harness {
	t.Helper()
	url := os.Getenv("TEST_NATS_URL")
	if url == "" {
		t.Skip("TEST_NATS_URL not set; skipping NATS-backed test")
	}

	cfg := &config.Config{
		NatsURL: url, StreamName: "VIDEO_JOBS", ConsumerName: "transcode-worker-test",
		AckWait: 2 * time.Second, InProgressInterval: 500 * time.Millisecond,
		MaxDeliver: 3, MaxAckPending: 1,
		NakDelays:             []time.Duration{100 * time.Millisecond, 100 * time.Millisecond},
		InternalWebhookSecret: "s3cret",
	}
	if tweak != nil {
		tweak(cfg)
	}

	admin, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	js, _ := admin.JetStream()
	_ = js.DeleteStream("VIDEO_JOBS")
	_ = js.DeleteStream("PIPEVID_STREAM")
	if _, err := js.AddStream(&nats.StreamConfig{Name: "VIDEO_JOBS", Subjects: []string{SubjectTranscode, SubjectTranscodeDLQ}, Duplicates: 2 * time.Minute}); err != nil {
		t.Fatal(err)
	}

	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	w, err := newWorker(cfg, nc, p, NewWebhookClient(backend.srv.URL, cfg.InternalWebhookSecret), zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := w.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("worker: %v", err)
		}
	}()
	h := &harness{t: t, cfg: cfg, js: js, worker: w, stop: func() { cancel(); <-done }}
	t.Cleanup(h.stop)
	return h
}

func (h *harness) publish(jobID string) {
	h.t.Helper()
	data, _ := json.Marshal(TranscodeJob{SchemaVersion: 1, JobID: jobID, VideoID: "video-" + jobID, SourceKey: "raw-uploads/x/raw.mp4", Renditions: []string{"720p", "480p"}, CreatedAt: time.Now().UTC()})
	if _, err := h.js.Publish(SubjectTranscode, data, nats.MsgId(jobID)); err != nil {
		h.t.Fatal(err)
	}
}

// settled waits until the consumer has nothing pending or in flight.
func (h *harness) settled(timeout time.Duration) *nats.ConsumerInfo {
	h.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		info, err := h.js.ConsumerInfo(h.cfg.StreamName, h.cfg.ConsumerName)
		if err == nil && info.NumPending == 0 && info.NumAckPending == 0 && info.NumRedelivered == 0 && info.Delivered.Consumer > 0 {
			return info
		}
		time.Sleep(100 * time.Millisecond)
	}
	info, _ := h.js.ConsumerInfo(h.cfg.StreamName, h.cfg.ConsumerName)
	h.t.Fatalf("consumer did not settle: %+v", info)
	return nil
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool, what string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestWorker_ConsumerMatchesContract(t *testing.T) {
	backend := newFakeBackend(t, "s3cret")
	h := setup(t, pipelineFunc(func(context.Context, TranscodeJob) (string, error) { return "/m", nil }), backend, func(c *config.Config) {
		c.AckWait = 60 * time.Second
	})

	var info *nats.ConsumerInfo
	waitFor(t, 5*time.Second, func() bool {
		var err error
		info, err = h.js.ConsumerInfo("VIDEO_JOBS", h.cfg.ConsumerName)
		return err == nil
	}, "consumer creation")

	if info.Config.AckWait != 60*time.Second || info.Config.MaxDeliver != 3 || info.Config.MaxAckPending != 1 ||
		info.Config.AckPolicy != nats.AckExplicitPolicy || info.Config.DeliverSubject != "" || info.Config.FilterSubject != SubjectTranscode {
		t.Fatalf("consumer config does not match the contract: %+v", info.Config)
	}
}

func TestWorker_SuccessSendsSignedWebhooksThenAcks(t *testing.T) {
	backend := newFakeBackend(t, "s3cret")
	var runs atomic.Int32
	h := setup(t, pipelineFunc(func(_ context.Context, job TranscodeJob) (string, error) {
		runs.Add(1)
		return "/hls-streams/videos/" + job.VideoID + "/master.m3u8", nil
	}), backend, nil)

	h.publish("job-ok")
	h.settled(10 * time.Second)

	if runs.Load() != 1 {
		t.Fatalf("pipeline ran %d times", runs.Load())
	}
	backend.mu.Lock()
	defer backend.mu.Unlock()
	if len(backend.keys) != 2 || backend.keys[0] != "job-ok:PROCESSING" || backend.keys[1] != "job-ok:COMPLETED" {
		t.Fatalf("webhooks = %v", backend.keys)
	}
	if backend.events[1].HLSMasterURL != "/hls-streams/videos/video-job-ok/master.m3u8" {
		t.Errorf("COMPLETED carries %q", backend.events[1].HLSMasterURL)
	}
}

// A job longer than AckWait is not redelivered thanks to in-progress acks.
func TestWorker_InProgressPreventsRedelivery(t *testing.T) {
	backend := newFakeBackend(t, "s3cret")
	var runs atomic.Int32
	h := setup(t, pipelineFunc(func(ctx context.Context, _ TranscodeJob) (string, error) {
		runs.Add(1)
		select {
		case <-time.After(5 * time.Second): // 2.5 × AckWait
		case <-ctx.Done():
		}
		return "/m", nil
	}), backend, nil)

	h.publish("job-slow")
	h.settled(15 * time.Second)

	if runs.Load() != 1 {
		t.Fatalf("a slow job was redelivered: pipeline ran %d times", runs.Load())
	}
}

// Failures are retried up to MaxDeliver, then the job is dead-lettered.
func TestWorker_RetriesThenDeadLetters(t *testing.T) {
	backend := newFakeBackend(t, "s3cret")
	var runs atomic.Int32
	h := setup(t, pipelineFunc(func(context.Context, TranscodeJob) (string, error) {
		runs.Add(1)
		return "", errors.New("ffmpeg exited 1")
	}), backend, nil)

	dlq := make(chan DLQMessage, 1)
	sub, err := h.js.Subscribe(SubjectTranscodeDLQ, func(m *nats.Msg) {
		var d DLQMessage
		_ = json.Unmarshal(m.Data, &d)
		dlq <- d
	}, nats.DeliverAll(), nats.AckNone())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()

	h.publish("job-bad")

	select {
	case d := <-dlq:
		if d.JobID != "job-bad" || d.Attempts != 3 || d.Error == "" || d.SchemaVersion != 1 {
			t.Fatalf("DLQ message = %+v", d)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("job never reached the DLQ")
	}
	if runs.Load() != 3 {
		t.Fatalf("pipeline ran %d times, want MaxDeliver = 3", runs.Load())
	}
	if backend.count(StatusCompleted) != 0 {
		t.Fatal("a failed job must not report COMPLETED")
	}
}

// The job is acked only after the COMPLETED webhook returns 2xx.
func TestWorker_WebhookFailureIsRetried(t *testing.T) {
	backend := newFakeBackend(t, "s3cret")
	backend.fail[1] = true // first COMPLETED → 500
	var runs atomic.Int32
	h := setup(t, pipelineFunc(func(context.Context, TranscodeJob) (string, error) {
		runs.Add(1)
		return "/m", nil
	}), backend, nil)

	h.publish("job-hook")
	waitFor(t, 10*time.Second, func() bool { return backend.count(StatusCompleted) == 2 }, "second COMPLETED")
	waitFor(t, 5*time.Second, func() bool {
		info, err := h.js.ConsumerInfo(h.cfg.StreamName, h.cfg.ConsumerName)
		return err == nil && info.NumAckPending == 0 && info.NumPending == 0
	}, "ack after the successful webhook")

	if runs.Load() != 2 {
		t.Fatalf("pipeline ran %d times, want 2", runs.Load())
	}
	backend.mu.Lock()
	defer backend.mu.Unlock()
	for _, k := range backend.keys {
		if k != "job-hook:PROCESSING" && k != "job-hook:COMPLETED" {
			t.Fatalf("retries must reuse the same idempotency keys, got %v", backend.keys)
		}
	}
}

// Scenario 1 of the proposal (§7.5), in miniature: the worker dies after a
// successful COMPLETED webhook but before acking; NATS redelivers after
// AckWait and the same idempotency key is sent again.
func TestWorker_CrashBeforeAckIsRedelivered(t *testing.T) {
	backend := newFakeBackend(t, "s3cret")
	var runs atomic.Int32
	h := setup(t, pipelineFunc(func(context.Context, TranscodeJob) (string, error) {
		runs.Add(1)
		return "/m", nil
	}), backend, func(c *config.Config) { c.FaultExitBeforeAck = true })

	exits := make(chan int, 1)
	h.worker.exit = func(code int) {
		h.cfg.FaultExitBeforeAck = false // "restart" without the fault flag
		exits <- code
	}

	h.publish("job-crash")
	select {
	case code := <-exits:
		if code != 3 {
			t.Fatalf("exit code %d", code)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("worker did not hit the fault point")
	}

	waitFor(t, 10*time.Second, func() bool { return backend.count(StatusCompleted) == 2 }, "redelivered COMPLETED")
	if runs.Load() != 2 {
		t.Fatalf("pipeline ran %d times", runs.Load())
	}
	backend.mu.Lock()
	defer backend.mu.Unlock()
	if backend.keys[1] != "job-crash:COMPLETED" || backend.keys[len(backend.keys)-1] != "job-crash:COMPLETED" {
		t.Fatalf("keys = %v", backend.keys)
	}
}

func TestWorker_InvalidPayloadIsTerminated(t *testing.T) {
	backend := newFakeBackend(t, "s3cret")
	var runs atomic.Int32
	h := setup(t, pipelineFunc(func(context.Context, TranscodeJob) (string, error) {
		runs.Add(1)
		return "/m", nil
	}), backend, nil)

	if _, err := h.js.Publish(SubjectTranscode, []byte(`{"video_id":"v","raw_object_key":"k"}`)); err != nil {
		t.Fatal(err)
	}
	h.settled(10 * time.Second)
	time.Sleep(3 * time.Second) // longer than AckWait: no redelivery

	if runs.Load() != 0 || len(backend.events) != 0 {
		t.Fatalf("an invalid job was processed (runs=%d, webhooks=%d)", runs.Load(), len(backend.events))
	}
}

func TestNakDelay(t *testing.T) {
	w := &TranscodeWorker{cfg: &config.Config{NakDelays: []time.Duration{10 * time.Second, 30 * time.Second}}}
	for attempt, want := range map[int]time.Duration{1: 10 * time.Second, 2: 30 * time.Second, 5: 30 * time.Second} {
		if got := w.nakDelay(attempt); got != want {
			t.Errorf("attempt %d: %v, want %v", attempt, got, want)
		}
	}
}
