# Annie

👾 A simple and clean video downloader built with Go

```console
 Site:    抖音 douyin.com
Title:    好冷  逢考必过
 Type:    mp4
 Size:    2.76 MiB (2762719 Bytes)

 2.63 MiB / 2.63 MiB [===================================] 100.00% 1.03 MiB/s 2s
```

## Install

To install Annie, please use `go get`, or compile yourself.

```bash
$ go get github.com/iawia002/annie
...
$ annie ...
```


## Get Started

### Download a video

```console
$ annie https://www.douyin.com/share/video/6509219899754155272

 Site:    抖音 douyin.com
Title:    好冷  逢考必过
 Type:    mp4
 Size:    2.63 MiB (2762719 Bytes)

 741.70 KiB / 2.63 MiB [=========>--------------------------]  27.49% 1.98 MiB/s
```

### Resume a download

You may use <kbd>Ctrl</kbd>+<kbd>C</kbd> to interrupt a download.

A temporary `.download` file is kept in the output directory. Next time you run `annie` with the same arguments, the download progress will resume from the last session.


## Supported Sites

Site | URL | Videos
--- | --- | ------
抖音 | <https://www.douyin.com> | ✓
哔哩哔哩 | <https://www.bilibili.com> | ✓


## Build

Make sure that this folder is in `GOPATH`, then:

```bash
$ go build
```
