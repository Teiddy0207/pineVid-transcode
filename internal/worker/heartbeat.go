package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// HeartbeatReporter periodically self-reports this process's real CPU/RAM
// usage and current job to the backend's admin dashboard, so
// GetWorkersStatus reflects actual running workers instead of hardcoded data.
type HeartbeatReporter struct {
	workerID      string
	backendURL    string
	internalToken string
	client        *http.Client

	processing atomic.Bool
	currentJob atomic.Value // string

	mu      sync.Mutex
	lastCPU cpuSample
}

type cpuSample struct {
	ticks uint64
	at    time.Time
}

func NewHeartbeatReporter(backendBaseURL, internalToken string) *HeartbeatReporter {
	hostname, _ := os.Hostname()
	workerID := fmt.Sprintf("worker-%s-%d", hostname, os.Getpid())

	h := &HeartbeatReporter{
		workerID:      workerID,
		backendURL:    strings.TrimRight(backendBaseURL, "/") + "/v1/admin/workers/heartbeat",
		internalToken: internalToken,
		client:        &http.Client{Timeout: 5 * time.Second},
	}
	h.currentJob.Store("")
	return h
}

// SetProcessing marks this worker as busy on the given job, reflected on the
// very next heartbeat tick.
func (h *HeartbeatReporter) SetProcessing(jobID string) {
	h.processing.Store(true)
	h.currentJob.Store(jobID)
}

// SetIdle marks this worker as free again.
func (h *HeartbeatReporter) SetIdle() {
	h.processing.Store(false)
	h.currentJob.Store("")
}

// Start sends an immediate heartbeat, then keeps sending one every interval
// until ctx is done. Runs in its own goroutine; never blocks the caller.
func (h *HeartbeatReporter) Start(ctx context.Context, interval time.Duration) {
	go func() {
		h.send()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.send()
			}
		}
	}()
}

func (h *HeartbeatReporter) send() {
	status := "idle"
	if h.processing.Load() {
		status = "processing"
	}
	currentJob, _ := h.currentJob.Load().(string)

	cpuPct, ramMB := h.readSelfUsage()

	payload := map[string]any{
		"worker_id":    h.workerID,
		"status":       status,
		"current_job":  currentJob,
		"cpu_usage":    cpuPct,
		"ram_usage_mb": ramMB,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, h.backendURL, bytes.NewBuffer(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(internalTokenHeader, h.internalToken)

	resp, err := h.client.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}

// readSelfUsage returns this process's real CPU% (since the previous sample)
// and resident memory in MB, read from /proc (Linux). Returns 0 values on the
// very first call (no delta to compare against yet) or on non-Linux hosts.
func (h *HeartbeatReporter) readSelfUsage() (cpuPercent, ramMB float64) {
	ramMB = readRSSMB()

	ticks, err := readProcCPUTicks()
	if err != nil {
		return 0, ramMB
	}

	now := time.Now()
	h.mu.Lock()
	prev := h.lastCPU
	h.lastCPU = cpuSample{ticks: ticks, at: now}
	h.mu.Unlock()

	if prev.ticks == 0 || ticks < prev.ticks {
		return 0, ramMB
	}

	const clockTicksPerSecond = 100.0 // USER_HZ, standard on Linux
	elapsedSeconds := now.Sub(prev.at).Seconds()
	if elapsedSeconds <= 0 {
		return 0, ramMB
	}

	cpuSeconds := float64(ticks-prev.ticks) / clockTicksPerSecond
	cpuPercent = (cpuSeconds / elapsedSeconds) * 100.0
	return cpuPercent, ramMB
}

// readProcCPUTicks returns utime+stime (in clock ticks) from /proc/self/stat.
// The comm field (2nd field) can itself contain spaces/parens, so fields are
// parsed after the last ')' rather than by naive whitespace-splitting.
func readProcCPUTicks() (uint64, error) {
	data, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return 0, err
	}

	s := string(data)
	idx := strings.LastIndex(s, ")")
	if idx == -1 || idx+2 >= len(s) {
		return 0, fmt.Errorf("unexpected /proc/self/stat format")
	}

	fields := strings.Fields(s[idx+2:])
	// After "pid (comm) ", the next field is `state` (stat field 3); utime is
	// field 14 and stime is field 15, i.e. fields[11] and fields[12] here.
	if len(fields) < 13 {
		return 0, fmt.Errorf("unexpected /proc/self/stat field count: %d", len(fields))
	}

	utime, err := strconv.ParseUint(fields[11], 10, 64)
	if err != nil {
		return 0, err
	}
	stime, err := strconv.ParseUint(fields[12], 10, 64)
	if err != nil {
		return 0, err
	}

	return utime + stime, nil
}

// readRSSMB returns this process's resident set size in MB from /proc/self/status.
func readRSSMB() float64 {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0
	}

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0
		}
		kb, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return 0
		}
		return kb / 1024.0
	}
	return 0
}
