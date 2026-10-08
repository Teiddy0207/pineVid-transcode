package worker

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

// Same reference value as backend/pkg/webhooksig: the two Sign copies must
// agree byte for byte. printf '1700000000.{}' | openssl dgst -sha256 -hmac secret
func TestSignMatchesBackend(t *testing.T) {
	if got := Sign("secret", 1700000000, []byte("{}")); got != "b8569b78799ff9e3cbff0fc2d63a33a2b57f3282abd07c37ae5e8e7d79a5f163" {
		t.Fatalf("signature drifted from the backend: %s", got)
	}
}

func TestWebhookSend(t *testing.T) {
	var gotKey, gotTS, gotSig string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey, gotTS, gotSig = r.Header.Get("Idempotency-Key"), r.Header.Get("X-Timestamp"), r.Header.Get("X-Signature")
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"applied":false,"reason":"duplicate"}`))
	}))
	defer srv.Close()

	c := NewWebhookClient(srv.URL, "s3cret")
	c.now = func() time.Time { return time.Unix(1_700_000_000, 0) }

	res, err := c.Send(context.Background(), WebhookEvent{SchemaVersion: 1, JobID: "job-1", VideoID: "v1", Status: StatusCompleted, HLSMasterURL: "/hls/m.m3u8"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Applied || res.Reason != "duplicate" {
		t.Errorf("reply not parsed: %+v", res)
	}
	if gotKey != "job-1:COMPLETED" {
		t.Errorf("Idempotency-Key = %q", gotKey)
	}
	ts, _ := strconv.ParseInt(gotTS, 10, 64)
	if gotSig != Sign("s3cret", ts, gotBody) {
		t.Error("signature does not cover the exact body sent")
	}
	var ev WebhookEvent
	if err := json.Unmarshal(gotBody, &ev); err != nil || ev.Status != StatusCompleted || ev.SchemaVersion != 1 {
		t.Errorf("body = %s", gotBody)
	}
}

func TestWebhookSend_Non2xxIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"invalid webhook signature"}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	if _, err := NewWebhookClient(srv.URL, "x").Send(context.Background(), WebhookEvent{JobID: "j", Status: StatusCompleted}); err == nil {
		t.Fatal("a 401 must not count as delivered (the job would be acked)")
	}
}

func TestDecodeJob(t *testing.T) {
	good := `{"schema_version":1,"job_id":"j","video_id":"v","source_key":"k","renditions":["720p","480p"],"created_at":"2026-09-25T08:00:00Z"}`
	job, err := DecodeJob([]byte(good))
	if err != nil || job.JobID != "j" || len(job.Renditions) != 2 {
		t.Fatalf("job=%+v err=%v", job, err)
	}

	for name, bad := range map[string]string{
		"old payload (no schema)": `{"video_id":"v","raw_object_key":"k","user_id":"u"}`,
		"future schema":           `{"schema_version":2,"job_id":"j","video_id":"v","source_key":"k"}`,
		"missing job_id":          `{"schema_version":1,"video_id":"v","source_key":"k"}`,
		"not json":                `garbage`,
	} {
		if _, err := DecodeJob([]byte(bad)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
