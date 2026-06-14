package downloader

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/OldBeLight/lux/extractors"
	"github.com/OldBeLight/lux/request"
)

func TestDownloadProgressSingleFile(t *testing.T) {
	body := []byte("hello lux progress")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		_, _ = w.Write(body)
	}))
	defer server.Close()

	recorder := newProgressRecorder()
	data := testDownloadData(server.URL, int64(len(body)), "mp4")
	err := New(Options{
		Silent:     true,
		OutputPath: t.TempDir(),
		RetryTimes: 1,
		Progress:   recorder.Record,
	}).Download(data)
	if err != nil {
		t.Fatal(err)
	}

	events := recorder.Events()
	assertHasStatus(t, events, ProgressPreparing)
	assertHasStatus(t, events, ProgressDownloading)
	finished := lastEventWithStatus(t, events, ProgressFinished)
	if finished.DownloadedBytes != int64(len(body)) {
		t.Fatalf("finished downloaded bytes = %d, want %d", finished.DownloadedBytes, len(body))
	}
	if finished.TotalBytes != int64(len(body)) {
		t.Fatalf("finished total bytes = %d, want %d", finished.TotalBytes, len(body))
	}
	if finished.PartIndex != 0 || finished.PartCount != 1 {
		t.Fatalf("finished part context = %d/%d, want 0/1", finished.PartIndex, finished.PartCount)
	}
}

func TestDownloadUsesRequestContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request reached server after context was canceled")
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := New(Options{
		Silent:     true,
		OutputPath: t.TempDir(),
		RetryTimes: 1,
		Request: request.Options{
			RetryTimes: 1,
			Context:    ctx,
		},
	}).Download(testDownloadData(server.URL, 1, "mp4"))
	if err == nil {
		t.Fatal("Download() error = nil, want canceled context error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Download() error = %v, want context.Canceled", err)
	}
}

func TestDownloadUsesRequestOptions(t *testing.T) {
	body := []byte("hello library request options")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "downloader-agent" {
			t.Fatalf("User-Agent = %q, want downloader-agent", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		_, _ = w.Write(body)
	}))
	defer server.Close()

	err := New(Options{
		Silent:     true,
		OutputPath: t.TempDir(),
		RetryTimes: 1,
		Request: request.Options{
			RetryTimes: 1,
			UserAgent:  "downloader-agent",
		},
	}).Download(testDownloadData(server.URL, int64(len(body)), "mp4"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSortedStreamsForLibrary(t *testing.T) {
	streams := map[string]*extractors.Stream{
		"small": {ID: "small", Size: 1},
		"large": {ID: "large", Size: 10},
	}

	sorted := SortedStreamsForLibrary(streams)
	if len(sorted) != 2 || sorted[0].ID != "large" || sorted[1].ID != "small" {
		t.Fatalf("sorted streams = %#v", sorted)
	}
}

func TestDownloadProgressSkippedExistingOutput(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "sample.mp4")
	if err := os.WriteFile(outputPath, []byte("already here"), 0644); err != nil {
		t.Fatal(err)
	}

	recorder := newProgressRecorder()
	data := testDownloadData("http://example.invalid/video.mp4", int64(len("already here")), "mp4")
	err := New(Options{
		Silent:     true,
		OutputPath: dir,
		RetryTimes: 1,
		Progress:   recorder.Record,
	}).Download(data)
	if err != nil {
		t.Fatal(err)
	}

	events := recorder.Events()
	skipped := lastEventWithStatus(t, events, ProgressSkipped)
	if skipped.DownloadedBytes != skipped.TotalBytes {
		t.Fatalf("skipped event downloaded=%d total=%d, want complete", skipped.DownloadedBytes, skipped.TotalBytes)
	}
	assertHasStatus(t, events, ProgressFinished)
}

func TestDownloadProgressMultiThreadAggregatesOutput(t *testing.T) {
	body := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveRangeBody(t, w, r, body)
	}))
	defer server.Close()

	recorder := newProgressRecorder()
	data := testDownloadData(server.URL, int64(len(body)), "mp4")
	err := New(Options{
		Silent:       true,
		OutputPath:   t.TempDir(),
		RetryTimes:   1,
		MultiThread:  true,
		ThreadNumber: 2,
		Progress:     recorder.Record,
	}).Download(data)
	if err != nil {
		t.Fatal(err)
	}

	events := recorder.Events()
	finished := lastEventWithStatus(t, events, ProgressFinished)
	if finished.DownloadedBytes != int64(len(body)) || finished.TotalBytes != int64(len(body)) {
		t.Fatalf("finished aggregate = %d/%d, want %d/%d", finished.DownloadedBytes, finished.TotalBytes, len(body), len(body))
	}
	if maxDownloaded(events) > int64(len(body)) {
		t.Fatalf("progress exceeded total: max=%d total=%d", maxDownloaded(events), len(body))
	}
}

func TestDownloadProgressMultiFragmentAggregatesOutput(t *testing.T) {
	bodies := [][]byte{
		[]byte("first image bytes"),
		[]byte("second image bytes"),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		index, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/part"))
		if err != nil || index < 0 || index >= len(bodies) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(bodies[index])))
		_, _ = w.Write(bodies[index])
	}))
	defer server.Close()

	total := int64(len(bodies[0]) + len(bodies[1]))
	data := &extractors.Data{
		Site:  "test",
		Title: "sample",
		Type:  extractors.DataTypeImage,
		URL:   server.URL,
		Streams: map[string]*extractors.Stream{
			"default": {
				ID: "default",
				Parts: []*extractors.Part{
					{URL: server.URL + "/part0", Size: int64(len(bodies[0])), Ext: "jpg"},
					{URL: server.URL + "/part1", Size: int64(len(bodies[1])), Ext: "jpg"},
				},
				Size: total,
				Ext:  "jpg",
			},
		},
	}

	recorder := newProgressRecorder()
	err := New(Options{
		Silent:       true,
		OutputPath:   t.TempDir(),
		RetryTimes:   1,
		ThreadNumber: 2,
		Progress:     recorder.Record,
	}).Download(data)
	if err != nil {
		t.Fatal(err)
	}

	events := recorder.Events()
	finished := lastEventWithStatus(t, events, ProgressFinished)
	if finished.DownloadedBytes != total || finished.TotalBytes != total {
		t.Fatalf("finished aggregate = %d/%d, want %d/%d", finished.DownloadedBytes, finished.TotalBytes, total, total)
	}
	assertHasPartContext(t, events, 0, 2)
	assertHasPartContext(t, events, 1, 2)
}

func testDownloadData(url string, size int64, ext string) *extractors.Data {
	data := &extractors.Data{
		Site:  "test",
		Title: "sample",
		Type:  extractors.DataTypeVideo,
		URL:   url,
		Streams: map[string]*extractors.Stream{
			"default": {
				ID: "default",
				Parts: []*extractors.Part{
					{
						URL:  url,
						Size: size,
						Ext:  ext,
					},
				},
				Size: size,
				Ext:  ext,
			},
		},
	}
	return data
}

type progressRecorder struct {
	mu     sync.Mutex
	events []ProgressEvent
}

func newProgressRecorder() *progressRecorder {
	return &progressRecorder{}
}

func (recorder *progressRecorder) Record(event ProgressEvent) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	recorder.events = append(recorder.events, event)
}

func (recorder *progressRecorder) Events() []ProgressEvent {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return append([]ProgressEvent(nil), recorder.events...)
}

func serveRangeBody(t *testing.T, w http.ResponseWriter, r *http.Request, body []byte) {
	t.Helper()

	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		_, _ = w.Write(body)
		return
	}

	if !strings.HasPrefix(rangeHeader, "bytes=") {
		t.Fatalf("unexpected range header %q", rangeHeader)
	}
	bounds := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
	if len(bounds) != 2 {
		t.Fatalf("unexpected range header %q", rangeHeader)
	}

	start, err := strconv.Atoi(bounds[0])
	if err != nil {
		t.Fatal(err)
	}
	end := len(body) - 1
	if bounds[1] != "" {
		end, err = strconv.Atoi(bounds[1])
		if err != nil {
			t.Fatal(err)
		}
	}
	if start < 0 || end >= len(body) || start > end {
		t.Fatalf("invalid range %q for body length %d", rangeHeader, len(body))
	}

	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(body)))
	w.Header().Set("Content-Length", strconv.Itoa(end-start+1))
	w.WriteHeader(http.StatusPartialContent)
	_, _ = w.Write(body[start : end+1])
}

func assertHasPartContext(t *testing.T, events []ProgressEvent, partIndex, partCount int) {
	t.Helper()
	for _, event := range events {
		if event.PartIndex == partIndex && event.PartCount == partCount {
			return
		}
	}
	t.Fatalf("missing part context %d/%d in events %#v", partIndex, partCount, events)
}

func assertHasStatus(t *testing.T, events []ProgressEvent, status ProgressStatus) {
	t.Helper()
	for _, event := range events {
		if event.Status == status {
			return
		}
	}
	t.Fatalf("missing progress status %q in events %#v", status, events)
}

func lastEventWithStatus(t *testing.T, events []ProgressEvent, status ProgressStatus) ProgressEvent {
	t.Helper()
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Status == status {
			return events[i]
		}
	}
	t.Fatalf("missing progress status %q in events %#v", status, events)
	return ProgressEvent{}
}

func maxDownloaded(events []ProgressEvent) int64 {
	var max int64
	for _, event := range events {
		if event.DownloadedBytes > max {
			max = event.DownloadedBytes
		}
	}
	return max
}
