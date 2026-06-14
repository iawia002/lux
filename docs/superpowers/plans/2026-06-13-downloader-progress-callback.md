# Downloader Progress Callback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a structured downloader progress callback that reports aggregate final-output progress to Go library callers while preserving existing CLI progress behavior.

**Architecture:** Add exported progress types and an optional `Options.Progress` callback. Introduce a small thread-safe progress tracker owned by each `Download` call, and route all save, multi-thread save, skip, error, and merge paths through tracker snapshots. Keep `pb.ProgressBar` rendering unchanged for CLI users.

**Tech Stack:** Go 1.24, `httptest`, existing `pb/v3`, existing downloader/request/utils packages.

---

## File Structure

- Modify `downloader/types.go`: add exported progress status constants, `ProgressEvent`, and unexported tracker helpers.
- Modify `downloader/downloader.go`: add `Options.Progress`, pass progress context into save paths, emit aggregate events, and preserve existing progress bar calls.
- Modify `downloader/downloader_test.go`: replace external-network downloader tests with local `httptest` coverage for progress callback paths.
- Modify `README.md`: add a compact Go library callback example.

### Task 1: Progress API Types and Single-File Test

**Files:**
- Modify: `downloader/types.go`
- Modify: `downloader/downloader.go`
- Modify: `downloader/downloader_test.go`

- [ ] **Step 1: Write the failing single-file progress test**

Add a local test to `downloader/downloader_test.go` that uses `httptest.Server`, a temp output directory, and `Options.Progress` to collect events:

```go
func TestDownloadProgressSingleFile(t *testing.T) {
	body := []byte("hello lux progress")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		_, _ = w.Write(body)
	}))
	defer server.Close()

	var events []ProgressEvent
	data := testDownloadData(server.URL, int64(len(body)), "mp4")
	err := New(Options{
		Silent:     true,
		OutputPath: t.TempDir(),
		RetryTimes: 1,
		Progress: func(event ProgressEvent) {
			events = append(events, event)
		},
	}).Download(data)
	if err != nil {
		t.Fatal(err)
	}

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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./downloader -run TestDownloadProgressSingleFile -count=1`

Expected: compile failure because `ProgressEvent`, `ProgressPreparing`, `Options.Progress`, and helpers are not defined yet.

- [ ] **Step 3: Add exported API and minimal tracker**

In `downloader/types.go`, add:

```go
type ProgressStatus string

const (
	ProgressPreparing   ProgressStatus = "preparing"
	ProgressDownloading ProgressStatus = "downloading"
	ProgressFinished    ProgressStatus = "finished"
	ProgressMerging     ProgressStatus = "merging"
	ProgressSkipped     ProgressStatus = "skipped"
	ProgressError       ProgressStatus = "error"
)

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
```

In `downloader/downloader.go`, add `Progress func(ProgressEvent)` to `Options`.

Add unexported tracker helpers in `downloader/types.go`:

```go
type progressTracker struct {
	mu         sync.Mutex
	callback   func(ProgressEvent)
	outputPath string
	total      int64
	downloaded int64
	startedAt  time.Time
}

func newProgressTracker(callback func(ProgressEvent), outputPath string, total int64) *progressTracker
func (tracker *progressTracker) emit(status ProgressStatus, url string, partIndex, partCount int)
func (tracker *progressTracker) add(delta int64, status ProgressStatus, url string, partIndex, partCount int)
func (tracker *progressTracker) setDownloaded(downloaded int64, status ProgressStatus, url string, partIndex, partCount int)
```

Each helper must copy a snapshot while holding the mutex, release the mutex, then call `callback(snapshot)`.

- [ ] **Step 4: Wire single-file save path**

Change `save` signature to accept the tracker and part context:

```go
func (downloader *Downloader) save(part *extractors.Part, refer, fileName string, tracker *progressTracker, partIndex, partCount int) error
```

Change `writeFile` to accept tracker context and wrap the destination writer with a counting writer:

```go
func (downloader *Downloader) writeFile(url string, file *os.File, headers map[string]string, tracker *progressTracker, partIndex, partCount int) (int64, error)
```

After `io.Copy`, preserve the existing error behavior. For each write, keep using `downloader.Bar.NewProxyWriter(file)` and update tracker bytes from the same write stream.

In `Download`, create a tracker after `mergedFilePath` is known:

```go
tracker := newProgressTracker(downloader.option.Progress, mergedFilePath, stream.Size)
tracker.emit(ProgressPreparing, "", 0, len(stream.Parts))
```

Call `tracker.emit(ProgressFinished, "", 0, 1)` after successful single-file download and before returning.

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./downloader -run TestDownloadProgressSingleFile -count=1`

Expected: PASS.

### Task 2: Skip and Resume Events

**Files:**
- Modify: `downloader/downloader.go`
- Modify: `downloader/downloader_test.go`

- [ ] **Step 1: Write skipped-output test**

Add a test that pre-creates the final output file in a temp directory, downloads matching test data, and asserts `skipped` and `finished`:

```go
func TestDownloadProgressSkippedExistingOutput(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "sample.mp4")
	if err := os.WriteFile(outputPath, []byte("already here"), 0644); err != nil {
		t.Fatal(err)
	}

	var events []ProgressEvent
	data := testDownloadData("http://example.invalid/video.mp4", int64(len("already here")), "mp4")
	err := New(Options{
		Silent:     true,
		OutputPath: dir,
		RetryTimes: 1,
		Progress: func(event ProgressEvent) {
			events = append(events, event)
		},
	}).Download(data)
	if err != nil {
		t.Fatal(err)
	}

	skipped := lastEventWithStatus(t, events, ProgressSkipped)
	if skipped.DownloadedBytes != skipped.TotalBytes {
		t.Fatalf("skipped event downloaded=%d total=%d, want complete", skipped.DownloadedBytes, skipped.TotalBytes)
	}
	assertHasStatus(t, events, ProgressFinished)
}
```

- [ ] **Step 2: Run skipped test to verify it fails**

Run: `go test ./downloader -run TestDownloadProgressSkippedExistingOutput -count=1`

Expected: FAIL because merged-output skip currently returns without callback events.

- [ ] **Step 3: Emit skip and resume snapshots**

In the merged-output skip branch in `Download`, emit:

```go
tracker.setDownloaded(stream.Size, ProgressSkipped, "", 0, len(stream.Parts))
tracker.emit(ProgressFinished, "", 0, len(stream.Parts))
```

In `save`, when an existing complete part is skipped, call `tracker.add(fileSize, ProgressSkipped, part.URL, partIndex, partCount)` after `downloader.Bar.Add64(fileSize)`.

In `save`, when a temporary file is resumed, call `tracker.add(tempFileSize, ProgressDownloading, part.URL, partIndex, partCount)` after `downloader.Bar.Add64(tempFileSize)`.

- [ ] **Step 4: Run skipped test to verify it passes**

Run: `go test ./downloader -run TestDownloadProgressSkippedExistingOutput -count=1`

Expected: PASS.

### Task 3: Multi-Thread and Multi-Fragment Aggregate Progress

**Files:**
- Modify: `downloader/downloader.go`
- Modify: `downloader/downloader_test.go`

- [ ] **Step 1: Write multi-thread aggregate test**

Add a test server that honors `Range` requests and run the downloader with `MultiThread: true` and `ThreadNumber: 2`:

```go
func TestDownloadProgressMultiThreadAggregatesOutput(t *testing.T) {
	body := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveRangeBody(t, w, r, body)
	}))
	defer server.Close()

	var events []ProgressEvent
	data := testDownloadData(server.URL, int64(len(body)), "mp4")
	err := New(Options{
		Silent:       true,
		OutputPath:   t.TempDir(),
		RetryTimes:   1,
		MultiThread:  true,
		ThreadNumber: 2,
		Progress: func(event ProgressEvent) {
			events = append(events, event)
		},
	}).Download(data)
	if err != nil {
		t.Fatal(err)
	}

	finished := lastEventWithStatus(t, events, ProgressFinished)
	if finished.DownloadedBytes != int64(len(body)) || finished.TotalBytes != int64(len(body)) {
		t.Fatalf("finished aggregate = %d/%d, want %d/%d", finished.DownloadedBytes, finished.TotalBytes, len(body), len(body))
	}
	if maxDownloaded(events) > int64(len(body)) {
		t.Fatalf("progress exceeded total: max=%d total=%d", maxDownloaded(events), len(body))
	}
}
```

- [ ] **Step 2: Run multi-thread test to verify it fails or exposes missing wiring**

Run: `go test ./downloader -run TestDownloadProgressMultiThreadAggregatesOutput -count=1`

Expected: FAIL until `multiThreadSave` receives tracker context and range writes update aggregate bytes.

- [ ] **Step 3: Wire multiThreadSave**

Change signature:

```go
func (downloader *Downloader) multiThreadSave(dataPart *extractors.Part, refer, fileName string, tracker *progressTracker, partIndex, partCount int) error
```

For complete final file and complete temp file branches, emit `ProgressSkipped` or `ProgressDownloading` snapshots after existing `Bar.Add64`.

When `savedSize > 0`, call `tracker.add(savedSize, ProgressDownloading, dataPart.URL, partIndex, partCount)`.

Pass the tracker context into each `writeFile` call from worker goroutines. Ensure worker error appends remain under the existing `mu`, but callback calls happen through tracker snapshots rather than while holding `mu`.

Change `mergeMultiPart` calls in this function to a receiver method or wrapper that emits `ProgressMerging` before merge and returns the original error unchanged.

- [ ] **Step 4: Run multi-thread test to verify it passes**

Run: `go test ./downloader -run TestDownloadProgressMultiThreadAggregatesOutput -count=1`

Expected: PASS.

- [ ] **Step 5: Audit multi-fragment call sites**

In the multi-fragment loop, pass the same tracker to all `save` or `multiThreadSave` calls with `partIndex=index` and `partCount=len(stream.Parts)`.

Before video merge, emit `ProgressMerging`. After a successful merge or after non-video multi-part downloads finish, emit `ProgressFinished`.

On any save or merge error, emit `ProgressError` before returning the existing error.

### Task 4: Documentation and Full Verification

**Files:**
- Modify: `README.md`
- Modify: `downloader/downloader_test.go`

- [ ] **Step 1: Add README library example**

Add a compact section near the options documentation:

```markdown
### Go library progress callback

Go callers can receive structured aggregate progress from the downloader without using the CLI progress bar:

```go
d := downloader.New(downloader.Options{
	Progress: func(event downloader.ProgressEvent) {
		fmt.Printf("%s %d/%d bytes\n", event.Status, event.DownloadedBytes, event.TotalBytes)
	},
})
err := d.Download(data)
```

`DownloadedBytes` and `TotalBytes` describe the selected output as a whole. `PartIndex` and `PartCount` are only context for the currently active part. `URL` is the raw part URL when available and may include query parameters.
```

- [ ] **Step 2: Run downloader tests**

Run: `go test ./downloader -count=1`

Expected: PASS.

- [ ] **Step 3: Run broader relevant tests**

Run: `go test ./... -count=1`

Expected: PASS if the environment allows all extractor tests. If external-network extractor tests fail, record the failing packages and keep the downloader test result as the main verification.

- [ ] **Step 4: Final diff audit**

Run: `git diff -- downloader README.md`

Expected: only progress API, downloader event wiring, tests, and README documentation changed.

---

## Self-Review

Spec coverage: API types, aggregate semantics, single-file, multi-thread, multi-fragment, resume, skip, merge, errors, CLI compatibility, tests, and README are covered.

Placeholder scan: no placeholders remain.

Type consistency: the plan consistently uses `ProgressStatus`, `ProgressEvent`, `Options.Progress`, `progressTracker`, `ProgressPreparing`, `ProgressDownloading`, `ProgressFinished`, `ProgressMerging`, `ProgressSkipped`, and `ProgressError`.
