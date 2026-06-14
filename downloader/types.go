package downloader

import (
	"sync"
	"time"
)

// ProgressStatus is the structured status of a downloader progress event.
type ProgressStatus string

const (
	// ProgressPreparing indicates the downloader selected output and is preparing work.
	ProgressPreparing ProgressStatus = "preparing"
	// ProgressDownloading indicates bytes are present or have just been downloaded.
	ProgressDownloading ProgressStatus = "downloading"
	// ProgressFinished indicates the selected output is complete.
	ProgressFinished ProgressStatus = "finished"
	// ProgressMerging indicates merge work is starting or in progress.
	ProgressMerging ProgressStatus = "merging"
	// ProgressSkipped indicates an existing complete file or part was reused.
	ProgressSkipped ProgressStatus = "skipped"
	// ProgressError indicates a download or merge path is returning an error.
	ProgressError ProgressStatus = "error"
)

// ProgressEvent describes aggregate progress for the selected output.
type ProgressEvent struct {
	URL              string
	OutputPath       string
	DownloadedBytes  int64
	TotalBytes       int64
	SpeedBytesPerSec int64
	Status           ProgressStatus
	PartIndex        int
	PartCount        int
}

type progressTracker struct {
	mu         sync.Mutex
	callback   func(ProgressEvent)
	outputPath string
	total      int64
	downloaded int64
	startedAt  time.Time
}

func newProgressTracker(callback func(ProgressEvent), outputPath string, total int64) *progressTracker {
	return &progressTracker{
		callback:   callback,
		outputPath: outputPath,
		total:      total,
		startedAt:  time.Now(),
	}
}

func (tracker *progressTracker) emit(status ProgressStatus, url string, partIndex, partCount int) {
	if tracker == nil {
		return
	}
	tracker.mu.Lock()
	event := tracker.snapshot(status, url, partIndex, partCount)
	callback := tracker.callback
	tracker.mu.Unlock()
	if callback != nil {
		callback(event)
	}
}

func (tracker *progressTracker) add(delta int64, status ProgressStatus, url string, partIndex, partCount int) {
	if tracker == nil {
		return
	}
	tracker.mu.Lock()
	tracker.downloaded += delta
	event := tracker.snapshot(status, url, partIndex, partCount)
	callback := tracker.callback
	tracker.mu.Unlock()
	if callback != nil {
		callback(event)
	}
}

func (tracker *progressTracker) setDownloaded(downloaded int64, status ProgressStatus, url string, partIndex, partCount int) {
	if tracker == nil {
		return
	}
	tracker.mu.Lock()
	tracker.downloaded = downloaded
	event := tracker.snapshot(status, url, partIndex, partCount)
	callback := tracker.callback
	tracker.mu.Unlock()
	if callback != nil {
		callback(event)
	}
}

func (tracker *progressTracker) snapshot(status ProgressStatus, url string, partIndex, partCount int) ProgressEvent {
	elapsed := time.Since(tracker.startedAt).Seconds()
	var speed int64
	if elapsed > 0 {
		speed = int64(float64(tracker.downloaded) / elapsed)
	}
	return ProgressEvent{
		URL:              url,
		OutputPath:       tracker.outputPath,
		DownloadedBytes:  tracker.downloaded,
		TotalBytes:       tracker.total,
		SpeedBytesPerSec: speed,
		Status:           status,
		PartIndex:        partIndex,
		PartCount:        partCount,
	}
}

// Aria2RPCData defines the data structure of json RPC 2.0 info for Aria2
type Aria2RPCData struct {
	// More info about RPC interface please refer to
	// https://aria2.github.io/manual/en/html/aria2c.html#rpc-interface
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	// For a simple download, only implemented `addUri`
	Method string `json:"method"`
	// secret, uris, options
	Params [3]interface{} `json:"params"`
}

// Aria2Input is options for `aria2.addUri`
// https://aria2.github.io/manual/en/html/aria2c.html#id3
type Aria2Input struct {
	// The file name of the downloaded file
	Out string `json:"out"`
	// For a simple download, only add headers
	Header []string `json:"header"`
}

// FilePartMeta defines the data structure of file meta info.
type FilePartMeta struct {
	Index float32
	Start int64
	End   int64
	Cur   int64
}
