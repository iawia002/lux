# Downloader Progress Callback Design

## Goal

Add a stable structured progress callback to the `downloader` package so Go library callers can observe whole-output download progress without invoking the CLI. Existing CLI progress bar output and default downloader behavior must remain unchanged when the callback is unset.

## Public API

Extend `downloader.Options` with:

```go
Progress func(ProgressEvent)
```

Add exported progress types in the `downloader` package:

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

`URL` is the raw `extractors.Part.URL` for the currently active part when available. It may include query parameters from the extractor, so callers should avoid logging it if that is sensitive. The field may be empty for merged-output or file-only events.

## Event Semantics

`ProgressEvent` represents progress for the selected stream's final output, not an internal `.part` file. `DownloadedBytes` and `TotalBytes` are aggregate values for the whole output whenever the stream size is known.

`PartIndex` and `PartCount` are auxiliary context only. They identify the part currently being handled but do not change the meaning of `DownloadedBytes`.

Status meanings:

- `preparing`: the downloader has selected an output path and is about to download or resume.
- `downloading`: bytes were already present or have just been written.
- `skipped`: a complete output or part already exists.
- `merging`: merge work is starting or in progress.
- `finished`: the selected output is complete.
- `error`: a download or merge path is returning an error.

## Download Paths

Single-file downloads will emit `preparing`, one or more `downloading` events, and `finished`. Resumed temporary files will emit a `downloading` event reflecting the already-present bytes before network writes continue.

Multi-thread range downloads will aggregate bytes from all worker goroutines through a small thread-safe progress tracker. Callback calls will use snapshots and will not hold downloader locks while executing user code.

Multi-fragment downloads will share the same stream-level tracker across fragments. Each fragment updates the aggregate `DownloadedBytes`, while `PartIndex` and `PartCount` indicate the fragment context.

Existing complete files will emit `skipped` and `finished` with `DownloadedBytes == TotalBytes`.

Merge operations will emit `merging` before invoking the existing merge helpers and `finished` after a successful merge. Errors preserve the original error chain and additionally emit `error`.

## CLI Compatibility

The CLI will not set `Options.Progress`. The existing `pb.ProgressBar`, `bar.Add64(...)`, `bar.NewProxyWriter(...)`, and printed messages remain responsible for CLI rendering.

The callback API will be implemented alongside the current progress bar, not by replacing it.

## Testing

Add downloader unit tests that avoid external network access:

- Use `httptest.Server` for a normal single-file download and assert `downloading` and `finished` events with aggregate byte counts.
- Create an existing complete output file and assert `skipped` and `finished`.
- Cover a key multi-thread or multi-fragment path and assert the callback reports aggregate progress instead of only per-part byte counts.

Run targeted downloader tests first. Run broader Go tests as time allows and report exactly what was executed.

## Documentation

Update README or package documentation with a compact Go library example showing `downloader.Options{Progress: func(e downloader.ProgressEvent) { ... }}`.
