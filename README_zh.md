<h1 align="center">Lux</h1>

<p align="center"><i>Let there be Lux!</i></p>

<div align="center">
  <a href="https://codecov.io/gh/iawia002/lux">
    <img src="https://img.shields.io/codecov/c/github/iawia002/lux.svg?style=flat-square" alt="Codecov">
  </a>
  <a href="https://github.com/iawia002/lux/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/iawia002/lux/ci.yml?style=flat-square" alt="GitHub Workflow Status">
  </a>
  <a href="https://goreportcard.com/report/github.com/iawia002/lux">
    <img src="https://goreportcard.com/badge/github.com/iawia002/lux?style=flat-square" alt="Go Report Card">
  </a>
  <a href="https://github.com/iawia002/lux/releases">
    <img src="https://img.shields.io/github/release/iawia002/lux.svg?style=flat-square" alt="GitHub release">
  </a>
  <a href="https://formulae.brew.sh/formula/lux">
    <img src="https://img.shields.io/homebrew/v/lux.svg?style=flat-square" alt="Homebrew">
  </a>
</div>

👾 Lux 是用Go构建的一个快速和简单的视频下载工具。

- [安装](#installation)
  - [前置条件](#prerequisites)
  - [通过`go install`安装](#install-via-go-install)
  - [Homebrew (仅限macOS )](#homebrew-macos-only)
  - [Arch Linux](#arch-linux)
  - [Void Linux](#void-linux)
  - [在Windows上用Scoop安装](#scoop-on-windows)
  - [在Windows上用Chocolatey安装](#chocolatey-on-windows)
  - [在Windows/macOS/Linux中用Cask安装](#cask-on-windowsmacoslinux)
- [入门](#getting-started)
  - [下载单个视频](#download-a-video)
  - [下载其他东西](#download-anything-else)
  - [下载播放列表](#download-playlist)
  - [多个输入](#multiple-inputs)
  - [恢复下载](#resume-a-download)
  - [自动重试](#auto-retry)
  - [Cookies](#cookies)
  - [代理](#proxy)
  - [多线程](#multi-thread)
  - [短链接](#short-link)
    - [bilibili](#bilibili)
  - [使用指定的Referrer](#use-specified-referrer)
  - [指定输出路径和名称](#specify-the-output-path-and-name)
  - [调试模式](#debug-mode)
  - [重用提取的数据](#reuse-extracted-data)
  - [选项](#options)
    - [下载:](#download)
    - [网络:](#network)
    - [播放列表:](#playlist)
    - [文件系统:](#filesystem)
    - [副标题:](#subtitle)
    - [Youku:](#youku)
    - [aria2:](#aria2)
- [支持网站](#supported-sites)
- [已知的问题](#known-issues)
  - [优酷](#优酷)
- [贡献](#contributing)
- [作者](#authors)
- [类似的项目](#similar-projects)
- [许可](#license)

## 安装

### 前置条件

以下依赖项是必需的，必须单独安装。

- **[FFmpeg](https://www.ffmpeg.org)**

> **注**:FFmpeg不影响下载，只影响最终文件合并。

### 通过`go install`安装

要安装Lux，使用“go install”，或者从[Releases](https://github.com/iawia002/lux/releases)页面下载二进制文件。

```bash
$ go install github.com/iawia002/lux@latest
```

### Homebrew (仅限macOS)

对于 macOS 用户，您可以通过以下方式安装：`lux`

```bash
$ brew install lux
```

### Arch Linux

对于 Arch 用户，可以使用 [AUR](https://aur.archlinux.org/packages/lux-dl/) 包。

### Void Linux

对于 Void Linux 用户，您可以通过以下方式安装：`lux`

```
$ xbps-install -S lux
```

### 在Windows上用[Scoop](https://scoop.sh/)安装

```sh
$ scoop install lux
```

### 在Windows上用[Chocolatey](https://chocolatey.org/)安装

```
$ choco install lux
```

### 在Windows/macOS/Linux中用[Cask](https://github.com/axetroy/cask.rs)安装

```sh
$ cask install github.com/iawia002/lux
```

## 入门

用法:

```
lux [OPTIONS] URL [URL...]
```

### 下载视频

```console
$ lux "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

 Site:      YouTube youtube.com
 Title:     Rick Astley - Never Gonna Give You Up (Video)
 Type:      video
 Stream:
     [248]  -------------------
     Quality:         1080p video/webm; codecs="vp9"
     Size:            63.93 MiB (67038963 Bytes)
     # download with: lux -f 248 ...

 41.88 MiB / 63.93 MiB [=================>-------------]  65.51% 4.22 MiB/s 00m05s
```

该选项显示所有可用的视频质量，无需下载。`-i`

```console
$ lux -i "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

 Site:      YouTube youtube.com
 Title:     Rick Astley - Never Gonna Give You Up (Video)
 Type:      video
 Streams:   # All available quality
     [248]  -------------------
     Quality:         1080p video/webm; codecs="vp9"
     Size:            49.29 MiB (51687554 Bytes)
     # download with: lux -f 248 ...

     [137]  -------------------
     Quality:         1080p video/mp4; codecs="avc1.640028"
     Size:            43.45 MiB (45564306 Bytes)
     # download with: lux -f 137 ...

     [398]  -------------------
     Quality:         720p video/mp4; codecs="av01.0.05M.08"
     Size:            37.12 MiB (38926432 Bytes)
     # download with: lux -f 398 ...

     [136]  -------------------
     Quality:         720p video/mp4; codecs="avc1.4d401f"
     Size:            31.34 MiB (32867324 Bytes)
     # download with: lux -f 136 ...

     [247]  -------------------
     Quality:         720p video/webm; codecs="vp9"
     Size:            31.03 MiB (32536181 Bytes)
     # download with: lux -f 247 ...
```

用于下载选项输出中列出的特定流。`lux -f stream "URL"``-i`

### 下载其他任何内容

如果向 Lux 提供了特定资源的 URL，则直接下载该资源：

```console
$ lux "https://img9.bcyimg.com/drawer/15294/post/1799t/1f5a87801a0711e898b12b640777720f.jpg"

lux doesn't support this URL right now, but it will try to download it directly

 Site:      Universal
 Title:     1f5a87801a0711e898b12b640777720f
 Type:      image/jpeg
 Stream:
     [default]  -------------------
     Size:            1.00 MiB (1051042 Bytes)
     # download with: lux -f default "URL"

 1.00 MiB / 1.00 MiB [===================================] 100.00% 1.21 MiB/s 0s
```

### 下载播放列表

该选项下载整个播放列表而不是单个视频。`-p`

```console
$ lux -i -p "https://www.bilibili.com/bangumi/play/ep198061"

 Site:      哔哩哔哩 bilibili.com
 Title:     Doctor X 第四季：第一集
 Type:      video
 Streams:   # All available quality
     [default]  -------------------
     Quality:         高清 1080P
     Size:            845.66 MiB (886738354 Bytes)
     # download with: lux -f default "URL"


 Site:      哔哩哔哩 bilibili.com
 Title:     Doctor X 第四季：第二集
 Type:      video
 Streams:   # All available quality
     [default]  -------------------
     Quality:         高清 1080P
     Size:            930.71 MiB (975919195 Bytes)
     # download with: lux -f default "URL"

......
```

可以使用 或 选项指定列表的下载范围：`-start``-end``-items`

```
-start
    	Playlist video to start at (default 1)
-end
    	Playlist video to end at
-items
    	Playlist video items to download. Separated by commas like: 1,5,6,8-10
```

仅适用于哔哩哔哩播放列表：

```
-eto
  File name of each bilibili episode doesn't include the playlist title
```

### 多个输入

您还可以一次下载多个网址：

```console
$ lux -i "https://www.bilibili.com/video/av21877586" "https://www.bilibili.com/video/av21990740"

 Site:      哔哩哔哩 bilibili.com
 Title:     【莓机会了】甜到虐哭的13集单集MAD「我现在什么都不想干,更不想看14集」
 Type:      video
 Streams:   # All available quality
     [default]  -------------------
     Quality:         高清 1080P
     Size:            51.88 MiB (54403767 Bytes)
     # download with: lux -f default "URL"


 Site:      哔哩哔哩 bilibili.com
 Title:     【莓救了】甜到虐哭！！！国家队单集MAD-当熟悉的bgm响起，眼泪从脸颊滑下
 Type:      video
 Streams:   # All available quality
     [default]  -------------------
     Quality:         高清 1080P
     Size:            77.63 MiB (81404093 Bytes)
     # download with: lux -f default "URL"
```

这些 URL 将逐个下载。

您还可以使用该选项从文件中读取 URL：`-F`

```console
$ lux -F ~/Desktop/u.txt

 Site:      微博 weibo.com
 Title:     在Google，我们设计什么？ via@阑夕
 Type:      video
 Stream:
     [default]  -------------------
     Size:            19.19 MiB (20118196 Bytes)
     # download with: lux -f default "URL"

 19.19 MiB / 19.19 MiB [=================================] 100.00% 9.69 MiB/s 1s

......
```

可以使用 或 选项指定列表的下载范围：`-start``-end``-items`

```
-start
    	File line to start at (default 1)
-end
    	File line to end at
-items
    	File lines to download. Separated by commas like: 1,5,6,8-10
```

### 恢复下载

<kbd>Ctrl</kbd>+<kbd>C</kbd> 中断下载。

临时文件保存在输出目录中。如果使用相同的参数运行，则下载进度将从上一个会话恢复。`.download``lux`

### 自动重试

下载失败时，lux会自动重试，您可以通过选项指定重试次数（默认值为100）。`-retry`

### Cookies

如果访问视频需要 Cookie，则可以提供带有选项的 Cookie。`lux``-c`

Cookie 可以是以下格式或 [Netscape Cookie](https://curl.haxx.se/rfc/cookie_spec.html) 格式：

```console
name=value; name2=value2; ...
```

Cookie 可以是字符串或文本文件，通过以下两种方式之一提供 Cookie。

作为字符串：

```console
$ lux -c "name=value; name2=value2" "https://www.bilibili.com/video/av20203945"
```

作为文本文件：

```console
$ lux -c cookies.txt "https://www.bilibili.com/video/av20203945"
```

### 代理

您可以使用环境变量设置 HTTP/SOCKS5 代理：

```console
$ HTTP_PROXY="http://127.0.0.1:1087/" lux -i "https://www.youtube.com/watch?v=Gnbch2osEeo"
```

```console
$ HTTP_PROXY="socks5://127.0.0.1:1080/" lux -i "https://www.youtube.com/watch?v=Gnbch2osEeo"
```

### 多线程

使用或多个线程下载单个视频。`--multi-thread``-m`

使用 or 选项设置下载线程数（默认值为 10）。`--thread``-n`

> 注意：如果视频有多个片段，则实际下载线程的数量会增加。
>
> 例如：
>
> - 如果设置为 10，并且视频有 2 个片段，则实际将使用 20 个线程。`-n`
> - 如果视频有 20 个片段，同时只下载 10 个片段，实际线程数为 100。

### 短链接

#### 哔哩哔哩

您只需使用或编号即可下载哔哩哔哩的视频：`av``ep`

```console
$ lux -i ep198381 av21877586

 Site:      哔哩哔哩 bilibili.com
 Title:     狐妖小红娘：第79话 南国公主的吃货本色
 Type:      video
 Streams:   # All available quality
     [default]  -------------------
     Quality:         高清 1080P
     Size:            485.23 MiB (508798478 Bytes)
     # download with: lux -f default "URL"


 Site:      哔哩哔哩 bilibili.com
 Title:     【莓机会了】甜到虐哭的13集单集MAD「我现在什么都不想干,更不想看14集」
 Type:      video
 Streams:   # All available quality
     [default]  -------------------
     Quality:         高清 1080P
     Size:            51.88 MiB (54403767 Bytes)
     # download with: lux -f default "URL"
```

### 使用指定的引荐来源网址

反向链接可用于请求，选项为：`-r`

```console
$ lux -r "https://www.bilibili.com/video/av20383055/" "http://cn-scnc1-dx.acgvideo.com/"
```

### 指定输出路径和名称

该选项设置路径，选项设置下载文件的名称：`-o``-O`

```console
$ lux -o ../ -O "hello" "https://example.com"
```

### 调试模式

该选项输出网络请求消息：`-d`

```console
$ lux -i -d "http://www.bilibili.com/video/av20088587"

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
     # download with: lux -f default "URL"
```

### 重用提取的数据

该选项将以 JSON 格式打印提取的数据。`-j`

```console
$ lux -j "https://www.bilibili.com/video/av20203945"

{
    "site": "哔哩哔哩 bilibili.com",
    "title": "【2018拜年祭单品】相遇day by day",
    "type": "video",
    "streams": {
        "15": {
            "urls": [
                {
                    "url": "...",
                    "size": 18355205,
                    "ext": "flv"
                }
            ],
            "quality": "流畅 360P",
            "size": 18355205
        },
        "32": {
            "urls": [
                {
                    "url": "...",
                    "size": 40058632,
                    "ext": "flv"
                }
            ],
            "quality": "清晰 480P",
            "size": 40058632
        },
        "64": {
            "urls": [
                {
                    "url": "...",
                    "size": 82691087,
                    "ext": "flv"
                }
            ],
            "quality": "高清 720P",
            "size": 82691087
        },
        "80": {
            "urls": [
                {
                    "url": "...",
                    "size": 121735559,
                    "ext": "flv"
                }
            ],
            "quality": "高清 1080P",
            "size": 121735559
        }
    }
}
```

### 选项

```
  -i	Information only
  -F string
    	URLs file path
  -d	Debug mode
  -j	Print extracted data
  -s	Minimum outputs
  -v	Show version
```

#### 下载:

```
  -f string
    	Select specific stream to download
  -p	Download playlist
  -n int
    	The number of download thread (only works for multiple-parts video) (default 10)
  -c string
    	Cookie
  -r string
    	Use specified Referrer
  -cs int
    	HTTP chunk size for downloading (in MB) (default 1)
```

#### 网络:

```
  -retry int
    	How many times to retry when the download failed (default 10)
```

#### 播放列表:

```
  -start int
    	Playlist video to start at (default 1)
  -end int
    	Playlist video to end at
  -items string
    	Playlist video items to download. Separated by commas like: 1,5,6,8-10
```

#### 文件系统:

```
  -o string
    	Specify the output path
  -O string
    	Specify the output file name
```

#### 字幕:

```
  -C	Download captions
```

#### 优酷:

```
  -ccode string
    	Youku ccode (default "0502")
  -ckey string
    	Youku ckey (default "7B19C0AB12633B22E7FE81271162026020570708D6CC189E4924503C49D243A0DE6CD84A766832C2C99898FC5ED31F3709BB3CDD82C96492E721BDD381735026")
  -password string
    	Youku password
```

#### aria2:

> 注意：如果您使用 aria2 下载，则需要自己合并多部分视频。

```
  -aria2
    	Use Aria2 RPC to download
  -aria2addr string
    	Aria2 Address (default "localhost:6800")
  -aria2method string
    	Aria2 Method (default "http")
  -aria2token string
    	Aria2 RPC Token
```

## 支持的网站

| 网站             | 网址                                                         | 🎬 视频 | 🌁 图像 | 🔊 音频 | 📚 播放列表 | 🍪 VIP 权限 | 构建状态                                                     |
| ---------------- | ------------------------------------------------------------ | ------ | ------ | ------ | ---------- | ---------- | ------------------------------------------------------------ |
| 抖音             | <https://www.douyin.com>                                     | ✓      | ✓      |        |            |            | [![douyin](https://github.com/iawia002/lux/actions/workflows/stream_douyin.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_douyin.yml) |
| 哔哩哔哩         | <https://www.bilibili.com>                                   | ✓      |        |        | ✓          | ✓          | [![bilibili](https://github.com/iawia002/lux/actions/workflows/stream_bilibili.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_bilibili.yml) |
| 半次元           | <https://bcy.net>                                            |        | ✓      |        |            |            | [![bcy](https://github.com/iawia002/lux/actions/workflows/stream_bcy.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_bcy.yml) |
| pixivision       | <https://www.pixivision.net>                                 |        | ✓      |        |            |            | [![pixivision](https://github.com/iawia002/lux/actions/workflows/stream_pixivision.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_pixivision.yml) |
| 优酷             | <https://www.youku.com>                                      | ✓      |        |        |            | ✓          | [![youku](https://github.com/iawia002/lux/actions/workflows/stream_youku.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_youku.yml) |
| YouTube          | <https://www.youtube.com>                                    | ✓      |        |        | ✓          |            | [![youtube](https://github.com/iawia002/lux/actions/workflows/stream_youtube.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_youtube.yml) |
| 西瓜视频（头条） | <https://m.toutiao.com>, <https://v.ixigua.com>, <https://www.ixigua.com> | ✓      |        |        |            |            | [![ixigua](https://github.com/iawia002/lux/actions/workflows/stream_ixigua.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_ixigua.yml) |
| 爱奇艺           | <https://www.iqiyi.com>                                      | ✓      |        |        |            |            | [![iqiyi](https://github.com/iawia002/lux/actions/workflows/stream_iqiyi.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_iqiyi.yml) |
| 新片场           | <https://www.xinpianchang.com>                               | ✓      |        |        |            |            | [![xinpianchang](https://github.com/iawia002/lux/actions/workflows/stream_xinpianchang.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_xinpianchang.yml) |
| 芒果 TV          | <https://www.mgtv.com>                                       | ✓      |        |        |            |            | [![mgtv](https://github.com/iawia002/lux/actions/workflows/stream_mgtv.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_mgtv.yml) |
| 糖豆广场舞       | <https://www.tangdou.com>                                    | ✓      |        |        |            |            | [![tangdou](https://github.com/iawia002/lux/actions/workflows/stream_tangdou.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_tangdou.yml) |
| Tumblr           | <https://www.tumblr.com>                                     | ✓      | ✓      |        |            |            | [![tumblr](https://github.com/iawia002/lux/actions/workflows/stream_tumblr.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_tumblr.yml) |
| Vimeo            | <https://vimeo.com>                                          | ✓      |        |        |            |            | [![vimeo](https://github.com/iawia002/lux/actions/workflows/stream_vimeo.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_vimeo.yml) |
| Facebook         | <https://facebook.com>                                       | ✓      |        |        |            |            | [![facebook](https://github.com/iawia002/lux/actions/workflows/stream_facebook.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_facebook.yml) |
| 斗鱼视频         | <https://v.douyu.com>                                        | ✓      |        |        |            |            | [![douyu](https://github.com/iawia002/lux/actions/workflows/stream_douyu.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_douyu.yml) |
| 秒拍             | <https://www.miaopai.com>                                    | ✓      |        |        |            |            | [![miaopai](https://github.com/iawia002/lux/actions/workflows/stream_miaopai.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_miaopai.yml) |
| 微博             | <https://weibo.com>                                          | ✓      |        |        |            |            | [![weibo](https://github.com/iawia002/lux/actions/workflows/stream_weibo.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_weibo.yml) |
| Instagram        | <https://www.instagram.com>                                  | ✓      | ✓      |        |            |            | [![instagram](https://github.com/iawia002/lux/actions/workflows/stream_instagram.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_instagram.yml) |
| Twitter          | <https://twitter.com>                                        | ✓      |        |        |            |            | [![twitter](https://github.com/iawia002/lux/actions/workflows/stream_twitter.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_twitter.yml) |
| 腾讯视频         | <https://v.qq.com>                                           | ✓      |        |        |            |            | [![qq](https://github.com/iawia002/lux/actions/workflows/stream_qq.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_qq.yml) |
| 网易云音乐       | <https://music.163.com>                                      | ✓      |        |        |            |            | [![netease](https://github.com/iawia002/lux/actions/workflows/stream_netease.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_netease.yml) |
| 音悦台           | <https://yinyuetai.com>                                      | ✓      |        |        |            |            | [![yinyuetai](https://github.com/iawia002/lux/actions/workflows/stream_yinyuetai.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_yinyuetai.yml) |
| 极客时间         | <https://time.geekbang.org>                                  | ✓      |        |        |            |            | [![geekbang](https://github.com/iawia002/lux/actions/workflows/stream_geekbang.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_geekbang.yml) |
| Pornhub          | <https://pornhub.com>                                        | ✓      |        |        |            |            | [![pornhub](https://github.com/iawia002/lux/actions/workflows/stream_pornhub.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_pornhub.yml) |
| XVIDEOS          | <https://xvideos.com>                                        | ✓      |        |        |            |            | [![xvideos](https://github.com/iawia002/lux/actions/workflows/stream_xvideos.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_xvideos.yml) |
| 聯合新聞網       | <https://udn.com>                                            | ✓      |        |        |            |            | [![udn](https://github.com/iawia002/lux/actions/workflows/stream_udn.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_udn.yml) |
| TikTok           | <https://www.tiktok.com>                                     | ✓      |        |        |            |            | [![tiktok](https://github.com/iawia002/lux/actions/workflows/stream_tiktok.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_tiktok.yml) |
| Pinterest        | <https://www.pinterest.com>                                  | ✓      |        |        |            |            | [![pinterest](https://github.com/iawia002/lux/actions/workflows/stream_pinterest.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_pinterest.yml) |
| 好看视频         | <https://haokan.baidu.com>                                   | ✓      |        |        |            |            | [![haokan](https://github.com/iawia002/lux/actions/workflows/stream_haokan.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_haokan.yml) |
| AcFun            | <https://www.acfun.cn>                                       | ✓      |        |        | ✓          |            | [![acfun](https://github.com/iawia002/lux/actions/workflows/stream_acfun.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_acfun.yml) |
| Eporner          | <https://eporner.com>                                        | ✓      |        |        |            |            | [![eporner](https://github.com/iawia002/lux/actions/workflows/stream_eporner.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_eporner.yml) |
| StreamTape       | <https://streamtape.com>                                     | ✓      |        |        |            |            | [![streamtape](https://github.com/iawia002/lux/actions/workflows/stream_streamtape.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_streamtape.yml) |
| 虎扑             | <https://hupu.com>                                           | ✓      |        |        |            |            | [![hupu](https://github.com/iawia002/lux/actions/workflows/stream_hupu.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_hupu.yml) |
| 虎牙视频         | <https://v.huya.com>                                         | ✓      |        |        |            |            | [![huya](https://github.com/iawia002/lux/actions/workflows/stream_huya.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_huya.yml) |
| 喜马拉雅         | <https://www.ximalaya.com>                                   |        |        | ✓      |            |            | [![ximalaya](https://github.com/iawia002/lux/actions/workflows/stream_ximalaya.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_ximalaya.yml) |
| 快手             | <https://www.kuaishou.com>                                   | ✓      |        |        |            |            | [![kuaishou](https://github.com/iawia002/lux/actions/workflows/stream_kuaishou.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_kuaishou.yml) |
| Reddit           | <https://www.reddit.com>                                     | ✓      | ✓      |        |            |            | [![reddit](https://github.com/iawia002/lux/actions/workflows/stream_reddit.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_reddit.yml) |
| VKontakte        | <https://vk.com>                                             | ✓      |        |        |            |            | [![vk](https://github.com/iawia002/lux/actions/workflows/stream_vk.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_vk.yml/) |
| 知乎             | <https://zhihu.com>                                          | ✓      |        |        |            |            | [![zhihu](https://github.com/iawia002/lux/actions/workflows/stream_zhihu.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_zhihu.yml/) |
| Rumble           | <https://rumble.com>                                         | ✓      |        |        |            |            | [![rumble](https://github.com/iawia002/lux/actions/workflows/stream_rumble.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_rumble.yml/) |
| 小红书           | <https://xiaohongshu.com>                                    | ✓      |        |        |            |            | [![xiaohongshu](https://github.com/iawia002/lux/actions/workflows/stream_xiaohongshu.yml/badge.svg)](https://github.com/iawia002/lux/actions/workflows/stream_xiaohongshu.yml/) |


## 已知的问题

### 优酷

优酷的 `ccode` 经常变化导致 lux 不可用，如果你知道有新的可用的 `ccode`，可以直接使用 `lux -ccode ...` 而不用等待 lux 更新（当然，也欢迎你给我们提一个 Pull request 来更新默认的 `ccode`）

最好是每次下载都附带登录过的 Cookie 以避免部分 `ccode` 的问题

## 贡献

Lux 是一个开源项目，建立在开源项目之上。 点击 [这里](./CONTRIBUTING.md) 查看贡献指南。

感谢 [JetBrains](https://www.jetbrains.com/?from=lux) 提供的精彩的 IDE.

<a href="https://www.jetbrains.com/?from=lux"><img src="static/jetbrains-variant-3.svg" /></a>

## 作者

Code with ❤️ by [iawia002](https://github.com/iawia002) and lovely [contributors](https://github.com/iawia002/lux/graphs/contributors)

## 类似的项目

- [youtube-dl](https://github.com/rg3/youtube-dl)
- [you-get](https://github.com/soimort/you-get)
- [ytdl](https://github.com/rylio/ytdl)

## 许可

MIT

Copyright (c) 2018-present, iawia002
