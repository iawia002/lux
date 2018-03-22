# Annie

[![Build Status](https://travis-ci.org/iawia002/annie.svg?branch=master)](https://travis-ci.org/iawia002/annie)
[![codecov](https://codecov.io/gh/iawia002/annie/branch/master/graph/badge.svg)](https://codecov.io/gh/iawia002/annie)

👾 Annie is a fast, simple and clean video downloader built with Go. 

Annie helps users to download videos from supported websites, such as Youtube. With Annie, downloading videos and playlists has never been so easy!

```console
$ annie -c cookies.txt https://www.bilibili.com/video/av20203945/

   Site:    哔哩哔哩 bilibili.com
  Title:    【2018拜年祭单品】相遇day by day
   Type:    video
Quality:    高清 1080P60
   Size:    220.65 MiB (231363071 Bytes)

 2.06 MiB / 220.65 MiB [>-----------------------------]   0.93% 1.94 MiB/s 1m52s
```

## Install

### Prerequisites

The following dependencies are required and must be installed separately.

* **[FFmpeg](https://www.ffmpeg.org)**

> **Note**: FFmpeg does not affect the download, only affects the final file merge.

To install Annie, please use `go get`, download the binary file in the [Releases](https://github.com/iawia002/annie/releases) page, or compile yourself.

```bash
$ go get github.com/iawia002/annie
...
$ annie [args] URL
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

> Note: if you have special characters in your URL, we recommend putting URL in quotation marks. (thanks @tonyxyl for pointing this out)
> 
> `$ annie 'https://...'`

Annie does not support selecting specific video format to download, Annie will download the highest quality video that can be obtained.

### Download anything else

If you already got the URL of the exact resource you want, you can download it directly:

```console
$ annie https://img9.bcyimg.com/drawer/15294/post/1799t/1f5a87801a0711e898b12b640777720f.jpg

annie doesn't support this URL by now, but it will try to download it directly

 Site:    Universal
Title:    1f5a87801a0711e898b12b640777720f
 Type:    image/jpeg
 Size:    1.00 MiB (1051042 Bytes)

 1.00 MiB / 1.00 MiB [===================================] 100.00% 3.35 MiB/s 0s
```

### Download playlist

You can use the `-p` option to tell Annie to download the whole playlist rather than a single video.

```console
$ annie -i -p https://www.bilibili.com/bangumi/play/ep198061

 Site:    哔哩哔哩 bilibili.com
Title:    Doctor X 第四季：第一集
 Type:    video
 Size:    845.66 MiB (886738354 Bytes)


 Site:    哔哩哔哩 bilibili.com
Title:    Doctor X 第四季：第二集
 Type:    video
 Size:    930.71 MiB (975919195 Bytes)

...
```

### Resume a download

You may use <kbd>Ctrl</kbd>+<kbd>C</kbd> to interrupt a download.

A temporary `.download` file is kept in the output directory. Next time you run `annie` with the same arguments, the download progress will resume from the last session.

### Cookies

If you need to log in your account to access something (a private video or VIP only video), use the `-c` option to feed the browser cookies to `annie`.

**Note:**

* the formats of cookies as follow:

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
You can specify an HTTP proxy via `-x` option:

```console
$ annie -x http://127.0.0.1:7777 -i https://www.youtube.com/watch?v=Gnbch2osEeo
```

#### SOCKS5 proxy
You can also use `-s` option to specify a SOCKS5 proxy:

```console
$ annie -s 127.0.0.1:1080 -i https://www.youtube.com/watch?v=Gnbch2osEeo
```


### Use specified Referrer

You can use the `-r` option to tell Annie to use the given Referrer to request.

```console
$ annie -r https://www.bilibili.com/video/av20383055/ http://cn-scnc1-dx.acgvideo.com/...

...
```

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

### All available arguments

```console
$ annie -h

Usage of annie:
  -c string
    	Cookie
  -d	Debug mode
  -i	Information only
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


## Known issues
> net/http: request canceled (Client.Timeout exceeded while reading body)

This is a common issue. It is a network issue that can be solved by entering the download command again.


## Contributing
Annie is an open source project and welcome contributions 😉

Check out the [CONTRIBUTING](https://github.com/iawia002/annie/blob/master/CONTRIBUTING.md) guide to get started


## License

MIT
