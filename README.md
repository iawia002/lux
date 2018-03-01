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

### Prerequisites

The following dependencies are required and must be installed separately.

* **[FFmpeg](https://www.ffmpeg.org)**

> **Note**: FFmpeg does not affect the download, only affect the final file merge.

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

You can also use the `-i` option to view video information, skip download.

### Resume a download

You may use <kbd>Ctrl</kbd>+<kbd>C</kbd> to interrupt a download.

A temporary `.download` file is kept in the output directory. Next time you run `annie` with the same arguments, the download progress will resume from the last session.

### Debug Mode

You can use the `-d` option to see network request message.

```console
$ annie -i -d http://www.bilibili.com/video/av20088587

URL: http://www.bilibili.com/video/av20088587
Method: GET
Headers: map[User-Agent:[Mozilla/5.0 (Windows NT 10.0; WOW64; rv:51.0) Gecko/20100101 Firefox/51.0] Accept:[text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8] Accept-Charset:[UTF-8,*;q=0.5] Accept-Encoding:[gzip,deflate,sdch] Accept-Language:[en-US,en;q=0.8] Referer:[http://www.bilibili.com/video/av20088587]]
Status Code: 200

URL: https://interface.bilibili.com/v2/playurl?appkey=84956560bc028eb7&cid=32782944&otype=json&quality=0&type=&sign=708701ffaea9937d4541d5cc2f1cf3b1
Method: GET
Headers: map[Accept:[text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8] Accept-Charset:[UTF-8,*;q=0.5] Accept-Encoding:[gzip,deflate,sdch] Accept-Language:[en-US,en;q=0.8] User-Agent:[Mozilla/5.0 (Windows NT 10.0; WOW64; rv:51.0) Gecko/20100101 Firefox/51.0] Referer:[https://interface.bilibili.com/v2/playurl?appkey=84956560bc028eb7&cid=32782944&otype=json&quality=0&type=&sign=708701ffaea9937d4541d5cc2f1cf3b1]]
Status Code: 200

 Site:    哔哩哔哩 bilibili.com
Title:    燃油动力的遥控奥迪R8跑赛道
 Type:    flv
 Size:    64.38 MiB (67504795 Bytes)
```


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
