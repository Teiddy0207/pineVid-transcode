package job

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sonla/transcode/model"
	"sonla/transcode/util"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// safeJoin joins base and rel, rejecting any rel path (e.g. containing "..",
// or absolute) that would resolve outside of base. job.Source/job.Destination
// come straight from the API request, so this is the only thing standing
// between a crafted path and reading/writing files outside the intended dirs.
func safeJoin(base, rel string) (string, error) {
	cleanBase := filepath.Clean(base)
	full := filepath.Join(cleanBase, filepath.Clean(string(os.PathSeparator)+rel))
	if full != cleanBase && !strings.HasPrefix(full, cleanBase+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q escapes base directory %q", rel, base)
	}
	return full, nil
}

type Rendition struct {
	Width        int64
	Height       int64
	VideoBitrate int64
	AudioBitrate int64
}

type TranscodeCallback struct {
	Error     string `json:"error" form:"error"`
	RequestID string `json:"requestID" form:"requestID"`
	Status    string `json:"status" form:"status"`
}

const segmentTargetDuration = 4
const maxBitrateRatio = 1.07
const rateMonitorBufferRatio = 1.5

var renditions = [4]Rendition{
	Rendition{
		Width:        842,
		Height:       480,
		VideoBitrate: 1400,
		AudioBitrate: 92,
	},
	Rendition{
		Width:        1280,
		Height:       720,
		VideoBitrate: 2800,
		AudioBitrate: 128,
	},
	Rendition{
		Width:        1920,
		Height:       1080,
		VideoBitrate: 4000,
		AudioBitrate: 192,
	},
}

func Transcode(db *gorm.DB, job model.Job) {
	var sourceFolder = os.Getenv("SOURCE_FOLDER")
	var destinationFolder = os.Getenv("DESTINATION_FOLDER")
	var dir, _ = os.Getwd()

	/* Check app */
	if job.App == nil {
		util.Logger.Debug("Transcode app not found")
		job.Status = model.JOB_ERROR
		job.Error = "Ứng dụng không tồn tại!"
		if err := db.Omit("App").Save(&job).Error; err != nil {
			util.Logger.Error("Transcode db.save error", zap.String("error", err.Error()))
		}
		return
	}

	/* check source folder and destination folder */
	if sourceFolder == "" {
		sourceFolder = dir
	}
	if destinationFolder == "" {
		destinationFolder = dir
	}

	/* Check has input file */
	sourceFilePath, err := safeJoin(sourceFolder, job.Source)
	if err != nil {
		util.Logger.Debug("Transcode invalid source path", zap.String("source", job.Source), zap.String("error", err.Error()))
		job.Status = model.JOB_ERROR
		job.Error = "Đường dẫn tập tin không hợp lệ!"
		if err := db.Omit("App").Save(&job).Error; err != nil {
			util.Logger.Error("Transcode db.save error", zap.String("error", err.Error()))
		}
		Callback(TranscodeCallback{
			RequestID: job.RequestID,
			Status:    model.JOB_ERROR,
			Error:     "Đường dẫn tập tin không hợp lệ!",
		}, job.App.Key, job.Callback)
		return
	}
	if info, err := os.Stat(sourceFilePath); os.IsNotExist(err) {
		util.Logger.Debug("Transcode file not exist", zap.String("sourceFilePath", sourceFilePath))
		job.Status = model.JOB_ERROR
		job.Error = "Tập tin không tồn tại!"
		if err := db.Omit("App").Save(&job).Error; err != nil {
			util.Logger.Error("Transcode db.save error", zap.String("error", err.Error()))
		}
		Callback(TranscodeCallback{
			RequestID: job.RequestID,
			Status:    model.JOB_ERROR,
			Error:     "Tập tin không tồn tại!",
		}, job.App.Key, job.Callback)
		return
	} else if info.IsDir() {
		util.Logger.Debug("Transcode file is directory", zap.String("sourceFilePath", sourceFilePath))
		job.Status = model.JOB_ERROR
		job.Error = "Tập tin là thư mục!"
		job.Retry++
		if err := db.Omit("App").Save(&job).Error; err != nil {
			util.Logger.Error("Transcode db.save error", zap.String("error", err.Error()))
		}
		Callback(TranscodeCallback{
			RequestID: job.RequestID,
			Status:    model.JOB_ERROR,
			Error:     "Tập tin là thư mục!",
		}, job.App.Key, job.Callback)
		return
	}

	/* Create destination folder */
	destinationFolderPath, err := safeJoin(destinationFolder, job.Destination)
	if err != nil {
		util.Logger.Debug("Transcode invalid destination path", zap.String("destination", job.Destination), zap.String("error", err.Error()))
		job.Status = model.JOB_ERROR
		job.Error = "Đường dẫn đích không hợp lệ!"
		if err := db.Omit("App").Save(&job).Error; err != nil {
			util.Logger.Error("Transcode db.save error", zap.String("error", err.Error()))
		}
		Callback(TranscodeCallback{
			RequestID: job.RequestID,
			Status:    model.JOB_ERROR,
			Error:     "Đường dẫn đích không hợp lệ!",
		}, job.App.Key, job.Callback)
		return
	}
	_ = os.MkdirAll(destinationFolderPath, 0775)

	/* Get media info */
	var mediaInfo = util.GetMediaInfo(sourceFilePath)
	var videoWidth, videoHeight int64
	var duration float64
	for _, track := range mediaInfo.Media.Track {
		if track.Type == "Video" {
			videoWidth, _ = strconv.ParseInt(track.Width, 10, 64)
			videoHeight, _ = strconv.ParseInt(track.Height, 10, 64)
			duration, _ = strconv.ParseFloat(track.Duration, 64)
		}
	}

	/* Create cmd args (no shell involved — avoids command injection via job.Source/Destination) */
	var renditionArgs []string
	var masterPlaylist = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><smil title=\"\"><body><switch>"
	miscParams := []string{"-threads", "8", "-hide_banner", "-y", "-loglevel", "panic", "-progress", "-", "-stats_period", "5"}
	options := []string{"-map_metadata", "-1", "-c:a", "aac", "-ar", "48000", "-c:v", "libx264", "-profile:v", "main", "-preset", "slow", "-crf", "18"}

	appendRendition := func(width, height, videoBitrate, audioBitrate int64) {
		renditionArgs = append(renditionArgs, options...)
		renditionArgs = append(renditionArgs, "-vf", fmt.Sprintf("scale=%d:-2", width))
		renditionArgs = append(renditionArgs,
			"-b:v", fmt.Sprintf("%dk", videoBitrate),
			"-maxrate", fmt.Sprintf("%dk", int(float64(videoBitrate)*maxBitrateRatio)),
			"-bufsize", fmt.Sprintf("%dk", int(float64(videoBitrate)*rateMonitorBufferRatio)),
			"-b:a", fmt.Sprintf("%dk", audioBitrate),
		)
		renditionArgs = append(renditionArgs, filepath.Join(destinationFolderPath, fmt.Sprintf("%d.mp4", height)))
		masterPlaylist += fmt.Sprintf("<video height=\"%d\" src=\"%d.mp4\" width=\"%d\"><param name=\"videoBitrate\" value=\"%d\" valuetype=\"data\"></param><param name=\"audioBitrate\" value=\"%d\" valuetype=\"data\"></param></video>", height, height, width, videoBitrate, audioBitrate)
	}

	for _, re := range renditions {
		if videoWidth >= re.Width {
			appendRendition(re.Width, re.Height, re.VideoBitrate, re.AudioBitrate)
		}
	}

	if len(renditionArgs) == 0 {
		appendRendition(videoWidth, videoHeight, 1400, 96)
	}

	masterPlaylist += "</switch></body></smil>"

	ffmpegArgs := append([]string{}, miscParams...)
	ffmpegArgs = append(ffmpegArgs, "-i", sourceFilePath)
	ffmpegArgs = append(ffmpegArgs, renditionArgs...)
	util.Logger.Debug("Transcode shell", zap.String("command", "ffmpeg "+strings.Join(ffmpegArgs, " ")))

	if err := Shell(ffmpegArgs, duration, db, job); err != nil {
		util.Logger.Error("Transcode Shell error", zap.String("error", err.Error()))
		job.Status = model.JOB_ERROR
		job.Error = "Transcode error!"
		job.Retry++
		if err := db.Omit("App").Save(&job).Error; err != nil {
			util.Logger.Error("Transcode db.save error", zap.String("error", err.Error()))
		}
		Callback(TranscodeCallback{
			RequestID: job.RequestID,
			Status:    model.JOB_ERROR,
			Error:     "Transcode error!",
		}, job.App.Key, job.Callback)
		return
	}

	/* Write master playlist */
	if f, err := os.Create(filepath.Join(destinationFolderPath, fmt.Sprintf("%s.smil", job.RequestID))); err != nil {
		util.Logger.Error("Transcode os.Create error", zap.String("error", err.Error()))
		job.Status = model.JOB_ERROR
		job.Error = "Lỗi tạo master playlist!"
		job.Retry++
		if err := db.Omit("App").Save(&job).Error; err != nil {
			util.Logger.Error("Transcode db.save error", zap.String("error", err.Error()))
		}
		Callback(TranscodeCallback{
			RequestID: job.RequestID,
			Status:    model.JOB_ERROR,
			Error:     "Lỗi tạo master playlist!",
		}, job.App.Key, job.Callback)
		return
	} else {
		if _, err := f.WriteString(masterPlaylist); err != nil {
			util.Logger.Error("Transcode f.WriteString error", zap.String("error", err.Error()))
			job.Status = model.JOB_ERROR
			job.Error = "Lỗi ghi master playlist!"
			job.Retry++
			if err := db.Omit("App").Save(&job).Error; err != nil {
				util.Logger.Error("Transcode db.save error", zap.String("error", err.Error()))
			}
			Callback(TranscodeCallback{
				Error:     "Lỗi ghi master playlist!",
				RequestID: job.RequestID,
				Status:    model.JOB_ERROR,
			}, job.App.Key, job.Callback)
			return
		} else {
			_ = f.Close()
			job.Status = model.JOB_DONE
			job.Error = ""
			job.Progress = 100
			job.Retry++
			if err := db.Omit("App").Save(&job).Error; err != nil {
				util.Logger.Error("Transcode db.save error", zap.String("error", err.Error()))
			}
			Callback(TranscodeCallback{
				Error:     "",
				RequestID: job.RequestID,
				Status:    model.JOB_DONE,
			}, job.App.Key, job.Callback)
			util.Logger.Info("Transcode success", zap.String("id", job.ID), zap.String("requestID", job.RequestID))
			return
		}
	}
}

func Shell(args []string, duration float64, db *gorm.DB, job model.Job) error {
	cmd := exec.Command("ffmpeg", args...)
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		m := scanner.Text()
		if strings.HasPrefix(m, "out_time_ms=") && duration > 0 {
			outTime, _ := strconv.ParseInt(strings.ReplaceAll(m, "out_time_ms=", ""), 10, 64)
			outTime = outTime / 1000000
			job.Status = model.JOB_RUNNING
			job.Error = ""
			job.Progress = int(float64(outTime) / duration * 100)
			if err := db.Omit("App").Save(&job).Error; err != nil {
				util.Logger.Error("Shell db.save error", zap.String("error", err.Error()))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		util.Logger.Error("Shell scanner error", zap.String("error", err.Error()))
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func Callback(data TranscodeCallback, key, url string) {
	dataJson, _ := json.Marshal(data)
	checksum := util.MD5Hash(string(dataJson) + key)
	bodyJson, _ := json.Marshal(map[string]interface{}{
		"checksum": checksum,
		"data":     data,
	})

	util.Logger.Info("Callback", zap.String("data", string(dataJson)), zap.String("checksum", checksum))

	if resp, err := util.HttpRequest(url, "POST", map[string]string{"Content-type": "application/json"}, bytes.NewBuffer(bodyJson)); err != nil {
		util.Logger.Error("Callback util.HttpRequest error", zap.String("error", err.Error()))
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			util.Logger.Info("Callback success", zap.String("body", string(bodyJson)))
		} else {
			body, _ := io.ReadAll(resp.Body)
			util.Logger.Error("Callback util.HttpRequest error", zap.String("body", string(body)))
		}
	}
}
