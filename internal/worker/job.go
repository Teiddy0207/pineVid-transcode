package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SchemaVersion is the message-contract version this worker speaks.
const SchemaVersion = 1

// Subjects from docs/message-contract.md §1.
const (
	SubjectTranscode    = "video.transcode"
	SubjectTranscodeDLQ = "video.transcode.dlq"
)

// TranscodeJob is the video.transcode payload (contract §2.1).
type TranscodeJob struct {
	SchemaVersion int       `json:"schema_version"`
	JobID         string    `json:"job_id"`
	VideoID       string    `json:"video_id"`
	SourceKey     string    `json:"source_key"`
	Renditions    []string  `json:"renditions"`
	CreatedAt     time.Time `json:"created_at"`
}

// DLQMessage is the video.transcode.dlq payload (contract §2.2).
type DLQMessage struct {
	SchemaVersion int       `json:"schema_version"`
	JobID         string    `json:"job_id"`
	VideoID       string    `json:"video_id"`
	Error         string    `json:"error"`
	Attempts      int       `json:"attempts"`
	FailedAt      time.Time `json:"failed_at"`
}

// DecodeJob parses and validates a job. An error means the message can never
// succeed and must be terminated, not retried.
func DecodeJob(data []byte) (TranscodeJob, error) {
	var job TranscodeJob
	if err := json.Unmarshal(data, &job); err != nil {
		return TranscodeJob{}, fmt.Errorf("invalid job payload: %w", err)
	}
	if job.SchemaVersion != SchemaVersion {
		return TranscodeJob{}, fmt.Errorf("unsupported schema_version %d", job.SchemaVersion)
	}
	if job.JobID == "" || job.VideoID == "" || job.SourceKey == "" {
		return TranscodeJob{}, errors.New("job_id, video_id and source_key are required")
	}
	return job, nil
}
