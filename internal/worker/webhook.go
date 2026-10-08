package worker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Webhook statuses (docs/message-contract.md §4.2).
const (
	StatusProcessing = "PROCESSING"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"
)

// WebhookEvent is the body of POST /internal/webhooks/transcode.
type WebhookEvent struct {
	SchemaVersion int    `json:"schema_version"`
	JobID         string `json:"job_id"`
	VideoID       string `json:"video_id"`
	Status        string `json:"status"`
	HLSMasterURL  string `json:"hls_master_url,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

// WebhookResult is the backend's reply; Applied=false is a successful no-op.
type WebhookResult struct {
	Applied bool   `json:"applied"`
	Reason  string `json:"reason,omitempty"`
}

// Sign must stay byte-for-byte identical to backend/pkg/webhooksig.Sign.
func Sign(secret string, ts int64, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10)))
	mac.Write([]byte("."))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

type WebhookClient struct {
	url    string
	secret string
	client *http.Client
	now    func() time.Time
}

func NewWebhookClient(url, secret string) *WebhookClient {
	return &WebhookClient{url: url, secret: secret, client: &http.Client{Timeout: 10 * time.Second}, now: time.Now}
}

// Send posts a signed event with Idempotency-Key <job_id>:<status>. Any
// non-2xx reply is an error, so the caller does not ack the job.
func (c *WebhookClient) Send(ctx context.Context, ev WebhookEvent) (WebhookResult, error) {
	body, err := json.Marshal(ev)
	if err != nil {
		return WebhookResult{}, err
	}

	ts := c.now().Unix()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return WebhookResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", ev.JobID+":"+ev.Status)
	req.Header.Set("X-Timestamp", strconv.FormatInt(ts, 10))
	req.Header.Set("X-Signature", Sign(c.secret, ts, body))

	resp, err := c.client.Do(req)
	if err != nil {
		return WebhookResult{}, fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return WebhookResult{}, fmt.Errorf("webhook returned %d: %s", resp.StatusCode, bytes.TrimSpace(respBody))
	}

	var res WebhookResult
	_ = json.Unmarshal(respBody, &res)
	return res, nil
}
