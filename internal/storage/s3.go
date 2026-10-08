package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sonla/transcode/internal/config"
	"sort"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	minio *minio.Client
	cfg   *config.Config
}

func NewS3Client(cfg *config.Config) (*S3Client, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MinIO S3: %w", err)
	}

	return &S3Client{
		minio: client,
		cfg:   cfg,
	}, nil
}

// DownloadRawVideo downloads a raw video file from S3 raw-videos bucket to destPath
func (s *S3Client) DownloadRawVideo(ctx context.Context, objectKey, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Ensure raw bucket exists
	exists, err := s.minio.BucketExists(ctx, s.cfg.RawBucket)
	if err != nil {
		return fmt.Errorf("failed to check raw bucket: %w", err)
	}
	if !exists {
		if err := s.minio.MakeBucket(ctx, s.cfg.RawBucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create raw bucket: %w", err)
		}
	}

	// Poll briefly for the raw object to become visible (upload from the
	// frontend and this worker picking up the NATS job can race slightly),
	// but a genuinely missing/unreadable object must be a real error —
	// silently substituting a placeholder video here would make the worker
	// report "complete" for uploads that never actually made it to S3.
	var lastStatErr error
	var found bool
	for attempt := 1; attempt <= 5; attempt++ {
		stat, err := s.minio.StatObject(ctx, s.cfg.RawBucket, objectKey, minio.StatObjectOptions{})
		if err == nil && stat.Size > 0 {
			found = true
			break
		}
		lastStatErr = err
		time.Sleep(1 * time.Second)
	}
	if !found {
		if lastStatErr != nil {
			return fmt.Errorf("raw object %q not found in bucket %q: %w", objectKey, s.cfg.RawBucket, lastStatErr)
		}
		return fmt.Errorf("raw object %q not found in bucket %q (empty file)", objectKey, s.cfg.RawBucket)
	}

	obj, err := s.minio.GetObject(ctx, s.cfg.RawBucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get raw object %q from S3: %w", objectKey, err)
	}
	defer obj.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, obj); err != nil {
		return fmt.Errorf("failed to copy raw object %q to local disk: %w", objectKey, err)
	}

	return nil
}

// UploadHLSFiles recursively uploads all .m3u8 and .ts files from localDir to MinIO HLSBucket under prefix
func (s *S3Client) UploadHLSFiles(ctx context.Context, localDir, prefix string) error {
	// Ensure HLS bucket exists
	exists, err := s.minio.BucketExists(ctx, s.cfg.HLSBucket)
	if err != nil {
		return fmt.Errorf("failed to check HLS bucket: %w", err)
	}
	if !exists {
		if err := s.minio.MakeBucket(ctx, s.cfg.HLSBucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("failed to create HLS bucket: %w", err)
		}
	}

	var files []string
	if err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return err
	}

	// master.m3u8 goes last: until it exists the video is not playable, so a
	// half-finished upload never exposes a playlist pointing at missing
	// segments.
	sort.SliceStable(files, func(i, j int) bool {
		return filepath.Base(files[i]) != "master.m3u8" && filepath.Base(files[j]) == "master.m3u8"
	})

	for _, path := range files {
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		s3Key := fmt.Sprintf("%s/%s", strings.TrimSuffix(prefix, "/"), relPath)

		contentType := "application/octet-stream"
		if strings.HasSuffix(path, ".m3u8") {
			contentType = "application/x-mpegURL"
		} else if strings.HasSuffix(path, ".ts") {
			contentType = "video/MP2T"
		}

		_, err = s.minio.FPutObject(ctx, s.cfg.HLSBucket, s3Key, path, minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			return fmt.Errorf("failed to upload %s to S3: %w", path, err)
		}
	}

	return nil
}
