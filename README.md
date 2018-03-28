<p align="center"><img src="static/logo.png" alt="Annie" height="100px"></p>

[![Build Status](https://travis-ci.org/iawia002/annie.svg?branch=master)](https://travis-ci.org/iawia002/annie)
[![codecov](https://codecov.io/gh/iawia002/annie/branch/master/graph/badge.svg)](https://codecov.io/gh/iawia002/annie)

👾 Annie is a fast, simple and clean video downloader built with Go.

Annie allows users to download videos and playlists from supported websites like Bilibili:

```console
$ annie -c cookies.txt https://www.bilibili.com/video/av20203945/

 Site:      哔哩哔哩 bilibili.com
 Title:     【2018拜年祭单品】相遇day by day
 Type:      video
 Stream:
     [default]  -------------------
     Quality:         高清 1080P60
     Size:            220.65 MiB (231363071 Bytes)
     # download with: annie -f default "URL"

 16.03 MiB / 220.65 MiB [==>----------------------------]   7.26% 9.65 MiB/s 19s
```


* [Install](#install)
* [Get Started](#get-started)
* [Supported Sites](#supported-sites)
* [Known issues](#known-issues)
* [About this project](#about-this-project)
* [Contributing](#contributing)
* [Similar projects](#similar-projects)
* [License](#license)


## Install

### Prerequisites

The following dependencies are required and must be installed separately.

* **[FFmpeg](https://www.ffmpeg.org)**

> **Note**: FFmpeg does not affect the download, only affects the final file merge.

To install Annie, use `go get`, download the binary file in the [Releases](https://github.com/iawia002/annie/releases) page, or compile yourself.

```bash
$ go get github.com/iawia002/annie
...
$ annie [args] URL
```


## Get Started

### Download a video

```console
$ annie -s 127.0.0.1:1080 https://youtu.be/Gnbch2osEeo

 Site:      YouTube youtube.com
 Title:     Multifandom Mashup 2017
 Type:      video
 Stream:
     [default]  -------------------
     Quality:         hd720
     Size:            57.97 MiB (60785404 Bytes)
     # download with: annie -f default "URL"

 11.93 MiB / 57.97 MiB [======>-------------------------]  20.57% 19.03 MiB/s 2s
```

> Note: wrap the URL in quotation marks if it contains special characters. (thanks @tonyxyl for pointing this out)
>
> `$ annie 'https://...'`

The `-i` option displays all available formats information without downloading.

```console
$ annie -i -s 127.0.0.1:1080 https://youtu.be/Gnbch2osEeo

 Site:      YouTube youtube.com
 Title:     Multifandom Mashup 2017
 Type:      video
 Streams:   # All available quality
     [43]  -------------------
     Quality:         medium
     Size:            31.95 MiB (33505824 Bytes)
     # download with: annie -f 43 "URL"

     [18]  -------------------
     Quality:         medium
     Size:            24.81 MiB (26011062 Bytes)
     # download with: annie -f 18 "URL"

     [36]  -------------------
     Quality:         small
     Size:            8.67 MiB (9088579 Bytes)
     # download with: annie -f 36 "URL"

     [17]  -------------------
     Quality:         small
     Size:            3.10 MiB (3248257 Bytes)
     # download with: annie -f 17 "URL"

     [default]  -------------------
     Quality:         hd720
     Size:            57.97 MiB (60785404 Bytes)
     # download with: annie -f default "URL"
```

Use `annie -f format "URL"` to download a specific format.

### Download anything else

If Annie is provided the URL of a specific resource, it will be downloaded directly:

```console
$ annie https://img9.bcyimg.com/drawer/15294/post/1799t/1f5a87801a0711e898b12b640777720f.jpg

annie doesn't support this URL right now, but it will try to download it directly

 Site:      Universal
 Title:     1f5a87801a0711e898b12b640777720f
 Type:      image/jpeg
 Stream:
     [default]  -------------------
     Size:            1.00 MiB (1051042 Bytes)
     # download with: annie -f default "URL"

 1.00 MiB / 1.00 MiB [===================================] 100.00% 1.21 MiB/s 0s
```

### Download playlist

The `-p` option downloads an entire playlist instead of a single video.

```console
$ annie -i -p https://www.bilibili.com/bangumi/play/ep198061

 Site:      哔哩哔哩 bilibili.com
 Title:     Doctor X 第四季：第一集
 Type:      video
 Streams:   # All available quality
     [default]  -------------------
     Quality:         高清 1080P
     Size:            845.66 MiB (886738354 Bytes)
     # download with: annie -f default "URL"


 Site:      哔哩哔哩 bilibili.com
 Title:     Doctor X 第四季：第二集
 Type:      video
 Streams:   # All available quality
     [default]  -------------------
     Quality:         高清 1080P
     Size:            930.71 MiB (975919195 Bytes)
     # download with: annie -f default "URL"

......
```

### Resume a download

<kbd>Ctrl</kbd>+<kbd>C</kbd> interrupts a download.

A temporary `.download` file is kept in the output directory. If `annie` is run with the same arguments, the download progress will resume from the last session.

### Cookies

Cookies can be provided to `annie` with the `-c` option if they are required for accessing the video.

**Note:**

* cookies must match the following format:

```
name=value; name2=value2; ...
```

cookies can be a string or a file.

```console
$ annie -c "name=value; name2=value2" https://www.bilibili.com/video/av20203945

# or

$ annie -c cookies.txt https://www.bilibili.com/video/av20203945
```


### Proxy
#### HTTP proxy

An HTTP proxy can be specified with the `-x` option:

```console
$ annie -x http://127.0.0.1:7777 -i https://www.youtube.com/watch?v=Gnbch2osEeo
```

#### SOCKS5 proxy

A SOCKS5 proxy can be specified with the `-s` option:

```console
$ annie -s 127.0.0.1:1080 -i https://www.youtube.com/watch?v=Gnbch2osEeo
```


### Use specified Referrer

A Referrer can be used for the request with the `-r` option:


```console
$ annie -r https://www.bilibili.com/video/av20383055/ http://cn-scnc1-dx.acgvideo.com/...

...
```

### Specify the output path and name

The `-o` option sets the path, and `-O` option sets the name of the downloaded file:

```console
$ annie -o ../ -O "hello" https://...
```

### Debug Mode

The `-d` option enables network request messages:

```console
$ annie -i -d http://www.bilibili.com/video/av20088587

URL:         http://www.bilibili.com/video/av20088587
Method:      GET
Headers:     http.Header{
    "Referer":         {"http://www.bilibili.com/video/av20088587"},
    "Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
    "Accept-Charset":  {"UTF-8,*;q=0.5"},
    "Accept-Encoding": {"gzip,deflate,sdch"},
    "Accept-Language": {"en-US,en;q=0.8"},
    "User-Agent":      {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.146 Safari/537.36"},
}
Status Code: 200

URL:         https://interface.bilibili.com/v2/playurl?appkey=84956560bc028eb7&cid=32782944&otype=json&qn=116&quality=116&type=&sign=fb2e3f261fec398652f96d358517e535
Method:      GET
Headers:     http.Header{
    "Accept-Charset":  {"UTF-8,*;q=0.5"},
    "Accept-Encoding": {"gzip,deflate,sdch"},
    "Accept-Language": {"en-US,en;q=0.8"},
    "User-Agent":      {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.146 Safari/537.36"},
    "Referer":         {"https://interface.bilibili.com/v2/playurl?appkey=84956560bc028eb7&cid=32782944&otype=json&qn=116&quality=116&type=&sign=fb2e3f261fec398652f96d358517e535"},
    "Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
}
Status Code: 200

 Site:      哔哩哔哩 bilibili.com
 Title:     燃油动力的遥控奥迪R8跑赛道
 Type:      video
 Streams:   # All available quality
     [default]  -------------------
     Quality:         高清 1080P
     Size:            64.38 MiB (67504795 Bytes)
     # download with: annie -f default "URL"
```

### All available arguments

```console
$ annie -h

Usage of annie:
  -O string
    	Specify the output file name
  -c string
    	Cookie
  -d	Debug mode
  -f string
    	Select specific format to download
  -i	Information only
  -o string
    	Specify the output path
  -p	Download playlist
  -r string
    	Use specified Referrer
  -s string
    	SOCKS5 proxy
  -v	Show version
  -x string
    	HTTP proxy
```


## Supported Sites

Site | URL | 🎬 Videos | 🌁 Images | 📚 Playlist
--- | --- | ---------| -------- | ---------
抖音 | <https://www.douyin.com> | ✓ | | |
哔哩哔哩 | <https://www.bilibili.com> | ✓ | | ✓ |
半次元 | <https://bcy.net> | | ✓ | |
pixivision | <https://www.pixivision.net> | | ✓ | |
优酷 | <https://www.youku.com> | ✓ | | |
YouTube | <https://www.youtube.com> | ✓ | | ✓ |
爱奇艺 | <https://www.iqiyi.com> | ✓ | | |
芒果TV | <https://www.mgtv.com> | ✓ | | |
Tumblr | <https://www.tumblr.com> | ✓ | ✓ | |
Vimeo | <https://vimeo.com> | ✓ | | |
Facebook | <https://facebook.com> | ✓ | | |
斗鱼视频 | <https://v.douyu.com> | ✓ | | |


## Known issues


## About this project

I am just a college student and this is one of my amateur projects(I need to finish my school stuff first). I am very happy and surprised that so many people like this project, thank you all. 🙇‍♂️


## Contributing

Annie is an open source project and built on the top of open source projects. If you are interested, welcome to contribute, let's make Annie better, together 💪

Check out the [Contributing Guide](./CONTRIBUTING.md) to get started.

Special thanks to [@Yasujizr](https://github.com/Yasujizr) who designed the amazing logo!


## Similar projects

* [youtube-dl](https://github.com/rg3/youtube-dl)
* [you-get](https://github.com/soimort/you-get)


## License

MIT

Copyright (c) 2018-present, iawia002
